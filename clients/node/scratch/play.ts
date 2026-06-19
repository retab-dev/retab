// Scratch script: exercise the Node SDK against local server.
// Run with: bun run scratch/play.ts
import { Retab, RetabError } from '../src/index.js';
import type { WorkflowRun, WorkflowRunStep } from '../src/index.js';

function errDetail(e: unknown): string {
  if (e instanceof RetabError) return `${e.status} ${e.responseBody}`;
  return (e as Error).message;
}

const API_KEY = 'sk_retab_khCajUA8OAo0sFVw89g4CoDW0n9Rv6PBxFdY7OL7XV4';
const BASE_URL = 'http://localhost:4000';
const WORKFLOW_ID = 'workflow_ClQgW5YKOxUOjDyz2Tt_3';
const START_BLOCK = 'block_cKT5b6HgBfjv-XAlXeHJ1';
// Existing uploaded file on this workflow (deed.tiff)
const EXISTING_FILE = {
  id: 'file_LPjuee2tTZgfM_Km5yh_G',
  filename: 'deed.tiff',
  mime_type: 'image/tiff',
};

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
const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

function isCleanJson(v: unknown): boolean {
  // Clean = JSON-serializable round-trips and is not a string-encoded JSON blob.
  try {
    JSON.stringify(v);
  } catch {
    return false;
  }
  if (typeof v === 'string') {
    const s = v.trim();
    if ((s.startsWith('{') && s.endsWith('}')) || (s.startsWith('[') && s.endsWith(']'))) {
      return false; // double-encoded JSON string — not clean
    }
  }
  return true;
}

function lc(lifecycle: unknown): string {
  return (lifecycle as { status?: string })?.status ?? '?';
}

async function pollTerminal<T>(
  fetchOne: () => Promise<T>,
  statusOf: (v: T) => string,
  timeoutMs = 120_000
): Promise<T> {
  const start = Date.now();
  const terminal = new Set(['completed', 'error', 'cancelled']);
  let last = await fetchOne();
  while (Date.now() - start < timeoutMs) {
    last = await fetchOne();
    if (terminal.has(statusOf(last))) return last;
    await sleep(2000);
  }
  return last;
}

async function waitForRun(runId: string, timeoutMs = 120_000): Promise<WorkflowRun> {
  const start = Date.now();
  let last: WorkflowRun | null = null;
  const terminal = new Set(['completed', 'error', 'cancelled', 'awaiting_review']);
  while (Date.now() - start < timeoutMs) {
    last = await client.workflows.runs.get(runId);
    const status = (last.lifecycle as { status: string }).status;
    if (terminal.has(status)) return last;
    await sleep(2000);
  }
  if (!last) throw new Error('run never fetched');
  return last;
}

async function main() {
  section('1. Workflow get');
  const wf = await client.workflows.get(WORKFLOW_ID);
  ok('workflow fetched', wf.id === WORKFLOW_ID, `${wf.name}`);
  ok('has published version', !!wf.published);

  section('2. Create run');
  const created = await client.workflows.runs.create(WORKFLOW_ID, {
    [START_BLOCK]: EXISTING_FILE,
  });
  console.log(
    `  run id: ${created.id}, initial status: ${(created.lifecycle as { status: string }).status}`
  );
  ok('run created', !!created.id);
  ok('inputs echoed', !!created.inputs?.documents?.[START_BLOCK]);

  section('3. Poll run to terminal state');
  const run = await waitForRun(created.id);
  const status = (run.lifecycle as { status: string }).status;
  console.log(`  terminal status: ${status}`);
  ok(
    'run reached terminal/awaiting',
    ['completed', 'awaiting_review', 'error'].includes(status),
    status
  );
  console.log(`  timing: ${JSON.stringify(run.timing)}`);

  section('4. List run steps + verify clean JSON in handleOutputs');
  const stepsPage = await client.workflows.steps.list({ runId: created.id, limit: 50 });
  const steps: WorkflowRunStep[] = stepsPage.data;
  ok('steps returned', steps.length > 0, `${steps.length} steps`);

  // NOTE: the list endpoint strips handle payloads for bandwidth (server-side
  // projection). handleOutputs are only populated on the single-step GET.
  const listHasPayloads = steps.some((s) => Object.keys(s.handleOutputs ?? {}).length > 0);
  console.log(
    `  list endpoint exposes handle payloads: ${listHasPayloads} (expected false — stripped for bandwidth)`
  );

  const completed = steps.filter((s) => (s.lifecycle as { status: string }).status === 'completed');
  console.log(`  completed steps: ${completed.map((s) => s.blockLabel).join(', ')}`);

  let jsonHandleCount = 0;
  let cleanCount = 0;
  let testTarget: { stepId: string; blockId: string; outputHandleId: string } | null = null;

  for (const s of completed) {
    // Fetch the FULL step to get handle payloads.
    const full = await client.workflows.steps.get(s.stepId, { runId: created.id });
    const outKeys = Object.keys(full.handleOutputs ?? {});
    console.log(
      `  • ${full.blockLabel} [${full.blockType}] — outHandles: [${outKeys.join(', ')}] artifact=${full.artifact?.id ?? 'none'}`
    );
    for (const [hk, payload] of Object.entries(full.handleOutputs ?? {})) {
      if (payload.type === 'file') {
        console.log(`      ↳ ${hk} (file) doc=${payload.document?.id}`);
        continue;
      }
      jsonHandleCount++;
      const clean = isCleanJson(payload.data);
      if (clean) cleanCount++;
      const preview = JSON.stringify(payload.data)?.slice(0, 200);
      console.log(`      ↳ ${hk} (type=${payload.type}) clean=${clean} data=${preview}`);
      ok(`handleOutput ${full.blockLabel}.${hk} is clean JSON`, clean);
      if (!testTarget && full.blockType === 'extract') {
        testTarget = { stepId: full.stepId, blockId: full.blockId, outputHandleId: hk };
      }
    }
  }
  console.log(
    `  summary: ${jsonHandleCount} json handle outputs (completed steps), ${cleanCount} clean`
  );
  ok('found clean JSON in at least one handleOutput', cleanCount > 0);

  section('5. Experiments: create + run');
  let experimentRunId: string | null = null;
  try {
    // block_id is required (unless cloning from source_experiment_id).
    const expBlockId = testTarget?.blockId ?? 'block_NOUoDOXYboTy8McT2ZugO';
    const exp = await client.workflows.experiments.create(
      WORKFLOW_ID,
      expBlockId,
      // capture documents from the run we just made
      [{ runId: created.id }],
      null,
      3, // n_consensus must be 3, 5, or 7
      `sdk-play-exp-${Date.now()}`
    );
    console.log(`  experiment id: ${exp.id}`);
    ok('experiment created', !!exp.id);

    const fetched = await client.workflows.experiments.get(exp.id);
    ok('experiment get', fetched.id === exp.id);

    const expList = await client.workflows.experiments.list({ workflowId: WORKFLOW_ID, limit: 5 });
    ok(
      'experiment list',
      expList.data.some((e) => e.id === exp.id)
    );

    const expRun = await client.workflows.experiments.runs.create(exp.id, WORKFLOW_ID);
    experimentRunId = expRun.id;
    console.log(`  experiment run id: ${expRun.id} status=${lc(expRun.lifecycle)}`);
    ok('experiment run created', !!expRun.id);

    const expTerm = await pollTerminal(
      () => client.workflows.experiments.runs.get(experimentRunId!),
      (r) => lc(r.lifecycle)
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
      runId: experimentRunId!,
      limit: 10,
    });
    console.log(`  experiment results: ${expResults.data.length}`);
    ok('experiment results fetched', Array.isArray(expResults.data));
    try {
      const metrics = await client.workflows.experiments.metrics.get({
        runId: experimentRunId!,
        view: 'summary',
      });
      console.log(`  experiment metrics: ${JSON.stringify(metrics).slice(0, 160)}`);
      ok('experiment metrics fetched', !!metrics);
    } catch (e) {
      console.log(`  ⚠️  metrics: ${errDetail(e)}`);
    }
  } catch (e) {
    fail++;
    console.log(`  ❌ experiment flow threw: ${errDetail(e)}`);
  }

  section('6. Tests: create + run');
  try {
    // start_document is not testable; target a completed extract block whose
    // real output handle we discovered above.
    if (!testTarget) throw new Error('no extract step with a JSON output handle to target');
    console.log(`  targeting ${testTarget.blockId} handle=${testTarget.outputHandleId}`);
    const test = await client.workflows.evals.create(
      WORKFLOW_ID,
      { type: 'block', blockId: testTarget.blockId },
      { type: 'run_step', runId: created.id, stepId: testTarget.stepId },
      {
        target: { outputHandleId: testTarget.outputHandleId },
        condition: { kind: 'exists' },
        label: 'output exists',
      },
      `sdk-play-test-${Date.now()}`
    );
    console.log(`  test id: ${test.id}`);
    ok('eval created', !!test.id);

    const gotEval = await client.workflows.evals.get(test.id);
    ok('eval get', gotEval.id === test.id);

    const evalList = await client.workflows.evals.list({ workflowId: WORKFLOW_ID, limit: 5 });
    ok(
      'eval list',
      evalList.data.some((t) => t.id === test.id)
    );

    const evalRun = await client.workflows.evals.runs.create(WORKFLOW_ID, {
      type: 'single',
      testId: test.id,
    });
    console.log(`  eval run id: ${evalRun.id} status=${lc(evalRun.lifecycle)}`);
    ok('eval run created', !!evalRun.id);

    const testTerm = await pollTerminal(
      () => client.workflows.evals.runs.get(evalRun.id),
      (r) => lc(r.lifecycle)
    );
    console.log(
      `  eval run terminal: ${lc(testTerm.lifecycle)} total=${testTerm.totalTests} outcome=${JSON.stringify(testTerm.counts?.outcome)}`
    );
    ok(
      'eval run reached terminal',
      ['completed', 'error'].includes(lc(testTerm.lifecycle)),
      lc(testTerm.lifecycle)
    );

    const evalResults = await client.workflows.evals.results.list({ runId: evalRun.id, limit: 10 });
    console.log(`  eval results: ${evalResults.data.length}`);
    for (const r of evalResults.data) {
      console.log(`    - ${JSON.stringify(r).slice(0, 200)}`);
    }
    ok('eval results fetched', Array.isArray(evalResults.data));
  } catch (e) {
    fail++;
    console.log(`  ❌ test flow threw: ${errDetail(e)}`);
  }

  section('RESULTS');
  console.log(`  PASS=${pass} FAIL=${fail}`);
  process.exit(fail > 0 ? 1 : 0);
}

main().catch((e) => {
  console.error('FATAL', e);
  process.exit(2);
});
