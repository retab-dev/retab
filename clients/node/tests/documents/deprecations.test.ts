import { describe, expect, test } from 'bun:test';

import APIV1 from '../../src/api/client.js';
import { AbstractClient } from '../../src/client.js';

type RecordedRequest = {
  url: string;
  method: string;
  params?: Record<string, unknown>;
  headers?: Record<string, unknown>;
  bodyMime?: 'application/json' | 'multipart/form-data';
  body?: Record<string, unknown>;
};

class MockClient extends AbstractClient {
  public requests: RecordedRequest[] = [];

  protected async _fetch(params: RecordedRequest): Promise<Response> {
    this.requests.push(params);

    if (params.method === 'DELETE') {
      return new Response(null, { status: 204 });
    }

    switch (params.url) {
      case '/documents/extract':
        return new Response(
          JSON.stringify({
            id: 'chatcmpl_123',
            object: 'chat.completion',
            created: 1710000000,
            model: 'retab-small',
            choices: [
              {
                index: 0,
                finish_reason: 'stop',
                message: {
                  role: 'assistant',
                  content: '{"invoice_number":"INV-001"}',
                  parsed: { invoice_number: 'INV-001' },
                },
              },
            ],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } }
        );
      case '/documents/extractions':
        return new Response(
          JSON.stringify({
            id: 'chunk_123',
            object: 'chat.completion.chunk',
            created: 1710000000,
            model: 'retab-small',
            choices: [
              {
                index: 0,
                delta: {
                  content: '{}',
                  is_valid_json: true,
                  flat_parsed: {},
                  flat_likelihoods: {},
                  flat_deleted_keys: [],
                },
              },
            ],
          }),
          { status: 200, headers: { 'Content-Type': 'application/stream+json' } }
        );
      case '/documents/split':
        return new Response(
          JSON.stringify({
            splits: [{ name: 'invoice', pages: [1], partitions: [] }],
            consensus: { choices: [] },
            usage: { page_count: 1, credits: 1.0 },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } }
        );
      case '/documents/classify':
        return new Response(
          JSON.stringify({
            classification: { category: 'invoice', reasoning: 'Detected invoice markers.' },
            consensus: { choices: [], likelihood: 1.0 },
            usage: { page_count: 1, credits: 1.0 },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } }
        );
      default:
        return new Response(
          JSON.stringify({
            id: 'extr_123',
            file: {
              id: 'file_123',
              filename: 'invoice.txt',
              mime_type: 'text/plain',
            },
            model: 'retab-small',
            json_schema: { type: 'object', properties: {} },
            n_consensus: 1,
            image_resolution_dpi: 192,
            output: { invoice_number: 'INV-001' },
            consensus: { choices: [], likelihoods: null },
            metadata: {},
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } }
        );
    }
  }
}

const SAMPLE_DOCUMENT = {
  filename: 'invoice.txt',
  url: 'data:text/plain;base64,aW52b2ljZQ==',
  mime_type: 'text/plain',
};

function withWarnCapture(run: () => Promise<void>): Promise<string[]> {
  const warnings: string[] = [];
  const originalWarn = console.warn;
  console.warn = (...args: unknown[]) => {
    warnings.push(args.join(' '));
  };
  return run()
    .finally(() => {
      console.warn = originalWarn;
    })
    .then(() => warnings);
}

describe('Node SDK legacy document deprecations', () => {
  test('documents.extract warns and keeps the compatibility route', async () => {
    const client = new APIV1(new MockClient());
    const warnings = await withWarnCapture(async () => {
      await client.documents.extract({
        document: SAMPLE_DOCUMENT,
        model: 'retab-small',
        json_schema: { type: 'object', properties: { invoice_number: { type: 'string' } } },
      });
    });

    expect(warnings).toContain(
      'client.documents.extract is deprecated; use client.extractions.create'
    );
  });

  test('documents.extract_stream warns and keeps the compatibility route', async () => {
    const mockClient = new MockClient();
    const client = new APIV1(mockClient);
    const warnings = await withWarnCapture(async () => {
      await client.documents.extract_stream({
        document: SAMPLE_DOCUMENT,
        model: 'retab-small',
        json_schema: { type: 'object', properties: { invoice_number: { type: 'string' } } },
      });
    });

    expect(warnings).toContain(
      'client.documents.extract_stream is deprecated; use POST /v1/extractions/stream'
    );
    expect(mockClient.requests[0]?.url).toBe('/documents/extractions');
  });

  test('documents.split warns and keeps the compatibility route', async () => {
    const warnings = await withWarnCapture(async () => {
      await new APIV1(new MockClient()).documents.split({
        document: SAMPLE_DOCUMENT,
        model: 'retab-small',
        subdocuments: [{ name: 'invoice', description: 'Invoice documents' }],
      });
    });

    expect(warnings).toContain('client.documents.split is deprecated; use client.splits.create');
  });

  test('documents.classify warns and keeps the compatibility route', async () => {
    const warnings = await withWarnCapture(async () => {
      await new APIV1(new MockClient()).documents.classify({
        document: SAMPLE_DOCUMENT,
        model: 'retab-small',
        categories: [{ name: 'invoice', description: 'Invoice documents' }],
      });
    });

    expect(warnings).toContain(
      'client.documents.classify is deprecated; use client.classifications.create'
    );
  });

  test('extractions.delete uses the resource route', async () => {
    const mockClient = new MockClient();
    const client = new APIV1(mockClient);

    await client.extractions.delete('extr_123');

    expect(mockClient.requests[0]?.url).toBe('/extractions/extr_123');
    expect(mockClient.requests[0]?.method).toBe('DELETE');
  });
});
