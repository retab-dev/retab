import { describe, expect, test } from 'bun:test';

import { Retab, ZStepExecutionResponse, ZStepExecutionsBatchResponse } from '../src';

describe('Node SDK smoke coverage', () => {
  test('exports workflow artifact execution schemas', () => {
    const step = ZStepExecutionResponse.parse({
      block_id: 'extract-1',
      block_type: 'extract',
      block_label: 'Extract',
      status: 'completed',
      artifacts: [
        {
          operation: 'extraction',
          id: 'ext_123',
        },
      ],
      artifact_view: {
        block_type: 'extract',
        artifacts: [
          {
            operation: 'extraction',
            id: 'ext_123',
          },
        ],
        data: {
          output: { invoice_number: 'INV-001' },
        },
      },
      handle_outputs: {
        'output-json-0': {
          type: 'json',
          data: { invoice_number: 'INV-001' },
        },
      },
    });

    expect(step.artifacts[0]?.id).toBe('ext_123');
    expect(step.artifact_view?.data?.output.invoice_number).toBe('INV-001');

    const batch = ZStepExecutionsBatchResponse.parse({
      executions: {
        'extract-1': step,
      },
    });

    expect(batch.executions['extract-1']?.artifacts[0]?.operation).toBe('extraction');
  });

  test('exposes project iteration methods on the public client', () => {
    const client = new Retab({ apiKey: 'test_key' });
    const iterations = client.projects.datasets.iterations as Record<string, unknown>;
    const expectedMethods = [
      'create',
      'get',
      'list',
      'updateDraft',
      'delete',
      'finalize',
      'getSchema',
      'processDocuments',
      'getDocument',
      'listDocuments',
      'updateDocument',
      'deleteDocument',
      'getMetrics',
    ];

    for (const methodName of expectedMethods) {
      expect(typeof iterations[methodName]).toBe('function');
    }
  });
});
