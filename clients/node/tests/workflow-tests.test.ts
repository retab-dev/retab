import { describe, expect, test } from 'bun:test';

import { AbstractClient } from '../src/client';
import APIWorkflows from '../src/api/workflows/client';
import APIWorkflowTests from '../src/api/workflows/tests/client';
import APIWorkflowTestRuns from '../src/api/workflows/tests/runs/client';

class MockClient extends AbstractClient {
  public lastFetchParams: Record<string, unknown> | null = null;
  private readonly responseBody: unknown;

  constructor(responseBody: unknown) {
    super();
    this.responseBody = responseBody;
  }

  protected async _fetch(params: {
    url: string;
    method: string;
    params?: Record<string, unknown>;
    headers?: Record<string, unknown>;
  }): Promise<Response> {
    this.lastFetchParams = params;
    return new Response(JSON.stringify(this.responseBody), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    });
  }
}

const NOW = '2026-05-01T14:30:00Z';
const WORKFLOW_REF = {
  workflow_id: 'wf_abc123',
  version_id: 'draft_1',
  name_at_run_time: 'Q1 workflow',
  requested_version: 'draft',
};
const TRIGGER = { type: 'api' };
const PENDING = { status: 'pending' };
const COMPLETED = { status: 'completed' };
const TIMING = { created_at: NOW, started_at: NOW, completed_at: NOW };

const TEST_RESPONSE = {
  id: 'wfnodetest_abc',
  workflow_id: 'wf_abc123',
  target: { type: 'block', block_id: 'block_extract' },
  source: { type: 'manual', handle_inputs: {} },
  name: 'Q1 invoice total',
  assertion: {
    id: 'assert_xyz',
    target: { output_handle_id: 'output-json-0', path: 'total' },
    condition: { kind: 'equals', expected: 1234.56 },
    label: null,
  },
  assertion_drift_status: 'valid',
  schema_drift: 'none',
  validation_status: 'valid',
  validation_issues: [],
  latest_run_summary: null,
  latest_passing_run_summary: null,
  latest_failing_run_summary: null,
  created_at: NOW,
  updated_at: NOW,
};

const RUN_RESPONSE = {
  id: 'wftestrun_q1z2',
  workflow: WORKFLOW_REF,
  trigger: TRIGGER,
  lifecycle: COMPLETED,
  timing: TIMING,
  target: { type: 'block', block_id: 'block_extract' },
  test_id: null,
  total_tests: 1,
  counts: { passed: 1 },
};

const RESULT_RESPONSE = {
  id: 'wfresult_abc',
  run_id: 'wftestrun_q1z2',
  test_id: 'wfnodetest_abc',
  lifecycle: COMPLETED,
  timing: TIMING,
  target: { type: 'block', block_id: 'block_extract' },
  source: { type: 'manual', handle_inputs: {} },
  execution_fingerprint: 'f1',
  outputs: { 'output-json-0': { total: 1234.56 } },
  warnings: [],
  skipped: false,
  verdict: { status: 'passed' },
};

describe('workflows.tests wiring', () => {
  test('APIWorkflows exposes a `tests` sub-resource', () => {
    // Without this wire, the docs example
    // `client.workflows.tests.create(...)` raises a TypeError at
    // call time. Pin the wire so a future refactor can't drop it
    // silently.
    const workflows = new APIWorkflows(new MockClient({}));
    expect(workflows.tests).toBeInstanceOf(APIWorkflowTests);
  });

  test('APIWorkflowTests exposes a `runs` sub-resource', () => {
    const tests = new APIWorkflowTests(new MockClient({}));
    expect(tests.runs).toBeInstanceOf(APIWorkflowTestRuns);
  });

  test('APIWorkflowTests does not expose wait helpers', () => {
    const tests = new APIWorkflowTests(new MockClient({}));

    expect('waitForCompletion' in tests).toBe(false);
    expect('wait_for_completion' in tests).toBe(false);
  });
});

describe('workflows.tests.create()', () => {
  test('POSTs to the tests route with target/source/assertion/name in the body', async () => {
    const mockClient = new MockClient(TEST_RESPONSE);
    const tests = new APIWorkflowTests(mockClient);

    const test = await tests.create({
      workflowId: 'wf_abc123',
      target: { type: 'block', block_id: 'block_extract' },
      source: { type: 'manual', handle_inputs: {} },
      assertion: {
        target: { output_handle_id: 'output-json-0', path: 'total' },
        condition: { kind: 'equals', expected: 1234.56 },
      },
      name: 'Q1 invoice total',
    });

    expect(mockClient.lastFetchParams).toMatchObject({
      url: '/workflows/tests',
      method: 'POST',
      body: {
        workflow_id: 'wf_abc123',
        target: { type: 'block', block_id: 'block_extract' },
        source: { type: 'manual', handle_inputs: {} },
        assertion: {
          target: { output_handle_id: 'output-json-0', path: 'total' },
          condition: { kind: 'equals', expected: 1234.56 },
        },
        name: 'Q1 invoice total',
      },
    });
    expect(test.id).toBe('wfnodetest_abc');
  });

  test("omits `name` from the body when the caller doesn't supply it", async () => {
    const mockClient = new MockClient(TEST_RESPONSE);
    const tests = new APIWorkflowTests(mockClient);

    await tests.create({
      workflowId: 'wf_abc123',
      target: { type: 'block', block_id: 'block_extract' },
      source: {
        type: 'run_step',
        run_id: 'wfrun_xyz',
        step_id: 'block_extract',
      },
      assertion: {
        target: { output_handle_id: 'output-json-0', path: 'total' },
        condition: { kind: 'exists' },
      },
    });

    const body = (mockClient.lastFetchParams as { body: Record<string, unknown> }).body;
    expect('name' in body).toBe(false);
    expect(body.source).toEqual({
      type: 'run_step',
      run_id: 'wfrun_xyz',
      step_id: 'block_extract',
    });
  });
});

describe('workflows.tests CRUD URL shapes', () => {
  test('get() uses the test detail route', async () => {
    const mockClient = new MockClient(TEST_RESPONSE);
    const tests = new APIWorkflowTests(mockClient);

    await tests.get({ testId: 'wfnodetest_abc' });

    expect(mockClient.lastFetchParams).toMatchObject({
      url: '/workflows/tests/wfnodetest_abc',
      method: 'GET',
    });
  });

  test('list() uses the tests route with `target_block_id` filter', async () => {
    const mockClient = new MockClient({
      data: [],
      list_metadata: { before: null, after: null },
    });
    const tests = new APIWorkflowTests(mockClient);

    const result = await tests.list({
      workflowId: 'wf_abc123',
      targetBlockId: 'block_extract',
      limit: 25,
    });

    expect(mockClient.lastFetchParams).toMatchObject({
      url: '/workflows/tests?workflow_id=wf_abc123',
      method: 'GET',
      params: { limit: 25, target_block_id: 'block_extract' },
    });
    expect(result.data).toEqual([]);
    expect(result.list_metadata.before).toBeNull();
    expect(result.list_metadata.after).toBeNull();
  });

  test('list() omits target_block_id when undefined', async () => {
    // Passing `target_block_id: undefined` would have the backend
    // string-coerce to the literal "undefined" — make sure the SDK
    // omits the param entirely.
    const mockClient = new MockClient({
      data: [],
      list_metadata: { before: null, after: null },
    });
    const tests = new APIWorkflowTests(mockClient);

    await tests.list({ workflowId: 'wf_abc123' });

    const params = (mockClient.lastFetchParams as { params: Record<string, unknown> }).params;
    expect('target_block_id' in params).toBe(false);
  });

  test('delete() uses the test detail route with DELETE', async () => {
    // Body shape from the backend is empty (204) — the SDK doesn't
    // try to parse it.
    const mockClient = new MockClient(null);
    const tests = new APIWorkflowTests(mockClient);

    await tests.delete({
      testId: 'wfnodetest_abc',
    });

    expect(mockClient.lastFetchParams).toMatchObject({
      url: '/workflows/tests/wfnodetest_abc',
      method: 'DELETE',
    });
  });
});

describe('workflows.tests.update()', () => {
  test('PATCH body only carries the fields the caller passed', async () => {
    // Including a field with `undefined` would either trip Pydantic
    // validation or — worse — clear the field on the stored doc.
    // The SDK MUST omit, not nullify.
    const mockClient = new MockClient(TEST_RESPONSE);
    const tests = new APIWorkflowTests(mockClient);

    await tests.update({
      testId: 'wfnodetest_abc',
      name: 'renamed',
    });

    expect(mockClient.lastFetchParams).toMatchObject({
      url: '/workflows/tests/wfnodetest_abc',
      method: 'PATCH',
      body: { name: 'renamed' },
    });
    const body = (mockClient.lastFetchParams as { body: Record<string, unknown> }).body;
    expect(Object.keys(body)).toEqual(['name']);
  });

  test('PATCH with assertion only includes assertion in the body', async () => {
    const mockClient = new MockClient(TEST_RESPONSE);
    const tests = new APIWorkflowTests(mockClient);

    await tests.update({
      testId: 'wfnodetest_abc',
      assertion: {
        target: { output_handle_id: 'output-json-0', path: 'vendor.name' },
        condition: { kind: 'matches_regex', expected: '^Acme.*' },
      },
    });

    const body = (mockClient.lastFetchParams as { body: Record<string, unknown> }).body;
    expect(Object.keys(body)).toEqual(['assertion']);
    expect(body.assertion).toMatchObject({
      condition: { kind: 'matches_regex' },
    });
  });
});

describe('workflows.tests.runs.create()', () => {
  test('runs.create({testId}) posts only test_id', async () => {
    const mockClient = new MockClient({ ...RUN_RESPONSE, lifecycle: PENDING });
    const tests = new APIWorkflowTests(mockClient);

    const response = await tests.runs.create({
      workflowId: 'wf_abc123',
      testId: 'wfnodetest_abc',
    });

    expect(mockClient.lastFetchParams).toMatchObject({
      url: '/workflows/tests/runs',
      method: 'POST',
      body: { workflow_id: 'wf_abc123', test_id: 'wfnodetest_abc' },
    });
    expect(response.lifecycle.status).toBe('pending');
    expect(response.id).toBe('wftestrun_q1z2');
  });

  test('runs.create({target, nConsensus}) posts target + n_consensus snake_case', async () => {
    const mockClient = new MockClient({ ...RUN_RESPONSE, lifecycle: PENDING });
    const tests = new APIWorkflowTests(mockClient);

    await tests.runs.create({
      workflowId: 'wf_abc123',
      target: { type: 'block', block_id: 'block_extract' },
      nConsensus: 5,
    });

    expect(mockClient.lastFetchParams).toMatchObject({
      body: {
        workflow_id: 'wf_abc123',
        target: { type: 'block', block_id: 'block_extract' },
        n_consensus: 5,
      },
    });
  });

  test('runs.create({}) posts workflow_id to run every test in the workflow', async () => {
    const mockClient = new MockClient({ ...RUN_RESPONSE, lifecycle: PENDING });
    const tests = new APIWorkflowTests(mockClient);

    await tests.runs.create({ workflowId: 'wf_abc123' });

    const body = (mockClient.lastFetchParams as { body: Record<string, unknown> }).body;
    expect(body).toEqual({ workflow_id: 'wf_abc123' });
  });
});

describe('workflows.tests.runs', () => {
  test('runs.list() uses the canonical runs route', async () => {
    const mockClient = new MockClient({
      data: [],
      list_metadata: { before: null, after: null },
    });
    const tests = new APIWorkflowTests(mockClient);

    const result = await tests.runs.list({
      workflowId: 'wf_abc123',
      testId: 'wfnodetest_abc',
      limit: 10,
    });

    expect(mockClient.lastFetchParams).toMatchObject({
      url: '/workflows/tests/runs',
      method: 'GET',
      params: {
        limit: 10,
        workflow_id: 'wf_abc123',
        test_id: 'wfnodetest_abc',
      },
    });
    expect(result.data).toEqual([]);
    expect(result.list_metadata.before).toBeNull();
    expect(result.list_metadata.after).toBeNull();
  });

  test('runs.get() uses the run-id-first route', async () => {
    const mockClient = new MockClient(RUN_RESPONSE);
    const tests = new APIWorkflowTests(mockClient);

    const run = await tests.runs.get({
      runId: 'wftestrun_q1z2',
    });

    expect(mockClient.lastFetchParams).toMatchObject({
      url: '/workflows/tests/runs/wftestrun_q1z2',
      method: 'GET',
    });
    expect(run.lifecycle.status).toBe('completed');
  });

  test('runs.cancel() uses the run-id-first route', async () => {
    const mockClient = new MockClient({ ...RUN_RESPONSE, lifecycle: { status: 'cancelled' } });
    const tests = new APIWorkflowTests(mockClient);

    const run = await tests.runs.cancel({
      runId: 'wftestrun_q1z2',
    });

    expect(mockClient.lastFetchParams).toMatchObject({
      url: '/workflows/tests/runs/wftestrun_q1z2/cancel',
      method: 'POST',
    });
    expect(run.lifecycle.status).toBe('cancelled');
  });

  test('runs.results.list() uses the flat results route with a run filter', async () => {
    const mockClient = new MockClient({
      data: [RESULT_RESPONSE],
      list_metadata: { before: null, after: null },
    });
    const tests = new APIWorkflowTests(mockClient);

    const result = await tests.runs.results.list({
      runId: 'wftestrun_q1z2',
    });

    expect(mockClient.lastFetchParams).toMatchObject({
      url: '/workflows/tests/results',
      method: 'GET',
      params: { run_id: 'wftestrun_q1z2', limit: 20 },
    });
    expect(result.data[0]?.test_id).toBe('wfnodetest_abc');
    expect(result.data[0]?.outputs).toEqual({ 'output-json-0': { total: 1234.56 } });
  });

  test('runs.results.get() uses the flat result id route', async () => {
    const mockClient = new MockClient(RESULT_RESPONSE);
    const tests = new APIWorkflowTests(mockClient);

    const result = await tests.runs.results.get({
      resultId: 'wfresult_abc',
    });

    expect(mockClient.lastFetchParams).toMatchObject({
      url: '/workflows/tests/results/wfresult_abc',
      method: 'GET',
    });
    expect(result.id).toBe('wfresult_abc');
    expect(result.test_id).toBe('wfnodetest_abc');
  });

  test('legacy execute and scoped run aliases are not exposed', () => {
    const tests = new APIWorkflowTests(new MockClient({}));
    expect('execute' in tests).toBe(false);
    expect('getExecution' in tests.runs).toBe(false);
    expect('get_execution' in tests.runs).toBe(false);
    expect(typeof tests.runs.results).toBe('object');
    expect(typeof tests.runs.results.get).toBe('function');
  });
});

describe('status enum coverage', () => {
  test('Z* drift enums match the OpenAPI literals', async () => {
    const { ZAssertionDriftStatus, ZSchemaDriftStatus, ZWorkflowTest } = await import(
      '../src/api/workflows/tests/types'
    );

    for (const value of ['valid', 'drifted', 'broken']) {
      expect(() => ZAssertionDriftStatus.parse(value)).not.toThrow();
    }
    expect(() => ZAssertionDriftStatus.parse('fresh')).toThrow();

    for (const value of ['none', 'partial', 'drifted', 'unknown']) {
      expect(() => ZSchemaDriftStatus.parse(value)).not.toThrow();
      expect(() =>
        ZWorkflowTest.parse({
          ...TEST_RESPONSE,
          assertion_drift_status: 'valid',
          schema_drift: value,
        })
      ).not.toThrow();
    }
    expect(() => ZSchemaDriftStatus.parse('fresh')).toThrow();
  });

  test('Z* schemas accept all 7 WorkflowTestRunStatus values', async () => {
    // Pin: a future status addition (or accidental shrinkage) trips
    // this test before it ships. Without this, a `cancelled` run
    // record from the backend would runtime-fail in the SDK.
    const { ZWorkflowTestRunStatus } = await import('../src/api/workflows/tests/types');
    for (const value of [
      'queued',
      'running',
      'passed',
      'failed',
      'blocked',
      'error',
      'cancelled',
    ]) {
      expect(() => ZWorkflowTestRunStatus.parse(value)).not.toThrow();
    }
  });

  test('Z* batch counts schema declares all 7 status buckets', async () => {
    const { ZWorkflowTestBatchExecutionCounts } = await import('../src/api/workflows/tests/types');
    const parsed = ZWorkflowTestBatchExecutionCounts.parse({});
    for (const key of ['queued', 'running', 'passed', 'failed', 'blocked', 'error', 'cancelled']) {
      expect(parsed).toHaveProperty(key, 0);
    }
  });

  test('workflow test run lifecycle/timing use the dedicated public state machine', async () => {
    const { ZWorkflowTestRun, ZWorkflowTestResult } = await import(
      '../src/api/workflows/tests/types'
    );

    expect(() =>
      ZWorkflowTestRun.parse({
        ...RUN_RESPONSE,
        lifecycle: { status: 'awaiting_review' },
      })
    ).toThrow();
    expect(() =>
      ZWorkflowTestResult.parse({
        ...RESULT_RESPONSE,
        lifecycle: { status: 'awaiting_review' },
      })
    ).toThrow();

    const parsed = ZWorkflowTestRun.parse({
      ...RUN_RESPONSE,
      timing: {
        ...TIMING,
        review_waiting_started_at: NOW,
      },
    });
    expect(Object.keys(parsed.timing).sort()).toEqual(['completed_at', 'created_at', 'started_at']);
  });
});

// ---------------------------------------------------------------------------
// Round 2 — gap coverage caught in second-pass review
// ---------------------------------------------------------------------------

describe('workflows.tests parity bridge installation', () => {
  test('APIWorkflowTests has canonical prepare_* aliases installed at runtime', () => {
    // The parity bridge auto-installs `prepare_create`, `prepare_get`, etc.
    // from `PYTHON_PUBLIC_PREPARE_METHODS` in `src/client.ts`. A typo in
    // either the registry key or the method name would silently disable
    // the bridge — pin the actual installed methods so a regression
    // breaks here, not in the python_parity test (which fails fast on
    // the FIRST violation and is harder to debug).
    const tests = new APIWorkflowTests(new MockClient({})) as unknown as Record<string, unknown>;
    for (const name of [
      'prepare_create',
      'prepare_get',
      'prepare_list',
      'prepare_update',
      'prepare_delete',
      'prepare_create_run',
    ]) {
      expect(typeof tests[name]).toBe('function');
    }
  });

  test('APIWorkflowTestRuns has canonical prepare_* aliases installed', () => {
    const runs = new APIWorkflowTestRuns(new MockClient({})) as unknown as Record<string, unknown>;
    expect(typeof runs.prepare_create).toBe('function');
    expect(typeof runs.prepare_list).toBe('function');
    expect(typeof runs.prepare_get).toBe('function');
    expect(typeof runs.prepare_cancel).toBe('function');
  });
});

describe('workflows.tests response parsing edge cases', () => {
  test('ZWorkflowTest accepts `assertion: null` from a legacy doc', async () => {
    // Pre-rewrite tests in storage may have `assertion=None`. The SDK's
    // schema declares the field as `.nullable().optional()` to mirror
    // the backend's relaxed `WorkflowTest.assertion: AssertionSpec | None
    // = None`. Without this, listing a workflow with even one orphan
    // would crash the client at parse time.
    const { ZWorkflowTest } = await import('../src/api/workflows/tests/types');
    const legacy = { ...TEST_RESPONSE, assertion: null };
    const parsed = ZWorkflowTest.parse(legacy);
    expect(parsed.assertion).toBeNull();
  });

  test('runs.list() correctly parses a populated data[] array', async () => {
    // The empty-array path is covered above. Pin the populated path so
    // a future schema drift (renaming `outputs` again, dropping
    // `verdict_summary`, etc.) breaks here — not in production with a
    // real run-history fetch.
    const populatedRunsResponse = {
      data: [
        RUN_RESPONSE,
        { ...RUN_RESPONSE, id: 'wftestrun_b', lifecycle: { status: 'running' } },
        {
          ...RUN_RESPONSE,
          id: 'wftestrun_c',
          lifecycle: { status: 'cancelled' },
        },
      ],
      list_metadata: { before: null, after: null },
    };
    const mockClient = new MockClient(populatedRunsResponse);
    const tests = new APIWorkflowTests(mockClient);

    const result = await tests.runs.list({
      workflowId: 'wf_abc123',
      testId: 'wfnodetest_abc',
    });

    expect(result.data).toHaveLength(3);
    expect(result.data[0]?.lifecycle.status).toBe('completed');
    expect(result.data[1]?.lifecycle.status).toBe('running');
    expect(result.data[2]?.lifecycle.status).toBe('cancelled');
  });

  test('parent run schema strips legacy job and batch ids', async () => {
    const { ZWorkflowTestRun } = await import('../src/api/workflows/tests/types');
    const parsed = ZWorkflowTestRun.parse({
      ...RUN_RESPONSE,
      job_id: 'job_legacy',
      batch_id: 'batch_legacy',
    });

    expect('job_id' in parsed).toBe(false);
    expect('batch_id' in parsed).toBe(false);
    expect(parsed.id).toBe('wftestrun_q1z2');
    expect(parsed.lifecycle.status).toBe('completed');
  });
});

describe('workflows.tests param-merge precedence (sibling parity)', () => {
  test('list() lets caller-supplied options.params override the typed `limit`', async () => {
    // Matches APIWorkflowRuns.list precedence — the typed kwarg is the
    // default, but the `options.params` escape hatch wins. Allows
    // power users to pass through query params without forking the
    // SDK signature.
    const mockClient = new MockClient({
      data: [],
      list_metadata: { before: null, after: null },
    });
    const tests = new APIWorkflowTests(mockClient);

    await tests.list({ workflowId: 'wf_abc123', limit: 50 }, { params: { limit: 100 } });

    const params = (mockClient.lastFetchParams as { params: Record<string, unknown> }).params;
    expect(params.limit).toBe(100);
  });

  test('runs.list() also lets options.params override the typed `limit`', async () => {
    const mockClient = new MockClient({
      data: [],
      list_metadata: { before: null, after: null },
    });
    const tests = new APIWorkflowTests(mockClient);

    await tests.runs.list(
      { workflowId: 'wf_abc123', testId: 'wfnodetest_abc', limit: 10 },
      { params: { limit: 25 } }
    );

    const params = (mockClient.lastFetchParams as { params: Record<string, unknown> }).params;
    expect(params.limit).toBe(25);
  });
});
