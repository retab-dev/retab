/**
 * Pin the canonical PaginatedList envelope (`{data, list_metadata}`) for the
 * two list endpoints under `/v1/workflows/experiments?workflow_id={wf}`:
 *
 *   GET /v1/workflows/experiments?workflow_id={wf}
 *   GET /v1/workflows/experiments/runs?workflow_id=...&experiment_id=...
 *
 * Both routes used to return non-canonical shapes (a bare list and the
 * `{"runs": [...]}` named-key envelope respectively). The migration to the
 * shared envelope is a deliberate breaking change.
 */
import { describe, expect, test } from 'bun:test';

import { AbstractClient } from '../src/client';
import APIWorkflowExperiments from '../src/api/workflows/experiments/client';
import APIWorkflowExperimentRuns from '../src/api/workflows/experiments/runs/client';

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
  workflow_id: 'wf_1',
  version_id: 'draft_1',
  name_at_run_time: 'Q1',
  requested_version: 'draft',
};
const TRIGGER = { type: 'api' };
const TIMING = { created_at: NOW, started_at: NOW, completed_at: NOW };

const EXPERIMENT = {
  id: 'exp_1',
  workflow_id: 'wf_1',
  block_id: 'block_1',
  n_consensus: 5,
  document_count: 0,
  name: 'Q1',
  last_run_id: null,
  created_at: NOW,
  updated_at: NOW,
  status: 'draft',
  block_kind: 'extract',
  score: null,
  is_stale: false,
  schema_drift: 'unknown',
  schema_drift_detail: null,
};

const RUN = {
  id: 'exprun_1',
  workflow: WORKFLOW_REF,
  trigger: TRIGGER,
  lifecycle: { status: 'completed' },
  timing: TIMING,
  experiment_id: 'exp_1',
  block_id: 'block_1',
  block_kind: 'extract',
  n_consensus: 5,
  definition_fingerprint: 'fp',
  documents_fingerprint: 'fp_doc',
  score: 0.83,
  total_document_count: 1,
  completed_document_count: 1,
  document_count: 1,
  error_count: 0,
};

const RESULT = {
  id: 'expresult_1',
  run_id: 'exprun_1',
  experiment_id: 'exp_1',
  document_id: 'expdoc_1',
  lifecycle: { status: 'completed' },
  timing: TIMING,
  block_kind: 'extract',
  handle_inputs: { 'input-file-0': { type: 'file' } },
  artifact: { operation: 'extraction', id: 'ext_1' },
  attempt: 1,
};

describe('workflows.experiments.list paginated envelope', () => {
  test('experiment schema drift accepts all OpenAPI literals', async () => {
    const { ZExperimentResponse, ZSchemaDriftStatus } = await import(
      '../src/api/workflows/experiments/types'
    );

    for (const value of ['none', 'partial', 'drifted', 'unknown']) {
      expect(() => ZSchemaDriftStatus.parse(value)).not.toThrow();
      expect(() => ZExperimentResponse.parse({ ...EXPERIMENT, schema_drift: value })).not.toThrow();
    }
  });

  test('returns {data, list_metadata}', async () => {
    const mockClient = new MockClient({
      data: [EXPERIMENT],
      list_metadata: { before: null, after: null },
    });
    const experiments = new APIWorkflowExperiments(mockClient);

    const page = await experiments.list('wf_1');

    expect(mockClient.lastFetchParams).toMatchObject({
      url: '/workflows/experiments?workflow_id=wf_1',
      method: 'GET',
    });
    expect(page.data).toHaveLength(1);
    expect(page.data[0]?.id).toBe('exp_1');
    expect(page.list_metadata.before).toBeNull();
    expect(page.list_metadata.after).toBeNull();
  });
});

describe('workflows.experiments.runs.list paginated envelope', () => {
  test('runs.cancel parses the compact cancel response shape', async () => {
    const mockClient = new MockClient({
      id: 'exprun_1',
      lifecycle: { status: 'cancelled' },
    });
    const runs = new APIWorkflowExperimentRuns(mockClient);

    const response = await runs.cancel({ runId: 'exprun_1' });

    expect(mockClient.lastFetchParams).toMatchObject({
      url: '/workflows/experiments/runs/exprun_1/cancel',
      method: 'POST',
      body: {},
    });
    expect(response).toEqual({
      id: 'exprun_1',
      lifecycle: { status: 'cancelled' },
    });
  });

  test('returns {data, list_metadata} (was {runs: [...]})', async () => {
    const mockClient = new MockClient({
      data: [RUN],
      list_metadata: { before: null, after: null },
    });
    const runs = new APIWorkflowExperimentRuns(mockClient);

    const page = await runs.list({
      workflowId: 'wf_1',
      experimentId: 'exp_1',
    });

    expect(mockClient.lastFetchParams).toMatchObject({
      url: '/workflows/experiments/runs',
      method: 'GET',
      params: { workflow_id: 'wf_1', experiment_id: 'exp_1', limit: 20 },
    });
    expect(page.data).toHaveLength(1);
    expect(page.data[0]?.id).toBe('exprun_1');
    expect(page.data[0]?.timing.started_at).toBe(NOW);
    expect(page.list_metadata.before).toBeNull();
    expect(page.list_metadata.after).toBeNull();
  });

  test('serializes workflow-run-style filters', async () => {
    const mockClient = new MockClient({
      data: [],
      list_metadata: { before: null, after: null },
    });
    const runs = new APIWorkflowExperimentRuns(mockClient);

    await runs.list({
      workflowId: 'wf_1',
      experimentId: 'exp_1',
      blockId: 'block_1',
      status: 'completed',
      statuses: ['completed', 'error'],
      excludeStatus: 'cancelled',
      triggerType: 'api',
      triggerTypes: ['api', 'manual_run'],
      fromDate: new Date('2026-05-01T00:00:00.000Z'),
      toDate: '2026-05-18',
      sortBy: 'created_at',
      fields: ['id', 'lifecycle'],
      before: 'exprun_before',
      after: 'exprun_after',
      limit: 10,
      order: 'asc',
    });

    expect(mockClient.lastFetchParams).toMatchObject({
      url: '/workflows/experiments/runs',
      method: 'GET',
      params: {
        workflow_id: 'wf_1',
        experiment_id: 'exp_1',
        block_id: 'block_1',
        status: 'completed',
        statuses: 'completed,error',
        exclude_status: 'cancelled',
        trigger_type: 'api',
        trigger_types: 'api,manual_run',
        from_date: '2026-05-01',
        to_date: '2026-05-18',
        sort_by: 'created_at',
        fields: 'id,lifecycle',
        before: 'exprun_before',
        after: 'exprun_after',
        limit: 10,
        order: 'asc',
      },
    });
  });

  test('results.get uses the flat result id route', async () => {
    const mockClient = new MockClient(RESULT);
    const runs = new APIWorkflowExperimentRuns(mockClient);

    const result = await runs.results.get({ resultId: 'expresult_1' });

    expect(mockClient.lastFetchParams).toMatchObject({
      url: '/workflows/experiments/results/expresult_1',
      method: 'GET',
    });
    expect(result.id).toBe('expresult_1');
    expect(result.document_id).toBe('expdoc_1');
  });

  test('parent run schema strips legacy job and batch ids', async () => {
    const { ZExperimentRun } = await import('../src/api/workflows/experiments/types');
    const parsed = ZExperimentRun.parse({
      ...RUN,
      job_id: 'job_legacy',
      batch_id: 'batch_legacy',
    });

    expect('job_id' in parsed).toBe(false);
    expect('batch_id' in parsed).toBe(false);
    expect(parsed.id).toBe('exprun_1');
    expect(parsed.lifecycle.status).toBe('completed');
  });
});
