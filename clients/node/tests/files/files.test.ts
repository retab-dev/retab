// Offline request-contract tests for the files resource.
//
// Pins the wire contract of files.list query mapping, the upload/complete
// lifecycle, and the blueprint create/get/cancel paths using a captured-fetch
// injected into `new Retab({ apiKey, fetch })`. No server is contacted.
//
// NEVER edits src/.

import { describe, expect, test } from 'bun:test';

import { Retab } from '../../src/index.js';

type Captured = {
  method: string;
  url: URL;
  path: string;
  query: Record<string, string>;
  body: Record<string, unknown> | undefined;
};

function clientCapturing(response: unknown): { client: Retab; calls: Captured[] } {
  const calls: Captured[] = [];
  const client = new Retab({
    apiKey: 'test_key',
    fetch: (async (input: URL | RequestInfo, init?: RequestInit) => {
      const url = new URL(String(input));
      const query: Record<string, string> = {};
      for (const [k, v] of url.searchParams.entries()) query[k] = v;
      const rawBody = init?.body;
      const body =
        rawBody === undefined || rawBody === null
          ? undefined
          : (JSON.parse(String(rawBody)) as Record<string, unknown>);
      calls.push({ method: String(init?.method), url, path: url.pathname, query, body });
      return new Response(JSON.stringify(response), {
        status: 200,
        headers: { 'content-type': 'application/json' },
      });
    }) as typeof fetch,
  });
  return { client, calls };
}

function lastCall(calls: Captured[]): Captured {
  expect(calls.length).toBeGreaterThan(0);
  return calls[calls.length - 1];
}

describe('files.list query mapping', () => {
  const envelope = { data: [], list_metadata: { before: null, after: null } };

  test('snake_cases every list filter and paging param', async () => {
    const { client, calls } = clientCapturing(envelope);

    await client.files.list({
      filename: 'invoice.pdf',
      mimeType: 'application/pdf',
      fromDate: '2026-01-01',
      toDate: '2026-02-01',
      includeEmbeddings: true,
      sortBy: 'created_at',
      limit: 50,
      before: 'cur_before',
      after: 'cur_after',
      order: 'desc',
    });

    const call = lastCall(calls);
    expect(call.method).toBe('GET');
    expect(call.path).toBe('/v1/files');
    expect(call.query).toEqual({
      filename: 'invoice.pdf',
      mime_type: 'application/pdf',
      from_date: '2026-01-01',
      to_date: '2026-02-01',
      include_embeddings: 'true',
      sort_by: 'created_at',
      limit: '50',
      before: 'cur_before',
      after: 'cur_after',
      order: 'desc',
    });
  });

  test('omits unset filters (undefined/null are dropped from the query string)', async () => {
    const { client, calls } = clientCapturing(envelope);

    await client.files.list({ filename: 'only.pdf', limit: 10 });

    const call = lastCall(calls);
    expect(call.query).toEqual({ filename: 'only.pdf', limit: '10' });
    // The other knobs must not leak onto the wire when unset.
    expect('mime_type' in call.query).toBe(false);
    expect('include_embeddings' in call.query).toBe(false);
    expect('sort_by' in call.query).toBe(false);
    expect('from_date' in call.query).toBe(false);
  });
});

describe('files upload lifecycle', () => {
  test('create_upload posts filename/content_type/size_bytes/sha256', async () => {
    const { client, calls } = clientCapturing({
      file_id: 'file_1',
      upload_url: 'https://upload.example/x',
      upload_method: 'PUT',
      upload_headers: {},
      mime_data: { filename: 'invoice.pdf', url: 'https://storage.retab.com/org/file_1.pdf' },
      expires_at: '2026-01-01T00:00:00Z',
    });

    await client.files.create_upload('invoice.pdf', 12345, 'application/pdf', 'abc123');

    const call = lastCall(calls);
    expect(call.method).toBe('POST');
    expect(call.path).toBe('/v1/files/upload');
    expect(call.body).toEqual({
      filename: 'invoice.pdf',
      content_type: 'application/pdf',
      size_bytes: 12345,
      sha256: 'abc123',
    });
  });

  test('complete_upload targets /v1/files/upload/{fileId}/complete with sha256 body', async () => {
    const { client, calls } = clientCapturing({
      filename: 'invoice.pdf',
      url: 'https://storage.retab.com/org/file_1.pdf',
    });

    await client.files.complete_upload('file_1', 'abc123');

    const call = lastCall(calls);
    expect(call.method).toBe('POST');
    expect(call.path).toBe('/v1/files/upload/file_1/complete');
    expect(call.body).toEqual({ sha256: 'abc123' });
  });
});

describe('files blueprint paths', () => {
  test('create_blueprint posts file_id/intent/background to /v1/files/blueprints', async () => {
    const { client, calls } = clientCapturing({
      id: 'bp_1',
      file: { id: 'file_1', filename: 'a.pdf', mime_type: 'application/pdf' },
    });

    await client.files.create_blueprint('file_1', 'extract invoices', false);

    const call = lastCall(calls);
    expect(call.method).toBe('POST');
    expect(call.path).toBe('/v1/files/blueprints');
    expect(call.body).toEqual({
      file_id: 'file_1',
      intent: 'extract invoices',
      background: false,
    });
  });

  test('get_blueprint targets /v1/files/blueprints/{id} with include_output query', async () => {
    const { client, calls } = clientCapturing({
      id: 'bp_1',
      file: { id: 'file_1', filename: 'a.pdf', mime_type: 'application/pdf' },
    });

    await client.files.get_blueprint('bp_1', { includeOutput: true });

    const call = lastCall(calls);
    expect(call.method).toBe('GET');
    expect(call.path).toBe('/v1/files/blueprints/bp_1');
    expect(call.query).toEqual({ include_output: 'true' });
  });

  test('create_blueprint_cancel targets /v1/files/blueprints/{id}/cancel', async () => {
    const { client, calls } = clientCapturing({
      id: 'bp_1',
      file: { id: 'file_1', filename: 'a.pdf', mime_type: 'application/pdf' },
    });

    await client.files.create_blueprint_cancel('bp_1');

    const call = lastCall(calls);
    expect(call.method).toBe('POST');
    expect(call.path).toBe('/v1/files/blueprints/bp_1/cancel');
    expect(call.body).toBeUndefined();
  });

  test('get and get_download_link target the right per-file paths', async () => {
    const c1 = clientCapturing({ id: 'file_1', filename: 'a.pdf' });
    await c1.client.files.get('file_1');
    const getCall = lastCall(c1.calls);
    expect(getCall.method).toBe('GET');
    expect(getCall.path).toBe('/v1/files/file_1');

    const c2 = clientCapturing({
      url: 'https://signed.example/x',
      expires_at: '2026-01-01T00:00:00Z',
    });
    await c2.client.files.get_download_link('file_1');
    const linkCall = lastCall(c2.calls);
    expect(linkCall.method).toBe('GET');
    expect(linkCall.path).toBe('/v1/files/file_1/download-link');
  });
});
