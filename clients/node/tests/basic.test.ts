import { describe, expect, test } from 'bun:test';

import { Retab, ZStepExecutionResponse } from '../src';

describe('Node SDK smoke coverage', () => {
  test('exports workflow artifact execution schemas', () => {
    const step = ZStepExecutionResponse.parse({
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
    });

    expect(step.artifact?.id).toBe('ext_123');
    expect(step.artifact?.operation).toBe('extraction');
  });

  test('does not expose projects or evals on the public client', () => {
    const client = new Retab({ apiKey: 'test_key' });
    const publicClient = client as unknown as Record<string, unknown>;

    expect(publicClient.projects).toBeUndefined();
    expect(publicClient.evals).toBeUndefined();
  });
});
