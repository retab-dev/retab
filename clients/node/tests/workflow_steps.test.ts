import { describe, expect, test } from 'bun:test';

import { AbstractClient } from '../src/client';
import APIWorkflowArtifacts from '../src/api/workflows/artifacts/client';
import APIWorkflowRunSteps from '../src/api/workflows/runs/steps/client';
import * as workflowTypes from '../src/types';

class MockClient extends AbstractClient {
  public lastFetchParams: Record<string, unknown> | null = null;

  protected async _fetch(params: {
    url: string;
    method: string;
    params?: Record<string, unknown>;
    headers?: Record<string, unknown>;
    body?: Record<string, unknown>;
  }): Promise<Response> {
    this.lastFetchParams = params;
    return new Response(
      JSON.stringify({
        data: [
          {
            run_id: 'run_123',
            block_id: 'extract-1',
            step_id: 'extract-1',
            block_type: 'extract',
            block_label: 'Extract',
            lifecycle: { status: 'completed' },
            artifact: {
              operation: 'extraction',
              id: 'ext_123',
            },
          },
        ],
        list_metadata: { before: null, after: null },
      }),
      {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }
    );
  }
}

describe('workflow run steps client', () => {
  test('artifacts.get() accepts a step artifact ref', async () => {
    class ArtifactMockClient extends AbstractClient {
      public lastFetchParams: Record<string, unknown> | null = null;

      protected async _fetch(params: {
        url: string;
        method: string;
        params?: Record<string, unknown>;
        headers?: Record<string, unknown>;
        body?: Record<string, unknown>;
      }): Promise<Response> {
        this.lastFetchParams = params;
        return new Response(
          JSON.stringify({
            operation: 'hil_evaluation',
            id: 'heval_123',
            requires_human_review: true,
          }),
          {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
          }
        );
      }
    }

    const mockClient = new ArtifactMockClient();
    const artifactsClient = new APIWorkflowArtifacts(mockClient);
    const artifact = await artifactsClient.get({
      operation: 'hil_evaluation',
      id: 'heval_123',
    });

    expect(mockClient.lastFetchParams).toEqual({
      url: '/workflows/artifacts/hil_evaluation/heval_123',
      method: 'GET',
      params: undefined,
      headers: undefined,
    });
    expect(artifact.operation).toBe('hil_evaluation');
    expect(artifact.id).toBe('heval_123');
    expect(artifact.requires_human_review).toBe(true);
  });

  test('artifacts.list() uses run scoped artifact route and returns paginated envelope', async () => {
    class ArtifactListMockClient extends AbstractClient {
      public lastFetchParams: Record<string, unknown> | null = null;

      protected async _fetch(params: {
        url: string;
        method: string;
        params?: Record<string, unknown>;
        headers?: Record<string, unknown>;
        body?: Record<string, unknown>;
      }): Promise<Response> {
        this.lastFetchParams = params;
        return new Response(
          JSON.stringify({
            data: [{ operation: 'conditional_evaluation', id: 'ceval_123' }],
            list_metadata: { before: null, after: null },
          }),
          {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
          }
        );
      }
    }

    const mockClient = new ArtifactListMockClient();
    const artifactsClient = new APIWorkflowArtifacts(mockClient);
    const artifacts = await artifactsClient.list({
      runId: 'run_123',
      operation: 'conditional_evaluation',
      blockId: 'conditional-1',
    });

    expect(mockClient.lastFetchParams).toEqual({
      url: '/workflows/artifacts',
      method: 'GET',
      params: {
        run_id: 'run_123',
        operation: 'conditional_evaluation',
        block_id: 'conditional-1',
      },
      headers: undefined,
    });
    expect(artifacts.data).toHaveLength(1);
    expect(artifacts.data[0]?.operation).toBe('conditional_evaluation');
    expect(artifacts.list_metadata).toEqual({ before: null, after: null });
  });

  test('get() uses the public step artifact route', async () => {
    class GetMockClient extends AbstractClient {
      public lastFetchParams: Record<string, unknown> | null = null;

      protected async _fetch(params: {
        url: string;
        method: string;
        params?: Record<string, unknown>;
        headers?: Record<string, unknown>;
        body?: Record<string, unknown>;
      }): Promise<Response> {
        this.lastFetchParams = params;
        return new Response(
          JSON.stringify({
            block_id: 'extract-1',
            block_type: 'extract',
            block_label: 'Extract',
            lifecycle: { status: 'completed' },
            artifact: {
              operation: 'extraction',
              id: 'ext_123',
            },
            output: { removed: true },
            handle_outputs: {
              'output-json-0': {
                type: 'json',
                data: { invoice_number: 'INV-001' },
              },
            },
            handle_inputs: null,
          }),
          {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
          }
        );
      }
    }

    const mockClient = new GetMockClient();
    const stepsClient = new APIWorkflowRunSteps(mockClient);

    const step = await stepsClient.get('run_123', 'extract-1');

    expect(mockClient.lastFetchParams).toEqual({
      url: '/workflows/runs/run_123/steps/extract-1',
      method: 'GET',
      params: undefined,
      headers: undefined,
    });
    expect(step.block_id).toBe('extract-1');
    expect(step.artifact).toEqual({
      operation: 'extraction',
      id: 'ext_123',
    });
    expect('output' in step).toBe(false);
    expect('artifacts' in step).toBe(false);
    expect('artifact_view' in step).toBe(false);
    expect('metadata' in step).toBe(false);
    expect(step.handle_outputs?.['output-json-0']).toBeDefined();
  });

  test('does not export removed step execution response aliases', () => {
    const removedNames = [
      ['ZStep', 'Output', 'Response'].join(''),
      ['Step', 'Output', 'Response'].join(''),
      ['ZStep', 'Outputs', 'BatchResponse'].join(''),
      ['Step', 'Outputs', 'BatchResponse'].join(''),
      'ZStepExecutionsBatchResponse',
      'StepExecutionsBatchResponse',
      'ZTerminalState',
      'TerminalState',
    ];
    for (const removedName of removedNames) {
      expect(Object.prototype.hasOwnProperty.call(workflowTypes, removedName)).toBe(false);
    }
  });

  test('list() uses the full steps route and returns paginated envelope', async () => {
    const mockClient = new MockClient();
    const stepsClient = new APIWorkflowRunSteps(mockClient);

    const steps = await stepsClient.list('run_123');

    expect(mockClient.lastFetchParams).toEqual({
      url: '/workflows/runs/run_123/steps',
      method: 'GET',
      params: undefined,
      headers: undefined,
    });
    expect(steps.list_metadata).toEqual({ before: null, after: null });
    expect(steps.data).toHaveLength(1);
    expect(steps.data[0]?.block_id).toBe('extract-1');
    expect(steps.data[0]?.lifecycle.status).toBe('completed');
    expect(steps.data[0]?.artifact).toEqual({
      operation: 'extraction',
      id: 'ext_123',
    });
    expect(steps.data[0] && 'output' in steps.data[0]).toBe(false);
    expect(steps.data[0] && 'artifacts' in steps.data[0]).toBe(false);
    expect(steps.data[0] && 'artifact_view' in steps.data[0]).toBe(false);
    expect(steps.data[0] && 'metadata' in steps.data[0]).toBe(false);
    expect(steps.data[0] && 'input_document' in steps.data[0]).toBe(false);
    expect(steps.data[0] && 'output_document' in steps.data[0]).toBe(false);
    expect(steps.data[0] && 'split_documents' in steps.data[0]).toBe(false);
  });

  test('get() is the single-step execution fetch', async () => {
    const stepsClient = new APIWorkflowRunSteps(new MockClient());

    expect(stepsClient.get.length).toBe(3);
    await expect(
      (stepsClient.get as unknown as (runId: string) => Promise<unknown>)('run_123')
    ).rejects.toThrow('blockId is required');
  });

  test('only exposes get for full execution fetches', () => {
    const stepsClient = new APIWorkflowRunSteps(new MockClient());
    expect('getAll' in stepsClient).toBe(false);
    expect('get_all' in stepsClient).toBe(false);
    expect('getMany' in stepsClient).toBe(false);
    expect('get_many' in stepsClient).toBe(false);
  });

  test('get() accepts partition artifacts on for_each steps', async () => {
    class GetPartitionMockClient extends AbstractClient {
      protected async _fetch(): Promise<Response> {
        return new Response(
          JSON.stringify({
            block_id: 'for_each-1',
            block_type: 'for_each',
            block_label: 'For Each',
            lifecycle: { status: 'completed' },
            artifact: {
              operation: 'partition',
              id: 'prtn_123',
            },
            handle_outputs: null,
            handle_inputs: null,
          }),
          {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
          }
        );
      }
    }

    const stepsClient = new APIWorkflowRunSteps(new GetPartitionMockClient());
    const step = await stepsClient.get('run_123', 'for_each-1');

    expect(step.artifact).toEqual({
      operation: 'partition',
      id: 'prtn_123',
    });
  });
});
