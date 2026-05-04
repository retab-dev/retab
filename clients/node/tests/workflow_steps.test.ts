import { describe, expect, test } from 'bun:test';

import { AbstractClient } from '../src/client';
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
      JSON.stringify([
        {
          run_id: 'run_123',
          organization_id: 'org_123',
          block_id: 'extract-1',
          step_id: 'extract-1',
          block_type: 'extract',
          block_label: 'Extract',
          status: 'completed',
          execution: {
            inputs: {},
            outputs: {},
            artifacts: [
              {
                operation: 'extraction',
                id: 'ext_123',
              },
            ],
            metadata: {},
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
            status: 'completed',
            execution: {
              artifacts: [
                {
                  operation: 'extraction',
                  id: 'ext_123',
                },
              ],
              outputs: {
                'output-json-0': {
                  type: 'json',
                  data: { invoice_number: 'INV-001' },
                },
              },
              inputs: {},
              metadata: {},
            },
            output: { removed: true },
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
    expect(step.execution.artifacts).toEqual([
      {
        operation: 'extraction',
        id: 'ext_123',
      },
    ]);
    expect('output' in step).toBe(false);
    expect(step.execution.outputs['output-json-0']).toBeDefined();
  });

  test('does not export removed step execution response aliases', () => {
    const removedNames = [
      ['ZStep', 'Output', 'Response'].join(''),
      ['Step', 'Output', 'Response'].join(''),
      ['ZStep', 'Outputs', 'BatchResponse'].join(''),
      ['Step', 'Outputs', 'BatchResponse'].join(''),
    ];
    for (const removedName of removedNames) {
      expect(Object.prototype.hasOwnProperty.call(workflowTypes, removedName)).toBe(false);
    }
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
    expect(steps[0]?.execution.artifacts).toContainEqual({
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
            executions: {
              'extract-1': {
                block_id: 'extract-1',
                block_type: 'extract',
                block_label: 'Extract',
                status: 'completed',
                execution: {
                  artifacts: [
                    {
                      operation: 'extraction',
                      id: 'ext_456',
                    },
                  ],
                  outputs: {
                    'output-json-0': { type: 'json', data: { field: 'value' } },
                  },
                  inputs: {},
                  metadata: {},
                },
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
    expect(batch.executions['extract-1']?.block_id).toBe('extract-1');
    expect(batch.executions['extract-1']?.execution.artifacts).toContainEqual({
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
            execution: {
              inputs: {},
              outputs: {},
              artifacts: [
                {
                  operation: 'partition',
                  id: 'prtn_123',
                },
              ],
              metadata: {},
            },
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

    expect(step.execution.artifacts).toContainEqual({
      operation: 'partition',
      id: 'prtn_123',
    });
  });

});
