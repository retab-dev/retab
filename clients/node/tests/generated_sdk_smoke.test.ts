import { describe, expect, test } from 'bun:test';

import { PaginatedList, Retab, ZWorkflowBlockCreateRequestType } from '../src/index.js';

describe('generated SDK smoke coverage', () => {
  test('constructs the current workflow client surface', () => {
    const client = new Retab({ apiKey: 'test_key' });

    expect(client.workflows).toBeDefined();
    expect(client.workflows.runs).toBeDefined();
    expect(client.workflows.steps).toBeDefined();
    expect(client.workflows.artifacts).toBeDefined();
    expect(client.workflows.spec).toBeDefined();
    expect('specs' in client.workflows).toBe(false);
  });

  test('exports generated workflow schemas', () => {
    expect(ZWorkflowBlockCreateRequestType.parse('start_document')).toBe('start_document');
    expect(ZWorkflowBlockCreateRequestType.parse('extract')).toBe('extract');
  });

  test('pagination helper iterates the current page', async () => {
    const page = new PaginatedList({
      data: ['one', 'two'],
      list_metadata: { before: null, after: null },
    });

    const items: string[] = [];
    for await (const item of page) {
      items.push(item);
    }

    expect(items).toEqual(['one', 'two']);
    expect(page.listMetadata).toBe(page.list_metadata);
    expect(page.hasMore()).toBe(false);
  });

  test('workflow runs create coerces document maps per block id', async () => {
    let capturedBody: unknown;
    const fileRef = {
      id: 'file_existing',
      filename: 'stored.pdf',
      mimeType: 'application/pdf',
    };
    const client = new Retab({
      apiKey: 'test_key',
      fetch: async (_input, init) => {
        capturedBody = JSON.parse(String(init?.body));
        return new Response(
          JSON.stringify({
            id: 'run_1',
            workflow: {
              workflow_id: 'wf_1',
              version_id: 'ver_1',
              name_at_run_time: 'Workflow',
            },
            trigger: { type: 'manual' },
          }),
          { status: 200, headers: { 'content-type': 'application/json' } }
        );
      },
    });

    await client.workflows.runs.create(
      'wf_1',
      {
        start_document_1: 'https://example.com/remote.pdf?download=1',
        start_document_2: fileRef,
      },
      {
        start_json_1: {
          invoice_id: 'inv_sdk_smoke',
          line_items: [
            { description: 'warehouse handling', amount: 120 },
            { description: 'local delivery', amount: 80 },
          ],
          tax_rate: 0.2,
          stated_total: 240,
        },
      },
      'draft'
    );

    expect(capturedBody).toEqual({
      workflow_id: 'wf_1',
      documents: {
        start_document_1: {
          filename: 'remote.pdf',
          url: 'https://example.com/remote.pdf?download=1',
        },
        start_document_2: fileRef,
      },
      json_inputs: {
        start_json_1: {
          invoice_id: 'inv_sdk_smoke',
          line_items: [
            { description: 'warehouse handling', amount: 120 },
            { description: 'local delivery', amount: 80 },
          ],
          tax_rate: 0.2,
          stated_total: 240,
        },
      },
      version: 'draft',
    });
  });
});
