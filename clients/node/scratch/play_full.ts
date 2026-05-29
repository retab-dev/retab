// Self-contained end-to-end exercise of the Node SDK against a local server.
// Builds a workflow FROM SCRATCH (declarative spec.apply), publishes it, runs it
// on a real PDF, inspects steps + clean JSON in handle outputs, then creates and
// runs an experiment and a test.
//
// Run:  bun run scratch/play_full.ts
import { Retab, RetabError } from '../src/index.js';
import type { WorkflowRun, WorkflowRunStep } from '../src/index.js';

const API_KEY = 'sk_retab_khCajUA8OAo0sFVw89g4CoDW0n9Rv6PBxFdY7OL7XV4';
const BASE_URL = 'http://localhost:4000';
const PDF_PATH = '/Users/sachaichbiah/Local/retab/handmade-tests/booking_confirmation.pdf';

const client = new Retab({ apiKey: API_KEY, baseUrl: BASE_URL });

let pass = 0;
let fail = 0;
function ok(label: string, cond: boolean, detail?: string) {
  if (cond) {
    pass++;
    console.log(`  ✅ ${label}${detail ? ` — ${detail}` : ''}`);
  } else {
    fail++;
    console.log(`  ❌ ${label}${detail ? ` — ${detail}` : ''}`);
  }
}
function section(t: string) {
  console.log(`\n=== ${t} ===`);
}
function errDetail(e: unknown): string {
  if (e instanceof RetabError) return `${e.status} ${e.responseBody}`;
  return (e as Error)?.stack ?? String(e);
}
const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));
const lc = (lifecycle: unknown): string => (lifecycle as { status?: string })?.status ?? '?';

function isCleanJson(v: unknown): boolean {
  // Clean = JSON-serializable and NOT a double-encoded JSON string blob.
  try {
    JSON.stringify(v);
  } catch {
    return false;
  }
  if (typeof v === 'string') {
    const s = v.trim();
    if ((s.startsWith('{') && s.endsWith('}')) || (s.startsWith('[') && s.endsWith(']'))) {
      return false;
    }
  }
  return true;
}

async function pollTerminal<T>(
  fetchOne: () => Promise<T>,
  statusOf: (v: T) => string,
  terminal = new Set(['completed', 'error', 'cancelled', 'awaiting_review']),
  timeoutMs = 180_000
): Promise<T> {
  const start = Date.now();
  let last = await fetchOne();
  while (Date.now() - start < timeoutMs) {
    last = await fetchOne();
    if (terminal.has(statusOf(last))) return last;
    await sleep(2500);
  }
  return last;
}

// A small, fast extraction workflow: start_document -> extract.
function buildYaml(wfId: string): string {
  return `apiVersion: workflows.retab.com/v1alpha2
kind: Workflow
metadata:
  id: ${wfId}
  name: SDK Play E2E
  description: Created from scratch by the Node SDK exercise script
spec:
  blocks:
    start:
      type: start_document
      label: Document
    extract:
      type: extract
      label: Extract Booking
      config:
        model: retab-small
        image_resolution_dpi: 150
        n_consensus: 1
        inputs:
          - name: document
            type: file
        json_schema:
          type: object
          properties:
            summary:
              type: string
              description: One sentence summary of what this document is.
            total_amount:
              type: number
              description: The total amount stated on the document, 0 if none.
          required:
            - summary
            - total_amount
  edges:
    - from:
        block: start
        handle: output-file-0
      to:
        block: extract
        handle: input-file-document
`;
}

async function main() {
  const wfId = `wrk_sdkplay_${Date.now()}`;
  let workflowId = wfId;

  section('1. Validate + apply workflow spec (create from scratch)');
  const yaml = buildYaml(wfId);
  const validation = await client.workflows.spec.validate(yaml);
  console.log(`  validation: ${JSON.stringify(validation).slice(0, 200)}`);
  const applied = await client.workflows.spec.apply(yaml);
  workflowId = applied.workflowId;
  console.log(
    `  applied: workflowId=${applied.workflowId} created=${applied.created} blocks=${applied.blockCount} edges=${applied.edgeCount} action=${applied.action}`
  );
  ok('spec applied / workflow created', !!applied.workflowId && applied.blockCount === 2);
  ok('edge created', applied.edgeCount === 1);

  section('2. Inspect workflow, blocks, edges');
  const wf = await client.workflows.get(workflowId);
  ok('workflow fetched', wf.id === workflowId, wf.name);

  const blocks = (await client.workflows.blocks.list({ workflowId })).data;
  console.log(`  blocks: ${blocks.map((b) => `${b.label}[${b.type}]`).join(', ')}`);
  ok('two blocks present', blocks.length === 2);
  const startBlock = blocks.find((b) => b.type === 'start_document');
  const extractBlock = blocks.find((b) => b.type === 'extract');
  ok('start_document block present', !!startBlock);
  ok('extract block present', !!extractBlock);

  const edges = (await client.workflows.edges.list({ workflowId })).data;
  console.log(`  edges: ${edges.map((e) => `${e.sourceBlock}->${e.targetBlock}`).join(', ')}`);
  ok('edge wired start->extract', edges.length === 1);

  section('3. Publish workflow');
  const published = await client.workflows.publish(workflowId, 'initial publish from SDK script');
  ok(
    'workflow published',
    !!published.published,
    `version=${JSON.stringify(published.published).slice(0, 80)}`
  );

  section('4. Create run with a real PDF document');
  const created = await client.workflows.runs.create(workflowId, {
    [startBlock!.id]: PDF_PATH,
  });
  console.log(`  run id: ${created.id} status=${lc(created.lifecycle)}`);
  ok('run created', !!created.id);
  ok('inputs echoed', !!created.inputs?.documents?.[startBlock!.id]);

  section('5. Poll run to terminal state');
  const run = (await pollTerminal(
    () => client.workflows.runs.get(created.id),
    (r) => lc((r as WorkflowRun).lifecycle)
  )) as WorkflowRun;
  const status = lc(run.lifecycle);
  console.log(`  terminal status: ${status}  timing=${JSON.stringify(run.timing)}`);
  ok(
    'run reached terminal/awaiting',
    ['completed', 'awaiting_review', 'error'].includes(status),
    status
  );

  section('6. List run steps + verify clean JSON in handleOutputs');
  const stepsPage = await client.workflows.steps.list({ runId: created.id, limit: 50 });
  const steps: WorkflowRunStep[] = stepsPage.data;
  ok('steps returned', steps.length > 0, `${steps.length} steps`);
  console.log(`  steps: ${steps.map((s) => `${s.blockLabel}[${lc(s.lifecycle)}]`).join(', ')}`);

  const completed = steps.filter((s) => lc(s.lifecycle) === 'completed');
  let jsonHandleCount = 0;
  let cleanCount = 0;
  let testTarget: { stepId: string; blockId: string; outputHandleId: string } | null = null;

  for (const s of completed) {
    const full = await client.workflows.steps.get(s.stepId, { runId: created.id });
    const outKeys = Object.keys(full.handleOutputs ?? {});
    console.log(`  • ${full.blockLabel} [${full.blockType}] outHandles=[${outKeys.join(', ')}]`);
    for (const [hk, payload] of Object.entries(full.handleOutputs ?? {})) {
      if (payload.type === 'file') {
        console.log(`      ↳ ${hk} (file) doc=${payload.document?.id}`);
        continue;
      }
      jsonHandleCount++;
      const clean = isCleanJson(payload.data);
      if (clean) cleanCount++;
      console.log(
        `      ↳ ${hk} (${payload.type}) clean=${clean} data=${JSON.stringify(payload.data)?.slice(0, 240)}`
      );
      ok(`handleOutput ${full.blockLabel}.${hk} is clean JSON`, clean);
      if (!testTarget && full.blockType === 'extract') {
        testTarget = { stepId: full.stepId, blockId: full.blockId, outputHandleId: hk };
      }
    }
  }
  console.log(`  summary: ${jsonHandleCount} json handle outputs, ${cleanCount} clean`);
  ok('found clean JSON in a handleOutput', cleanCount > 0);

  section('7. Experiments: create + run + results + metrics');
  try {
    const expBlockId = testTarget?.blockId ?? extractBlock!.id;
    const exp = await client.workflows.experiments.create(
      workflowId,
      expBlockId,
      [{ runId: created.id }],
      null,
      3,
      `sdk-play-exp-${Date.now()}`
    );
    console.log(`  experiment id: ${exp.id}`);
    ok('experiment created', !!exp.id);
    ok('experiment get', (await client.workflows.experiments.get(exp.id)).id === exp.id);
    const expList = await client.workflows.experiments.list({ workflowId, limit: 5 });
    ok(
      'experiment list includes it',
      expList.data.some((e) => e.id === exp.id)
    );

    const expRun = await client.workflows.experiments.runs.create(exp.id, workflowId);
    console.log(`  experiment run id: ${expRun.id} status=${lc(expRun.lifecycle)}`);
    ok('experiment run created', !!expRun.id);

    const expTerm = await pollTerminal(
      () => client.workflows.experiments.runs.get(expRun.id),
      (r) => lc(r.lifecycle),
      new Set(['completed', 'error', 'cancelled'])
    );
    console.log(
      `  experiment run terminal: ${lc(expTerm.lifecycle)} score=${expTerm.score ?? 'n/a'} docs=${expTerm.completedDocumentCount}/${expTerm.totalDocumentCount} errors=${expTerm.errorCount}`
    );
    ok(
      'experiment run reached terminal',
      ['completed', 'error'].includes(lc(expTerm.lifecycle)),
      lc(expTerm.lifecycle)
    );

    const expResults = await client.workflows.experiments.results.list({
      runId: expRun.id,
      limit: 10,
    });
    console.log(`  experiment results: ${expResults.data.length}`);
    ok('experiment results fetched', Array.isArray(expResults.data));
    try {
      const metrics = await client.workflows.experiments.metrics.get({
        runId: expRun.id,
        view: 'summary',
      });
      console.log(`  experiment metrics: ${JSON.stringify(metrics).slice(0, 200)}`);
      ok('experiment metrics fetched', !!metrics);
    } catch (e) {
      console.log(`  ⚠️  metrics: ${errDetail(e)}`);
    }
  } catch (e) {
    fail++;
    console.log(`  ❌ experiment flow threw: ${errDetail(e)}`);
  }

  section('8. Tests: create + run + results');
  try {
    if (!testTarget) throw new Error('no extract step with a JSON output handle to target');
    console.log(`  targeting block=${testTarget.blockId} handle=${testTarget.outputHandleId}`);
    const test = await client.workflows.tests.create(
      workflowId,
      { type: 'block', blockId: testTarget.blockId },
      { type: 'run_step', runId: created.id, stepId: testTarget.stepId },
      {
        target: { outputHandleId: testTarget.outputHandleId },
        condition: { kind: 'exists' },
        label: 'extract output exists',
      },
      `sdk-play-test-${Date.now()}`
    );
    console.log(`  test id: ${test.id}`);
    ok('test created', !!test.id);
    ok('test get', (await client.workflows.tests.get(test.id)).id === test.id);
    const testList = await client.workflows.tests.list({ workflowId, limit: 5 });
    ok(
      'test list includes it',
      testList.data.some((t) => t.id === test.id)
    );

    const testRun = await client.workflows.tests.runs.create(workflowId, {
      type: 'single',
      testId: test.id,
    });
    console.log(`  test run id: ${testRun.id} status=${lc(testRun.lifecycle)}`);
    ok('test run created', !!testRun.id);

    const testTerm = await pollTerminal(
      () => client.workflows.tests.runs.get(testRun.id),
      (r) => lc(r.lifecycle),
      new Set(['completed', 'error', 'cancelled'])
    );
    console.log(
      `  test run terminal: ${lc(testTerm.lifecycle)} total=${testTerm.totalTests} outcome=${JSON.stringify(testTerm.counts?.outcome)}`
    );
    ok(
      'test run reached terminal',
      ['completed', 'error'].includes(lc(testTerm.lifecycle)),
      lc(testTerm.lifecycle)
    );

    const testResults = await client.workflows.tests.results.list({ runId: testRun.id, limit: 10 });
    console.log(`  test results: ${testResults.data.length}`);
    for (const r of testResults.data) console.log(`    - ${JSON.stringify(r).slice(0, 240)}`);
    ok('test results fetched', Array.isArray(testResults.data));
  } catch (e) {
    fail++;
    console.log(`  ❌ test flow threw: ${errDetail(e)}`);
  }

  section('RESULTS');
  console.log(`  workflow: ${workflowId}`);
  console.log(`  PASS=${pass} FAIL=${fail}`);
  process.exit(fail > 0 ? 1 : 0);
}

main().catch((e) => {
  console.error('FATAL', errDetail(e));
  process.exit(2);
});
