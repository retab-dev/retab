import { describe, expect, test } from 'bun:test';

import { AbstractClient } from '../src/client';
import APIWorkflows from '../src/api/workflows/client';
import APIWorkflowRuns from '../src/api/workflows/runs/client';
import APIWorkflowBlocks from '../src/api/workflows/blocks/client';
import APIWorkflowEdges from '../src/api/workflows/edges/client';
import APIWorkflowSpecs from '../src/api/workflows/specs/client';
import APIWorkflowSimulations from '../src/api/workflows/simulations/client';
import { ZStepExecutionResponse, ZWorkflowRun, ZWorkflowRunStep } from '../src/types';
import { ZHandlePayload } from '../src/generated_types';
import type { WorkflowRunExportResponse } from '../src/types';

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

// Build a minimal v2 WorkflowRun JSON payload for fixtures.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function makeV2Run(overrides: Record<string, any> = {}): Record<string, any> {
  const {
    id = 'run_1',
    workflow = {
      workflow_id: 'wf_1',
      version_id: 'ver_0123456789abcdef0123456789abcdef',
      name_at_run_time: 'Test',
    },
    trigger = { type: 'manual' },
    lifecycle = { status: 'running' },
    timing = {
      created_at: '2026-01-01T00:00:00Z',
      started_at: '2026-01-01T00:00:00Z',
    },
    ...rest
  } = overrides;
  return {
    id,
    workflow,
    trigger,
    lifecycle,
    timing,
    ...rest,
  };
}

describe('workflows client', () => {
  test('get() uses the workflow detail route', async () => {
    const mockClient = new MockClient({
      id: 'workflow_123',
      name: 'Test Workflow',
      description: '',
      published: null,
      email_trigger: { allowed_senders: [], allowed_domains: [] },
      created_at: '2026-03-12T10:00:00Z',
      updated_at: '2026-03-12T10:00:00Z',
    });
    const workflowsClient = new APIWorkflows(mockClient);

    const workflow = await workflowsClient.get('workflow_123');

    expect(mockClient.lastFetchParams).toEqual({
      url: '/workflows/workflow_123',
      method: 'GET',
      params: undefined,
      headers: undefined,
    });
    expect(workflow.id).toBe('workflow_123');
  });

  test('list() uses the workflows route with pagination params', async () => {
    const mockClient = new MockClient({
      data: [],
      list_metadata: { before: null, after: 'workflow_after' },
    });
    const workflowsClient = new APIWorkflows(mockClient);

    const result = await workflowsClient.list({
      limit: 5,
      order: 'asc',
      sortBy: 'updated_at',
      fields: 'id,name',
      after: 'workflow_before',
    });

    expect(mockClient.lastFetchParams).toEqual({
      url: '/workflows',
      method: 'GET',
      params: {
        limit: 5,
        order: 'asc',
        sort_by: 'updated_at',
        fields: 'id,name',
        after: 'workflow_before',
      },
      headers: undefined,
    });
    expect(result.list_metadata.after).toBe('workflow_after');
  });

  test('update() accepts an email trigger policy', async () => {
    const mockClient = new MockClient({
      id: 'workflow_123',
      name: 'Test Workflow',
      description: '',
      published: null,
      email_trigger: {
        allowed_senders: ['ops@example.com'],
        allowed_domains: ['example.com'],
      },
      created_at: '2026-03-12T10:00:00Z',
      updated_at: '2026-03-12T10:00:00Z',
    });
    const workflowsClient = new APIWorkflows(mockClient);

    const workflow = await workflowsClient.update('workflow_123', {
      emailTrigger: {
        allowedSenders: ['ops@example.com'],
        allowedDomains: ['example.com'],
      },
    });

    expect(mockClient.lastFetchParams).toEqual({
      url: '/workflows/workflow_123',
      method: 'PATCH',
      body: {
        email_trigger: {
          allowed_senders: ['ops@example.com'],
          allowed_domains: ['example.com'],
        },
      },
      params: undefined,
      headers: undefined,
    });
    expect(workflow.email_trigger.allowed_senders).toEqual(['ops@example.com']);
  });

  test('exposes specs subresource', () => {
    const workflowsClient = new APIWorkflows(new MockClient({}));

    expect(workflowsClient.specs).toBeInstanceOf(APIWorkflowSpecs);
  });

  test('specs.validate() uses the spec validate route', async () => {
    const mockClient = new MockClient({
      workflow_id: 'wf_1',
      block_count: 2,
      edge_count: 1,
      is_valid: true,
      diagnostics: { issues: [] },
    });
    const specsClient = new APIWorkflowSpecs(mockClient);

    const response = await specsClient.validate('spec: {}\n');

    expect(mockClient.lastFetchParams).toEqual({
      url: '/workflows/spec/validate',
      method: 'POST',
      body: { yaml_definition: 'spec: {}\n' },
      params: undefined,
      headers: undefined,
    });
    expect(response.is_valid).toBe(true);
  });

  test('specs.plan() uses the spec plan route', async () => {
    const mockClient = new MockClient({
      workflow_id: 'wf_1',
      action: 'noop',
      block_count: 2,
      edge_count: 1,
      diagnostics: { issues: [] },
      format_version: 'workflows-plan/v1',
      summary: { add: 0, change: 1, destroy: 0, replace: 0, noop: 0, total: 1, has_changes: true },
      resource_changes: [
        {
          address: 'workflow.wf_1.block.block_extract',
          target: 'block',
          target_id: 'block_extract',
          name: 'Extract',
          type: 'extract',
          actions: ['update'],
          summary: "Update block 'Extract'",
          change: {
            before: { config: { model: 'old' } },
            after: { config: { model: 'new' } },
            before_sensitive: {},
            after_sensitive: {},
            field_changes: [
              {
                path: ['config', 'model'],
                path_display: 'config.model',
                action: 'update',
                before: 'old',
                after: 'new',
              },
            ],
          },
        },
      ],
      rendered_plan: 'Plan: 0 to add, 1 to change, 0 to destroy.',
    });
    const specsClient = new APIWorkflowSpecs(mockClient);

    const response = await specsClient.plan('spec: {}\n');

    expect(mockClient.lastFetchParams).toEqual({
      url: '/workflows/spec/plan',
      method: 'POST',
      body: { yaml_definition: 'spec: {}\n' },
      params: undefined,
      headers: undefined,
    });
    expect(response.summary.change).toBe(1);
    expect(response.resource_changes[0].change.field_changes[0].path_display).toBe('config.model');
    expect(response.rendered_plan).toContain('1 to change');
  });

  test('specs.apply() uses the spec apply route', async () => {
    const mockClient = new MockClient({
      workflow_id: 'wf_1',
      created: false,
      block_count: 2,
      edge_count: 1,
      diagnostics: { issues: [] },
      format_version: 'workflows-plan/v1',
      summary: { add: 0, change: 0, destroy: 0, replace: 0, noop: 1, total: 0, has_changes: false },
      resource_changes: [],
      rendered_plan: 'No changes. Infrastructure is up-to-date.',
    });
    const specsClient = new APIWorkflowSpecs(mockClient);

    const response = await specsClient.apply('spec: {}\n');

    expect(mockClient.lastFetchParams).toEqual({
      url: '/workflows/spec/apply',
      method: 'POST',
      body: { yaml_definition: 'spec: {}\n' },
      params: undefined,
      headers: undefined,
    });
    expect(response.workflow_id).toBe('wf_1');
    expect(response.summary.noop).toBe(1);
    expect(response.resource_changes).toEqual([]);
  });

  test('specs.export() uses the spec export route', async () => {
    const mockClient = new MockClient({
      workflow_id: 'wf_1',
      yaml_definition: 'apiVersion: workflows.retab.com/v1alpha2\n',
    });
    const specsClient = new APIWorkflowSpecs(mockClient);

    const response = await specsClient.export('wf_1');

    expect(mockClient.lastFetchParams).toEqual({
      url: '/workflows/wf_1/spec',
      method: 'GET',
      params: undefined,
      headers: undefined,
    });
    expect(response.yaml_definition.startsWith('apiVersion:')).toBe(true);
  });

  test('create() sends POST to /workflows', async () => {
    const mockClient = new MockClient({
      id: 'wf_new',
      name: 'My Workflow',
      description: 'A test',
      published: null,
      email_trigger: { allowed_senders: [], allowed_domains: [] },
      created_at: '2026-03-12T10:00:00Z',
      updated_at: '2026-03-12T10:00:00Z',
    });
    const workflowsClient = new APIWorkflows(mockClient);

    const wf = await workflowsClient.create({ name: 'My Workflow', description: 'A test' });

    expect(mockClient.lastFetchParams?.url).toBe('/workflows');
    expect(mockClient.lastFetchParams?.method).toBe('POST');
    expect(wf.id).toBe('wf_new');
    expect(wf.name).toBe('My Workflow');
  });

  test('publish() sends POST to /workflows/{id}/publish', async () => {
    const mockClient = new MockClient({
      id: 'wf_1',
      name: 'Test',
      published: {
        version_id: 'ver_0123456789abcdef0123456789abcdef',
        published_at: '2026-03-12T10:00:00Z',
      },
      email_trigger: { allowed_senders: [], allowed_domains: [] },
      created_at: '2026-03-12T10:00:00Z',
      updated_at: '2026-03-12T10:00:00Z',
    });
    const workflowsClient = new APIWorkflows(mockClient);

    const wf = await workflowsClient.publish('wf_1', { description: 'v1' });

    expect(mockClient.lastFetchParams?.url).toBe('/workflows/wf_1/publish');
    expect(mockClient.lastFetchParams?.method).toBe('POST');
    expect(wf.published?.version_id).toBe('ver_0123456789abcdef0123456789abcdef');
  });

  test('removed workflow methods are not exposed', () => {
    const workflowsClient = new APIWorkflows(new MockClient({}));

    expect('duplicate' in workflowsClient).toBe(false);
    expect('getEntities' in workflowsClient).toBe(false);
    expect('get_entities' in workflowsClient).toBe(false);
    expect('getResolvedSchemas' in workflowsClient).toBe(false);
    expect('get_resolved_schemas' in workflowsClient).toBe(false);
    expect('prepare_get_resolved_schemas' in workflowsClient).toBe(false);
    expect('listSnapshots' in workflowsClient).toBe(false);
    expect('list_snapshots' in workflowsClient).toBe(false);
    expect('prepare_list_snapshots' in workflowsClient).toBe(false);
  });

  test('stale workflow aggregate models are not exported', async () => {
    const types = await import('../src/types');

    expect('ZWorkflowWithEntities' in types).toBe(false);
    expect('WorkflowWithEntities' in types).toBe(false);
    expect('ZWorkflowResolvedSchemasResponse' in types).toBe(false);
    expect('WorkflowResolvedSchemasResponse' in types).toBe(false);
    expect('ZBlockResolvedSchemasResponse' in types).toBe(false);
    expect('BlockResolvedSchemasResponse' in types).toBe(false);
  });

  test('prepare_diagnose exposes the Python prepared-request surface', () => {
    const workflowsClient = new APIWorkflows(new MockClient({}));

    expect(workflowsClient.prepare_diagnose('wf_1', false)).toEqual({
      url: '/workflows/wf_1/diagnose-graph',
      method: 'POST',
      body: {
        re_propagate: false,
      },
    });
    expect(workflowsClient.prepare_diagnose('wf_1').body.re_propagate).toBe(true);
  });

  test('diagnoseGraph preserves warning-only diagnoses', async () => {
    const workflowsClient = new APIWorkflows(
      new MockClient({
        is_valid: true,
        issues: [
          {
            severity: 'warning',
            code: 'MISSING_REVIEW_PREDICATE',
            message: 'Review gate needs a predicate',
            block_id: 'extract_1',
          },
        ],
        suggestions: [],
        stats: {
          total_blocks: 1,
          total_edges: 0,
          block_types: { start: 1 },
          start_document_blocks: 1,
        },
      })
    );

    const diagnosis = await workflowsClient.diagnoseGraph('wf_1', {
      blocks: [{ id: 'start-1', type: 'start_document' }],
      edges: [],
    });

    expect(diagnosis.is_valid).toBe(true);
    expect(diagnosis.issues[0]?.severity).toBe('warning');
    expect(diagnosis.issues[0]?.code).toBe('MISSING_REVIEW_PREDICATE');
  });

  test('diagnose() posts directly to the diagnose route', async () => {
    const mockClient = new MockClient({
      is_valid: true,
      issues: [],
      suggestions: [],
      stats: {
        total_blocks: 1,
        total_edges: 0,
        block_types: { start_document: 1 },
        start_document_blocks: 1,
      },
    });
    const workflowsClient = new APIWorkflows(mockClient);

    await workflowsClient.diagnose('wf_1', { rePropagate: false });

    expect(mockClient.lastFetchParams).toEqual({
      url: '/workflows/wf_1/diagnose-graph',
      method: 'POST',
      body: {
        re_propagate: false,
      },
      params: undefined,
      headers: undefined,
    });
  });

  test('artifacts prepare helpers expose the Python prepared-request surface', () => {
    const workflowsClient = new APIWorkflows(new MockClient({}));

    expect(workflowsClient.artifacts.prepare_list('run_1', 'extraction', 'extract-1')).toEqual({
      url: '/workflows/artifacts',
      method: 'GET',
      params: {
        run_id: 'run_1',
        operation: 'extraction',
        block_id: 'extract-1',
      },
    });
  });

  test('experiments expose Python snake_case and prepare helpers', () => {
    const workflowsClient = new APIWorkflows(new MockClient({}));

    expect('cancel' in workflowsClient.experiments).toBe(false);
    expect('get_metrics' in workflowsClient.experiments).toBe(false);
    expect('getMetrics' in workflowsClient.experiments).toBe(false);
    expect('get_content' in workflowsClient.experiments).toBe(false);
    expect('getContent' in workflowsClient.experiments).toBe(false);
    expect('duplicate' in workflowsClient.experiments).toBe(false);
    expect('prepare_duplicate' in workflowsClient.experiments).toBe(false);
    expect('listEligibleBlocks' in workflowsClient.experiments).toBe(false);
    expect('list_eligible_blocks' in workflowsClient.experiments).toBe(false);
    expect('prepare_list_eligible_blocks' in workflowsClient.experiments).toBe(false);
    expect('prepare_run_batch' in workflowsClient.experiments).toBe(false);
    expect('run_batch' in workflowsClient.experiments).toBe(false);
    expect('runBatch' in workflowsClient.experiments).toBe(false);
    expect(typeof workflowsClient.experiments.runs.get).toBe('function');
    expect(typeof workflowsClient.experiments.runs.cancel).toBe('function');
    expect(typeof workflowsClient.experiments.runs.results.get).toBe('function');
    expect(typeof workflowsClient.experiments.runs.metrics.get).toBe('function');
    expect('get_content' in workflowsClient.experiments.runs).toBe(false);
    expect('getContent' in workflowsClient.experiments.runs).toBe(false);
    expect('cancel_document' in workflowsClient.experiments.runs).toBe(false);
    expect('get_job' in workflowsClient.experiments.runs).toBe(false);
    expect('getJob' in workflowsClient.experiments.runs).toBe(false);
    expect('wait_for_completion' in workflowsClient.experiments.runs).toBe(false);
    expect('waitForCompletion' in workflowsClient.experiments.runs).toBe(false);
    expect(workflowsClient.experiments.runs.prepare_get('run_1')).toEqual({
      url: '/workflows/experiments/runs/run_1',
      method: 'GET',
    });
    expect(workflowsClient.experiments.runs.prepare_create('wf_1', 'exp_1')).toEqual({
      url: '/workflows/experiments/runs',
      method: 'POST',
      body: {
        experiment_id: 'exp_1',
        workflow_id: 'wf_1',
      },
    });
    expect('prepare_run_document' in workflowsClient.experiments.runs).toBe(false);
    expect('run_document' in workflowsClient.experiments.runs).toBe(false);
    expect(workflowsClient.experiments.runs.prepare_cancel('run_1')).toEqual({
      url: '/workflows/experiments/runs/run_1/cancel',
      method: 'POST',
      body: {},
    });
  });

  test('delete() sends DELETE to /workflows/{id}', async () => {
    const mockClient = new MockClient(null);
    const workflowsClient = new APIWorkflows(mockClient);

    await workflowsClient.delete('wf_1');

    expect(mockClient.lastFetchParams?.url).toBe('/workflows/wf_1');
    expect(mockClient.lastFetchParams?.method).toBe('DELETE');
  });

  test('runs.get() ignores legacy steps field (now fetched via workflows.steps.list())', async () => {
    // Older servers may still include `steps` in the run payload; the SDK
    // no longer surfaces it on the WorkflowRun type. Per-step records are
    // fetched via `client.workflows.steps.list(run_id)`.
    const mockClient = new MockClient({
      id: 'run_123',
      workflow: {
        workflow_id: 'workflow_123',
        version_id: 'ver_0123456789abcdef0123456789abcdef',
        name_at_run_time: 'Classifier Workflow',
      },
      trigger: { type: 'manual' },
      lifecycle: { status: 'running' },
      timing: {
        created_at: '2026-03-13T10:00:00Z',
        started_at: '2026-03-13T10:00:00Z',
      },
      steps: [
        {
          block_id: 'classifier-1',
          step_id: 'classifier-1',
          block_type: 'classifier',
          block_label: 'Classifier',
          status: 'completed',
        },
      ],
    });
    const runsClient = new APIWorkflowRuns(mockClient);

    const run = await runsClient.get('run_123');

    expect(mockClient.lastFetchParams).toEqual({
      url: '/workflows/runs/run_123',
      method: 'GET',
      params: undefined,
      headers: undefined,
    });
    // The legacy `steps` payload is silently dropped by the Zod schema —
    // it is no longer a property on the parsed WorkflowRun.
    expect((run as unknown as Record<string, unknown>).steps).toBeUndefined();
    expect(run.id).toBe('run_123');
    expect(run.lifecycle.status).toBe('running');
  });

  test('runs.v2 typed fields parse correctly', async () => {
    const mockClient = new MockClient({
      id: 'run_789',
      workflow: {
        workflow_id: 'wf_1',
        version_id: 'ver_abcdef0123456789abcdef0123456789',
        name_at_run_time: 'Test',
        requested_version: 'ver_abcdef0123456789abcdef0123456789',
      },
      trigger: { type: 'api', api_key_id: 'ak_1' },
      lifecycle: { status: 'completed' },
      timing: {
        created_at: '2026-01-01T00:00:00Z',
        started_at: '2026-01-01T00:00:00Z',
        completed_at: '2026-01-01T00:00:05Z',
        accumulated_review_waiting_ms: 5000,
      },
      inputs: { documents: {}, json_data: {} },
    });
    const runsClient = new APIWorkflowRuns(mockClient);

    const run = await runsClient.get('run_789');

    expect(run.trigger.type).toBe('api');
    expect(run.workflow.workflow_id).toBe('wf_1');
    expect(run.workflow.version_id).toBe('ver_abcdef0123456789abcdef0123456789');
    expect(run.timing.accumulated_review_waiting_ms).toBe(5000);
  });

  test('runs.list() serializes array and date filters', async () => {
    const mockClient = new MockClient({
      data: [],
      list_metadata: { before: null, after: null },
    });
    const runsClient = new APIWorkflowRuns(mockClient);

    await runsClient.list({
      workflowId: 'wf_1',
      statuses: ['completed', 'error'],
      triggerTypes: ['api', 'email'],
      fromDate: new Date('2026-01-01T00:00:00.000Z'),
      toDate: new Date('2026-01-31T00:00:00.000Z'),
      fields: ['id', 'status'],
      after: 'run_after',
    });

    expect(mockClient.lastFetchParams).toEqual({
      url: '/workflows/runs',
      method: 'GET',
      params: {
        workflow_id: 'wf_1',
        statuses: 'completed,error',
        trigger_types: 'api,email',
        from_date: '2026-01-01',
        to_date: '2026-01-31',
        fields: 'id,status',
        after: 'run_after',
        limit: 20,
        order: 'desc',
      },
      headers: undefined,
    });
  });

  test('removed block methods are not exposed', () => {
    const blocksClient = new APIWorkflowBlocks(new MockClient({}));

    expect('configHistory' in blocksClient).toBe(false);
    expect('config_history' in blocksClient).toBe(false);
    expect('prepare_config_history' in blocksClient).toBe(false);
    expect('getResolvedSchemas' in blocksClient).toBe(false);
    expect('get_resolved_schemas' in blocksClient).toBe(false);
    expect('prepare_get_resolved_schemas' in blocksClient).toBe(false);
    expect('createBatch' in blocksClient).toBe(false);
    expect('create_batch' in blocksClient).toBe(false);
    expect('listSimulations' in blocksClient).toBe(false);
    expect('list_simulations' in blocksClient).toBe(false);
    expect('prepare_list_simulations' in blocksClient).toBe(false);
    expect('simulate' in blocksClient).toBe(false);
    expect('prepare_simulate' in blocksClient).toBe(false);
  });

  test('workflows expose simulations subresource', () => {
    const workflowsClient = new APIWorkflows(new MockClient({}));

    expect(workflowsClient.simulations).toBeInstanceOf(APIWorkflowSimulations);
  });

  test('workflow simulations create uses top-level route', async () => {
    const mockClient = new MockClient({
      id: 'sim_1',
      workflow_id: 'wf_1',
      run_id: 'run_1',
      block_id: 'block_1',
      block_type: 'extract',
      success: true,
      created_at: '2026-03-12T10:00:00Z',
    });
    const simulationsClient = new APIWorkflowSimulations(mockClient);

    const simulation = await simulationsClient.create({
      runId: 'run_1',
      blockId: 'block_1',
      stepId: 'step_1',
      nConsensus: 5,
    });

    expect(mockClient.lastFetchParams).toEqual({
      url: '/workflows/simulations',
      method: 'POST',
      body: {
        run_id: 'run_1',
        block_id: 'block_1',
        step_id: 'step_1',
        n_consensus: 5,
      },
      params: undefined,
      headers: undefined,
    });
    expect('workflow_id' in (mockClient.lastFetchParams?.body as Record<string, unknown>)).toBe(
      false
    );
    expect(simulation.id).toBe('sim_1');
  });

  test('workflow simulations list uses top-level route', async () => {
    const mockClient = new MockClient({
      data: [
        {
          id: 'sim_1',
          workflow_id: 'wf_1',
          run_id: 'run_1',
          block_id: 'block_1',
          block_type: 'extract',
          success: true,
          created_at: '2026-03-12T10:00:00Z',
        },
      ],
      list_metadata: { before: null, after: null },
    });
    const simulationsClient = new APIWorkflowSimulations(mockClient);

    const result = await simulationsClient.list({
      runId: 'run_1',
      blockId: 'block_1',
      limit: 10,
    });

    expect(mockClient.lastFetchParams).toEqual({
      url: '/workflows/simulations',
      method: 'GET',
      params: {
        run_id: 'run_1',
        block_id: 'block_1',
        limit: 10,
      },
      headers: undefined,
    });
    expect(result.data[0].id).toBe('sim_1');
  });

  test('workflow blocks expose live-editing metadata', async () => {
    const mockClient = new MockClient({
      id: 'extract-1',
      workflow_id: 'wf_1',
      draft_version: 'draft_1',
      type: 'extract',
      label: 'Extract',
    });
    const blocksClient = new APIWorkflowBlocks(mockClient);

    const block = await blocksClient.get('wf_1', 'extract-1');

    expect(block.draft_version).toBe('draft_1');
  });

  test('workflow block objects ignore resolved schema sidecars', async () => {
    const mockClient = new MockClient({
      id: 'extract-1',
      workflow_id: 'wf_1',
      draft_version: 'draft_1',
      type: 'extract',
      label: 'Extract',
      resolved_schemas: {
        input_schemas: {},
        output_schemas: {
          'output-json-0': {
            type: 'object',
            properties: { invoice_number: { type: 'string' } },
          },
        },
      },
    });
    const blocksClient = new APIWorkflowBlocks(mockClient);

    const block = await blocksClient.get('wf_1', 'extract-1');

    expect('resolved_schemas' in block).toBe(false);
  });

  test('step execution responses ignore removed payload schemas', () => {
    const removedPayloadKey = ['raw', '_', 'output'].join('');
    const parsed = ZStepExecutionResponse.parse({
      block_id: 'extract-1',
      block_type: 'extract',
      block_label: 'Extract',
      lifecycle: { status: 'completed' },
      output: {
        data: { invoice_number: 'INV-001' },
        json_schema: { type: 'object' },
      },
      [removedPayloadKey]: {
        json_schema: { type: 'object' },
      },
      json_schema: { type: 'object' },
      handle_outputs: {
        'output-json-0': {
          type: 'json',
          data: { invoice_number: 'INV-001' },
        },
      },
      handle_inputs: {},
    });

    expect('output' in parsed).toBe(false);
    expect(removedPayloadKey in parsed).toBe(false);
    expect('json_schema' in parsed).toBe(false);
    expect('artifact_view' in parsed).toBe(false);
    expect('status' in parsed).toBe(false);
    expect('terminal' in parsed).toBe(false);
    expect(parsed.handle_outputs?.['output-json-0']?.data?.invoice_number).toBe('INV-001');
  });

  test('step execution responses accept json_ref handle outputs', () => {
    const parsed = ZStepExecutionResponse.parse({
      block_id: 'function-1',
      block_type: 'function',
      block_label: 'Function',
      lifecycle: { status: 'completed' },
      handle_outputs: {
        'output-json-0': {
          type: 'json_ref',
          artifact_ref: {
            operation: 'workflow_step_json',
            id: 'artifact_123',
            key: 'output-json-0',
          },
          preview: { truncated: true },
        },
      },
      handle_inputs: {},
    });

    expect(parsed.handle_outputs?.['output-json-0']?.type).toBe('json_ref');
    expect(parsed.handle_outputs?.['output-json-0']?.artifact_ref?.id).toBe('artifact_123');
    expect(parsed.handle_outputs?.['output-json-0']?.preview?.truncated).toBe(true);
  });

  test('handle payloads reject removed text type', () => {
    expect(() => ZHandlePayload.parse({ type: 'text', text: 'removed' })).toThrow();
  });

  test('workflow run steps expose top-level model', () => {
    const parsed = ZWorkflowRunStep.parse({
      run_id: 'run_123',
      block_id: 'extract-1',
      step_id: 'extract-1',
      block_type: 'extract',
      block_label: 'Extract',
      lifecycle: { status: 'completed' },
      model: 'retab-small',
      handle_outputs: {},
      handle_inputs: {},
    });

    expect(parsed.model).toBe('retab-small');
    expect('status' in parsed).toBe(false);
    expect('terminal' in parsed).toBe(false);
  });

  test('removed edge methods are not exposed', () => {
    const edgesClient = new APIWorkflowEdges(new MockClient({}));

    expect('createBatch' in edgesClient).toBe(false);
    expect('create_batch' in edgesClient).toBe(false);
    expect('deleteAll' in edgesClient).toBe(false);
    expect('delete_all' in edgesClient).toBe(false);
  });

  test('runs.create() passes FileRef documents without content', async () => {
    const mockClient = new MockClient(makeV2Run());
    const runsClient = new APIWorkflowRuns(mockClient);

    await runsClient.create({
      workflowId: 'wf_1',
      documents: {
        start_1: {
          id: 'file_existing',
          filename: 'invoice.pdf',
          mime_type: 'application/pdf',
        },
      },
    });

    expect(mockClient.lastFetchParams).toEqual({
      url: '/workflows/runs',
      method: 'POST',
      body: {
        workflow_id: 'wf_1',
        documents: {
          start_1: {
            id: 'file_existing',
            filename: 'invoice.pdf',
            mime_type: 'application/pdf',
          },
        },
        version: 'production',
      },
      params: undefined,
      headers: undefined,
    });
    expect(JSON.stringify(mockClient.lastFetchParams)).not.toContain('content');
  });

  test('runs.create() keeps MIMEData content for new documents', async () => {
    const mockClient = new MockClient(makeV2Run());
    const runsClient = new APIWorkflowRuns(mockClient);

    await runsClient.create({
      workflowId: 'wf_1',
      documents: {
        start_1: {
          filename: 'note.txt',
          url: 'data:text/plain;base64,aGVsbG8=',
        },
      },
    });

    expect(mockClient.lastFetchParams).toEqual({
      url: '/workflows/runs',
      method: 'POST',
      body: {
        workflow_id: 'wf_1',
        documents: {
          start_1: {
            filename: 'note.txt',
            content: 'aGVsbG8=',
            mime_type: 'text/plain',
          },
        },
        version: 'production',
      },
      params: undefined,
      headers: undefined,
    });
  });

  test('runs.cancel() sends POST to /cancel', async () => {
    const mockClient = new MockClient({
      run: makeV2Run({ lifecycle: { status: 'cancelled' } }),
      cancellation_status: 'cancelled',
    });
    const runsClient = new APIWorkflowRuns(mockClient);

    const result = await runsClient.cancel('run_1', { commandId: 'cmd_1' });

    expect(mockClient.lastFetchParams).toEqual({
      url: '/workflows/runs/run_1/cancel',
      method: 'POST',
      body: { command_id: 'cmd_1' },
      params: undefined,
      headers: undefined,
    });
    expect(result.cancellation_status).toBe('cancelled');
  });

  test('runs.restart() sends POST to flat run creation route', async () => {
    const mockClient = new MockClient(makeV2Run({ id: 'run_2', lifecycle: { status: 'running' } }));
    const runsClient = new APIWorkflowRuns(mockClient);

    const run = await runsClient.restart('run_1', { commandId: 'cmd_2' });

    expect(mockClient.lastFetchParams).toEqual({
      url: '/workflows/runs',
      method: 'POST',
      body: { restart_of: 'run_1', command_id: 'cmd_2' },
      params: undefined,
      headers: undefined,
    });
    expect(run.id).toBe('run_2');
  });

  test('runs.export() sends POST to /export-payload', async () => {
    const mockClient = new MockClient({
      csv_data: 'a,b\n1,2\n',
      rows: 1,
      columns: 2,
    } satisfies WorkflowRunExportResponse);
    const runsClient = new APIWorkflowRuns(mockClient);

    const result = await runsClient.export({
      workflowId: 'wf_1',
      blockId: 'extract-1',
      exportSource: 'outputs',
      selectedRunIds: ['run_1', 'run_2'],
      fromDate: new Date('2026-01-01T00:00:00.000Z'),
      toDate: new Date('2026-01-31T00:00:00.000Z'),
      triggerTypes: ['api'],
      preferredColumns: ['invoice_number', 'total_amount'],
    });

    expect(mockClient.lastFetchParams).toEqual({
      url: '/workflows/runs/export-payload',
      method: 'POST',
      body: {
        workflow_id: 'wf_1',
        block_id: 'extract-1',
        export_source: 'outputs',
        preferred_columns: ['invoice_number', 'total_amount'],
        selected_run_ids: ['run_1', 'run_2'],
        from_date: '2026-01-01',
        to_date: '2026-01-31',
        trigger_types: ['api'],
      },
      params: undefined,
      headers: undefined,
    });
    expect(result.rows).toBe(1);
  });
});

describe('workflow runs', () => {
  test('workflow runs do not expose wait helpers', () => {
    const runsClient = new APIWorkflowRuns(new MockClient({}));

    expect('waitForCompletion' in runsClient).toBe(false);
    expect('wait_for_completion' in runsClient).toBe(false);
    expect('createAndWait' in runsClient).toBe(false);
    expect('getConfig' in runsClient).toBe(false);
    expect('executionOrder' in runsClient).toBe(false);
    expect('getDocumentUrl' in runsClient).toBe(false);
  });
});
