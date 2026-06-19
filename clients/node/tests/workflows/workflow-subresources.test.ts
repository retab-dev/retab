// Offline request-contract tests for the workflow list sub-resources
// (runs, steps, blocks, edges, artifacts, tests, reviews, experiments).
//
// Mirrors Python tests/test_workflow_paginated_envelopes.py: each list method
// is exercised with a captured-fetch injected into `new Retab({ apiKey, fetch })`.
// We assert the HTTP method + path, the snake_cased workflow_id / run_id query,
// and that a `{ data, list_metadata }` envelope deserializes into a
// `PaginatedList` carrying the right items. No server is contacted.
//
// NEVER edits src/.

import { describe, expect, test } from 'bun:test';

import { PaginatedList, Retab } from '../../src/index.js';

type Captured = {
  method: string;
  url: URL;
  path: string;
  query: Record<string, string>;
};

function clientCapturing(response: unknown): { client: Retab; calls: Captured[] } {
  const calls: Captured[] = [];
  const client = new Retab({
    apiKey: 'test_key',
    fetch: (async (input: URL | RequestInfo, init?: RequestInit) => {
      const url = new URL(String(input));
      const query: Record<string, string> = {};
      for (const [k, v] of url.searchParams.entries()) query[k] = v;
      calls.push({ method: String(init?.method), url, path: url.pathname, query });
      return new Response(JSON.stringify(response), {
        status: 200,
        headers: { 'content-type': 'application/json' },
      });
    }) as typeof fetch,
  });
  return { client, calls };
}

function envelope(...items: Record<string, unknown>[]): {
  data: Record<string, unknown>[];
  list_metadata: { before: string | null; after: string | null };
} {
  return { data: items, list_metadata: { before: null, after: null } };
}

function lastCall(calls: Captured[]): Captured {
  expect(calls.length).toBeGreaterThan(0);
  return calls[calls.length - 1];
}

describe('workflow sub-resource list envelopes', () => {
  test('runs.list filters by workflow_id and deserializes WorkflowRun items', async () => {
    const { client, calls } = clientCapturing(
      envelope({
        id: 'run_1',
        workflow_id: 'wf_aaa',
        workflow_version_id: 'ver_1',
        trigger: { type: 'manual' },
        lifecycle: { status: 'pending' },
        timing: { created_at: '2026-01-01T00:00:00Z' },
      })
    );

    const result = await client.workflows.runs.list({ workflowId: 'wf_aaa' });

    const call = lastCall(calls);
    expect(call.method).toBe('GET');
    expect(call.path).toBe('/v1/workflows/runs');
    expect(call.query.workflow_id).toBe('wf_aaa');
    expect(result).toBeInstanceOf(PaginatedList);
    expect(result.data.length).toBe(1);
    expect(result.data[0].id).toBe('run_1');
    expect(result.data[0].workflowId).toBe('wf_aaa');
    expect(result.list_metadata.before).toBeNull();
    expect(result.list_metadata.after).toBeNull();
  });

  test('steps.list filters by run_id and deserializes WorkflowRunStep items', async () => {
    const { client, calls } = clientCapturing(
      envelope({
        block_id: 'extract-1',
        step_id: 'step-1',
        block_type: 'extract',
        block_label: 'Extract',
        lifecycle: { status: 'completed' },
        run_id: 'run_aaa',
        retry_count: 0,
      })
    );

    const result = await client.workflows.steps.list({ runId: 'run_aaa' });

    const call = lastCall(calls);
    expect(call.method).toBe('GET');
    expect(call.path).toBe('/v1/workflows/steps');
    expect(call.query.run_id).toBe('run_aaa');
    expect(result).toBeInstanceOf(PaginatedList);
    expect(result.data.length).toBe(1);
    expect(result.data[0].blockId).toBe('extract-1');
    expect(result.data[0].runId).toBe('run_aaa');
  });

  test('blocks.list requires + sends workflow_id and deserializes WorkflowBlock items', async () => {
    const { client, calls } = clientCapturing(
      envelope({
        id: 'extract-1',
        workflow_id: 'wf_aaa',
        type: 'extract',
        label: 'Extract',
        position_x: 0,
        position_y: 0,
        updated_at: '2026-01-01T00:00:00Z',
      })
    );

    const result = await client.workflows.blocks.list({ workflowId: 'wf_aaa' });

    const call = lastCall(calls);
    expect(call.method).toBe('GET');
    expect(call.path).toBe('/v1/workflows/blocks');
    expect(call.query.workflow_id).toBe('wf_aaa');
    expect(result).toBeInstanceOf(PaginatedList);
    expect(result.data.length).toBe(1);
    expect(result.data[0].id).toBe('extract-1');
    expect(result.data[0].workflowId).toBe('wf_aaa');
    expect(result.data[0].type).toBe('extract');
  });

  test('edges.list sends workflow_id and deserializes WorkflowEdgeDoc items', async () => {
    const { client, calls } = clientCapturing(
      envelope({
        id: 'edge-1',
        workflow_id: 'wf_aaa',
        source_block: 'start-1',
        target_block: 'extract-1',
        source_handle: 'output-file-0',
        target_handle: 'input-file-0',
        updated_at: '2026-01-01T00:00:00Z',
      })
    );

    const result = await client.workflows.edges.list({ workflowId: 'wf_aaa' });

    const call = lastCall(calls);
    expect(call.method).toBe('GET');
    expect(call.path).toBe('/v1/workflows/edges');
    expect(call.query.workflow_id).toBe('wf_aaa');
    expect(result).toBeInstanceOf(PaginatedList);
    expect(result.data.length).toBe(1);
    expect(result.data[0].id).toBe('edge-1');
    expect(result.data[0].sourceBlock).toBe('start-1');
    expect(result.data[0].targetHandle).toBe('input-file-0');
  });

  test('edges.get sends optional workflow_id scope and deserializes WorkflowEdgeDoc', async () => {
    const { client, calls } = clientCapturing({
      id: 'edge-1',
      workflow_id: 'wf_aaa',
      source_block: 'start-1',
      target_block: 'extract-1',
      source_handle: 'output-file-0',
      target_handle: 'input-file-0',
      updated_at: '2026-01-01T00:00:00Z',
    });

    const result = await client.workflows.edges.get('edge-1', { workflowId: 'wf_aaa' });

    const call = lastCall(calls);
    expect(call.method).toBe('GET');
    expect(call.path).toBe('/v1/workflows/edges/edge-1');
    expect(call.query.workflow_id).toBe('wf_aaa');
    expect(result.id).toBe('edge-1');
    expect(result.workflowId).toBe('wf_aaa');
    expect(result.targetBlock).toBe('extract-1');
  });

  test('artifacts.list filters by run_id and deserializes WorkflowArtifact items', async () => {
    const { client, calls } = clientCapturing(
      envelope({ operation: 'extraction', id: 'ext_123' }, { operation: 'parse', id: 'parse_456' })
    );

    const result = await client.workflows.artifacts.list({ runId: 'run_aaa' });

    const call = lastCall(calls);
    expect(call.method).toBe('GET');
    expect(call.path).toBe('/v1/workflows/artifacts');
    expect(call.query.run_id).toBe('run_aaa');
    expect(result).toBeInstanceOf(PaginatedList);
    expect(result.data.length).toBe(2);
    expect(result.data[0].operation).toBe('extraction');
    expect(result.data[0].id).toBe('ext_123');
    expect(result.data[1].id).toBe('parse_456');
  });

  test('tests.list sends workflow_id and deserializes WorkflowEval items', async () => {
    const { client, calls } = clientCapturing(
      envelope({
        id: 'test_1',
        workflow_id: 'wf_aaa',
        target: { type: 'block', block_id: 'extract-1' },
        source: { type: 'manual' },
        name: 'extract smoke',
      })
    );

    const result = await client.workflows.evals.list({ workflowId: 'wf_aaa' });

    const call = lastCall(calls);
    expect(call.method).toBe('GET');
    expect(call.path).toBe('/v1/workflows/evals');
    expect(call.query.workflow_id).toBe('wf_aaa');
    expect(result).toBeInstanceOf(PaginatedList);
    expect(result.data.length).toBe(1);
    expect(result.data[0].id).toBe('test_1');
    expect(result.data[0].workflowId).toBe('wf_aaa');
    expect(result.data[0].target.blockId).toBe('extract-1');
  });

  test('reviews.list filters by workflow_id/run_id and deserializes Review items', async () => {
    const { client, calls } = clientCapturing(
      envelope({
        id: 'review_1',
        workflow_id: 'wf_aaa',
        workflow_version_id: 'ver_1',
        run_id: 'run_aaa',
        block_id: 'extract-1',
        step_id: 'step-1',
        parent_step_id: null,
        iteration_key: null,
        block_type: 'extract',
        triggered_by: { kind: 'always' },
        created_at: '2026-01-01T00:00:00Z',
        decision: null,
      })
    );

    const result = await client.workflows.reviews.list({ workflowId: 'wf_aaa', runId: 'run_aaa' });

    const call = lastCall(calls);
    expect(call.method).toBe('GET');
    expect(call.path).toBe('/v1/workflows/reviews');
    expect(call.query.workflow_id).toBe('wf_aaa');
    expect(call.query.run_id).toBe('run_aaa');
    expect(result).toBeInstanceOf(PaginatedList);
    expect(result.data.length).toBe(1);
    expect(result.data[0].id).toBe('review_1');
    expect(result.data[0].workflowId).toBe('wf_aaa');
    expect(result.data[0].runId).toBe('run_aaa');
  });

  test('experiments.list sends workflow_id and deserializes WorkflowExperiment items', async () => {
    const { client, calls } = clientCapturing(
      envelope({
        id: 'exp_1',
        workflow_id: 'wf_aaa',
        block_id: 'extract-1',
        n_consensus: 1,
        document_count: 3,
        name: 'baseline',
        last_run_id: null,
        status: 'idle',
        block_type: 'extract',
      })
    );

    const result = await client.workflows.experiments.list({ workflowId: 'wf_aaa' });

    const call = lastCall(calls);
    expect(call.method).toBe('GET');
    expect(call.path).toBe('/v1/workflows/experiments');
    expect(call.query.workflow_id).toBe('wf_aaa');
    expect(result).toBeInstanceOf(PaginatedList);
    expect(result.data.length).toBe(1);
    expect(result.data[0].id).toBe('exp_1');
    expect(result.data[0].workflowId).toBe('wf_aaa');
    expect(result.data[0].documentCount).toBe(3);
  });

  test('an empty envelope deserializes into an empty PaginatedList', async () => {
    const { client } = clientCapturing(envelope());
    const result = await client.workflows.runs.list({ workflowId: 'wf_empty' });
    expect(result).toBeInstanceOf(PaginatedList);
    expect(result.data.length).toBe(0);
    expect(result.hasMore()).toBe(false);
  });
});
