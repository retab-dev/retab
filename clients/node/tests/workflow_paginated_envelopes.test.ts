import { describe, expect, test } from 'bun:test';

import { AbstractClient } from '../src/client';
import APIWorkflows from '../src/api/workflows/client';
import APIWorkflowBlocks from '../src/api/workflows/blocks/client';
import APIWorkflowEdges from '../src/api/workflows/edges/client';
import APIWorkflowArtifacts from '../src/api/workflows/artifacts/client';
import APIWorkflowRunSteps from '../src/api/workflows/runs/steps/client';

// Regression tests pinning the canonical
// `{ data, list_metadata: { before: null, after: null } }` envelope on the
// seven workflow GRAPH list endpoints converted from bare arrays.
//
//   GET /v1/workflows/{wf}/blocks                                     -> PaginatedList[WorkflowBlock]
//   GET /v1/workflows/{wf}/blocks/{id}/config-history                 -> PaginatedList[BlockConfigVersion]
//   GET /v1/workflows/{wf}/edges                                      -> PaginatedList[WorkflowEdgeDoc]
//   GET /v1/workflows/artifacts                                       -> PaginatedList[WorkflowArtifact]
//   GET /v1/workflows/{wf}/snapshots                                  -> PaginatedList[WorkflowSnapshot]
//   GET /v1/workflows/runs/{run}/steps                                -> PaginatedList[WorkflowRunStep]
//   GET /v1/workflows/runs/{run}/steps/{block}/simulations            -> PaginatedList[BlockSimulation]

class EnvelopeClient extends AbstractClient {
  public lastFetchParams: Record<string, unknown> | null = null;
  constructor(public payload: Record<string, unknown>) {
    super();
  }
  protected async _fetch(params: {
    url: string;
    method: string;
    params?: Record<string, unknown>;
    headers?: Record<string, unknown>;
    body?: Record<string, unknown>;
  }): Promise<Response> {
    this.lastFetchParams = params;
    return new Response(JSON.stringify(this.payload), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    });
  }
}

function envelope(...items: Array<Record<string, unknown>>): Record<string, unknown> {
  return { data: items, list_metadata: { before: null, after: null } };
}

describe('workflow graph list endpoints return canonical paginated envelope', () => {
  test('blocks.list()', async () => {
    const client = new EnvelopeClient(
      envelope({
        id: 'extract-1',
        workflow_id: 'wf_aaa',
        type: 'extract',
        label: 'Extract',
        position_x: 0,
        position_y: 0,
      })
    );
    const blocks = new APIWorkflowBlocks(client);
    const result = await blocks.list('wf_aaa');

    expect((client.lastFetchParams as { url: string }).url).toBe('/workflows/wf_aaa/blocks');
    expect(result.data).toHaveLength(1);
    expect(result.list_metadata).toEqual({ before: null, after: null });
  });

  test('blocks.configHistory()', async () => {
    const client = new EnvelopeClient(
      envelope({
        config_fingerprint: 'fp_1',
        block_type: 'extract',
        block_label: 'Extract',
        first_seen_at: '2026-01-01T00:00:00Z',
        last_seen_at: '2026-01-02T00:00:00Z',
        snapshot_versions: [1, 2],
        run_count: 7,
        is_current: true,
      })
    );
    const blocks = new APIWorkflowBlocks(client);
    const result = await blocks.configHistory('wf_aaa', 'extract-1');

    expect((client.lastFetchParams as { url: string }).url).toBe(
      '/workflows/wf_aaa/blocks/extract-1/config-history'
    );
    expect(result.data).toHaveLength(1);
    expect(result.list_metadata).toEqual({ before: null, after: null });
  });

  test('blocks.listSimulations()', async () => {
    const client = new EnvelopeClient(
      envelope({
        id: 'sim_1',
        workflow_id: 'wf_aaa',
        run_id: 'run_aaa',
        block_id: 'extract-1',
        block_type: 'extract',
        success: true,
      })
    );
    const blocks = new APIWorkflowBlocks(client);
    const result = await blocks.listSimulations('run_aaa', 'extract-1', { limit: 5 });

    expect((client.lastFetchParams as { url: string }).url).toBe(
      '/workflows/runs/run_aaa/steps/extract-1/simulations'
    );
    expect((client.lastFetchParams as { params: Record<string, unknown> }).params).toEqual({
      limit: 5,
    });
    expect(result.data).toHaveLength(1);
    expect(result.list_metadata).toEqual({ before: null, after: null });
  });

  test('edges.list()', async () => {
    const client = new EnvelopeClient(
      envelope({
        id: 'edge-1',
        workflow_id: 'wf_aaa',
        organization_id: 'org_1',
        source_block: 'start-1',
        target_block: 'extract-1',
      })
    );
    const edges = new APIWorkflowEdges(client);
    const result = await edges.list('wf_aaa');

    expect((client.lastFetchParams as { url: string }).url).toBe('/workflows/wf_aaa/edges');
    expect(result.data).toHaveLength(1);
    expect(result.list_metadata).toEqual({ before: null, after: null });
  });

  test('artifacts.list()', async () => {
    const client = new EnvelopeClient(envelope({ operation: 'extraction', id: 'ext_123' }));
    const artifacts = new APIWorkflowArtifacts(client);
    const result = await artifacts.list({ runId: 'run_aaa' });

    expect((client.lastFetchParams as { url: string }).url).toBe('/workflows/artifacts');
    expect(result.data).toHaveLength(1);
    expect(result.list_metadata).toEqual({ before: null, after: null });
  });

  test('runs.steps.list()', async () => {
    const client = new EnvelopeClient(
      envelope({
        run_id: 'run_aaa',
        organization_id: 'org_1',
        block_id: 'extract-1',
        step_id: 'extract-1',
        block_type: 'extract',
        block_label: 'Extract',
        lifecycle: { status: 'completed' },
      })
    );
    const steps = new APIWorkflowRunSteps(client);
    const result = await steps.list('run_aaa');

    expect((client.lastFetchParams as { url: string }).url).toBe('/workflows/runs/run_aaa/steps');
    expect(result.data).toHaveLength(1);
    expect(result.list_metadata).toEqual({ before: null, after: null });
  });

  test('listSnapshots()', async () => {
    const client = new EnvelopeClient(
      envelope({
        id: 'wfsn_1',
        snapshot_id: 'snap_1',
        workflow_id: 'wf_aaa',
        version: 3,
        description: 'third release',
        block_count: 4,
        edge_count: 3,
        published_at: '2026-04-01T12:00:00Z',
      })
    );
    const workflows = new APIWorkflows(client);
    const result = await workflows.listSnapshots('wf_aaa', { limit: 25 });

    expect((client.lastFetchParams as { url: string }).url).toBe('/workflows/wf_aaa/snapshots');
    expect((client.lastFetchParams as { params: Record<string, unknown> }).params).toEqual({
      limit: 25,
    });
    expect(result.data).toHaveLength(1);
    expect(result.list_metadata).toEqual({ before: null, after: null });
  });
});
