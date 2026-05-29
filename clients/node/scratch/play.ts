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

  let jsonHandleCount = 0;
  let cleanCount = 0;
  for (const s of steps) {
    const st = (s.lifecycle as { status: string }).status;
    const outKeys = Object.keys(s.handleOutputs ?? {});
    console.log(`  • ${s.blockLabel} [${s.blockType}] ${st} — outHandles: [${outKeys.join(', ')}]`);
    for (const [hk, payload] of Object.entries(s.handleOutputs ?? {})) {
      if (payload.type === 'json' || payload.data !== undefined) {
        jsonHandleCount++;
        const clean = isCleanJson(payload.data);
        if (clean) cleanCount++;
        const preview = JSON.stringify(payload.data)?.slice(0, 160);
        console.log(`      ↳ ${hk} (type=${payload.type}) clean=${clean} data=${preview}`);
        ok(`handleOutput ${s.blockLabel}.${hk} is clean JSON`, clean);
      } else if (payload.type === 'file') {
        console.log(`      ↳ ${hk} (file) doc=${payload.document?.id}`);
      }
    }
  }
  console.log(`  summary: ${jsonHandleCount} json handle outputs, ${cleanCount} clean`);

  // Also fetch a single step directly to test steps.get
  if (steps.length > 0) {
    const one = await client.workflows.steps.get(steps[0].stepId, { runId: created.id });
    ok('steps.get works', one.stepId === steps[0].stepId);
  }

  section('5. Experiments: create + run');
  let experimentRunId: string | null = null;
  try {
    const exp = await client.workflows.experiments.create(
      WORKFLOW_ID,
      null,
      // capture documents from the run we just made
      [{ workflowRunId: created.id }],
      null,
      1,
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
    console.log(
      `  experiment run id: ${expRun.id} status=${(expRun as { status?: string }).status ?? '?'}`
    );
    ok('experiment run created', !!expRun.id);
  } catch (e) {
    fail++;
    console.log(`  ❌ experiment flow threw: ${errDetail(e)}`);
  }

  section('6. Tests: create + run');
  try {
    // target the start_document block (always present + completed)
    const test = await client.workflows.tests.create(
      WORKFLOW_ID,
      { type: 'block', blockId: START_BLOCK },
      { type: 'run_step', runId: created.id },
      {
        target: { outputHandleId: START_BLOCK },
        condition: { kind: 'exists' },
        label: 'output exists',
      },
      `sdk-play-test-${Date.now()}`
    );
    console.log(`  test id: ${test.id}`);
    ok('test created', !!test.id);

    const gotTest = await client.workflows.tests.get(test.id);
    ok('test get', gotTest.id === test.id);

    const testList = await client.workflows.tests.list({ workflowId: WORKFLOW_ID, limit: 5 });
    ok(
      'test list',
      testList.data.some((t) => t.id === test.id)
    );

    const testRun = await client.workflows.tests.runs.create(WORKFLOW_ID, {
      type: 'single',
      testId: test.id,
    });
    console.log(
      `  test run id: ${testRun.id} status=${(testRun as { status?: string }).status ?? '?'}`
    );
    ok('test run created', !!testRun.id);
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
