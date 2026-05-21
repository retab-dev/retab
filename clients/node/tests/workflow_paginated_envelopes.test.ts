import { describe, expect, test } from 'bun:test';

import { AbstractClient } from '../src/client';
import APIWorkflowBlocks from '../src/api/workflows/blocks/client';
import APIWorkflowEdges from '../src/api/workflows/edges/client';
import APIWorkflowArtifacts from '../src/api/workflows/artifacts/client';
import APIWorkflowSteps from '../src/api/workflows/steps/client';

// Regression tests pinning the canonical
// `{ data, list_metadata: { before: null, after: null } }` envelope on the
// four workflow GRAPH list endpoints converted from bare arrays.
//
//   GET /v1/workflows/blocks?workflow_id={wf}                                     -> PaginatedList[WorkflowBlock]
//   GET /v1/workflows/edges?workflow_id={wf}                                      -> PaginatedList[WorkflowEdgeDoc]
//   GET /v1/workflows/artifacts                                       -> PaginatedList[WorkflowArtifact]
//   GET /v1/workflows/steps?run_id={run}                              -> PaginatedList[WorkflowRunStep]

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

    expect((client.lastFetchParams as { url: string }).url).toBe(
      '/workflows/blocks?workflow_id=wf_aaa'
    );
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

    expect((client.lastFetchParams as { url: string }).url).toBe(
      '/workflows/edges?workflow_id=wf_aaa'
    );
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

  test('steps.list()', async () => {
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
    const steps = new APIWorkflowSteps(client);
    const result = await steps.list('run_aaa');

    expect((client.lastFetchParams as { url: string }).url).toBe('/workflows/steps?run_id=run_aaa');
    expect(result.data).toHaveLength(1);
    expect(result.list_metadata).toEqual({ before: null, after: null });
  });
});
