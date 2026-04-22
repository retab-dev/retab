import { describe, expect, test } from 'bun:test';

import { AbstractClient } from '../src/client';
import APIWorkflowRunSteps from '../src/api/workflows/runs/steps/client';
import { ZExtractStepOutput, ZForEachSentinelStartStepOutput } from '../src/types';

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
      JSON.stringify([
        {
          run_id: 'run_123',
          organization_id: 'org_123',
          block_id: 'extract-1',
          step_id: 'extract-1',
          block_type: 'extract',
          block_label: 'Extract',
          status: 'completed',
          artifact: {
            operation: 'extraction',
            id: 'ext_123',
          },
        },
      ]),
      {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }
    );
  }
}

describe('workflow run steps client', () => {
  test('get() uses the public step route without raw output', async () => {
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
            status: 'completed',
            artifact: {
              operation: 'extraction',
              id: 'ext_123',
            },
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
    expect(step.handle_outputs?.['output-json-0']).toBeDefined();
  });

  test('list() uses the full steps route', async () => {
    const mockClient = new MockClient();
    const stepsClient = new APIWorkflowRunSteps(mockClient);

    const steps = await stepsClient.list('run_123');

    expect(mockClient.lastFetchParams).toEqual({
      url: '/workflows/runs/run_123/steps',
      method: 'GET',
      params: undefined,
      headers: undefined,
    });
    expect(steps).toHaveLength(1);
    expect(steps[0]?.block_id).toBe('extract-1');
    expect(steps[0]?.status).toBe('completed');
    expect(steps[0]?.artifact).toEqual({
      operation: 'extraction',
      id: 'ext_123',
    });
    expect(steps[0] && 'output' in steps[0]).toBe(false);
    expect(steps[0] && 'input_document' in steps[0]).toBe(false);
    expect(steps[0] && 'output_document' in steps[0]).toBe(false);
    expect(steps[0] && 'split_documents' in steps[0]).toBe(false);
  });

  test('getMany() sends POST to /steps/batch', async () => {
    class BatchMockClient extends AbstractClient {
      public lastFetchParams: Record<string, unknown> | null = null;

      protected async _fetch(params: {
        url: string;
        method: string;
        params?: Record<string, unknown>;
        headers?: Record<string, unknown>;
        body?: unknown;
      }): Promise<Response> {
        this.lastFetchParams = params;
        return new Response(
          JSON.stringify({
            outputs: {
              'extract-1': {
                block_id: 'extract-1',
                block_type: 'extract',
                block_label: 'Extract',
                status: 'completed',
                artifact: {
                  operation: 'extraction',
                  id: 'ext_456',
                },
                handle_outputs: {
                  'output-json-0': { type: 'json', data: { field: 'value' } },
                },
                handle_inputs: null,
              },
            },
          }),
          {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
          }
        );
      }
    }

    const mockClient = new BatchMockClient();
    const stepsClient = new APIWorkflowRunSteps(mockClient);

    const batch = await stepsClient.getMany('run_123', ['extract-1']);

    expect(mockClient.lastFetchParams?.url).toBe('/workflows/runs/run_123/steps/batch');
    expect(mockClient.lastFetchParams?.method).toBe('POST');
    expect(batch.outputs['extract-1']?.block_id).toBe('extract-1');
    expect(batch.outputs['extract-1']?.artifact).toEqual({
      operation: 'extraction',
      id: 'ext_456',
    });
  });

  test('get() accepts partition artifacts on for_each steps', async () => {
    class GetPartitionMockClient extends AbstractClient {
      protected async _fetch(): Promise<Response> {
        return new Response(
          JSON.stringify({
            block_id: 'for_each-1',
            block_type: 'for_each',
            block_label: 'For Each',
            status: 'completed',
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

  test('extract step output accepts canonical shape', () => {
    const parsed = ZExtractStepOutput.parse({
      output: { invoice_number: 'INV-001' },
      consensus: {
        choices: [{ invoice_number: 'INV-001' }],
        likelihoods: { invoice_number: 0.98 },
      },
      extraction_id: 'ext_123',
    });

    expect(parsed.output).toEqual({ invoice_number: 'INV-001' });
    expect(parsed.consensus).toEqual({
      choices: [{ invoice_number: 'INV-001' }],
      likelihoods: { invoice_number: 0.98 },
    });
  });

  test('extract step output normalizes legacy shape', () => {
    const parsed = ZExtractStepOutput.parse({
      extracted_data: { invoice_number: 'INV-002' },
      likelihoods: { invoice_number: 0.91 },
      consensus_details: [{ data: { invoice_number: 'INV-002' } }],
      extraction_id: 'ext_456',
    });

    expect(parsed.output).toEqual({ invoice_number: 'INV-002' });
    expect(parsed.consensus).toEqual({
      choices: [{ invoice_number: 'INV-002' }],
      likelihoods: { invoice_number: 0.91 },
    });
  });

  test('for_each sentinel start output accepts partition payload', () => {
    const parsed = ZForEachSentinelStartStepOutput.parse({
      message: 'Splitting document',
      mr_id: 'for_each-1',
      current_index: 0,
      total_items: 1,
      max_iterations: 1,
      is_first_iteration: true,
      map_method: 'split_by_key',
      partition_id: 'prtn_123',
      all_item_keys: ['invoice_1'],
    });

    expect(parsed.partition_id).toBe('prtn_123');
  });
});
