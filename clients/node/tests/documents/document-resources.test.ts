// Offline request-contract tests for the document processing sub-resources
// (extractions, parses, classifications, splits, partitions, edits).
//
// Mirrors the Python tests/test_processing_resources.py surface but pins the
// *wire contract* instead of live behavior: a captured-fetch is injected into
// `new Retab({ apiKey, fetch })` and we assert the HTTP method, path, and the
// exact snake_cased request body — including that `coerceMimeData` turns a
// data-URI string / Buffer document into a `{ filename, url }` MIMEData shape.
//
// NEVER edits src/. These are pure offline tests; no server is contacted.

import { describe, expect, test } from 'bun:test';

import { Retab } from '../../src/index.js';

type Captured = {
  method: string;
  url: URL;
  path: string;
  query: Record<string, string>;
  body: Record<string, unknown> | undefined;
};

/**
 * Build a Retab client whose fetch records the outgoing request and returns a
 * canned JSON response so the resource method's deserializer succeeds.
 */
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
      calls.push({
        method: String(init?.method),
        url,
        path: url.pathname,
        query,
        body,
      });
      return new Response(JSON.stringify(response), {
        status: 200,
        headers: { 'content-type': 'application/json' },
      });
    }) as typeof fetch,
  });
  return { client, calls };
}

// A small text data URI: "Invoice".
const TEXT_DATA_URI = 'data:text/plain;base64,SW52b2ljZQ==';

// A real PDF magic-byte buffer ("%PDF-1.4\n") so coerceMimeData sniffs it as
// application/pdf and produces a data: URL.
const PDF_BUFFER = Buffer.from('%PDF-1.4\n', 'utf8');

// Every processing-resource create response embeds the originating file
// reference; the deserializers read it unconditionally, so canned create
// responses must include it for deserialization to succeed.
const FILE_REF = { id: 'file_x', filename: 'doc.pdf', mime_type: 'application/pdf' };

function lastCall(calls: Captured[]): Captured {
  expect(calls.length).toBeGreaterThan(0);
  return calls[calls.length - 1];
}

describe('document resources — create wire contract', () => {
  test('extractions.create snake_cases every param and coerces the document', async () => {
    const { client, calls } = clientCapturing({
      id: 'extr_1',
      file: FILE_REF,
      output: {},
    });

    await client.extractions.create(
      TEXT_DATA_URI,
      { type: 'object', properties: { invoice_id: { type: 'string' } } },
      'retab-micro',
      96,
      'be precise',
      3,
      { tenant: 'acme' },
      [{ role: 'system', content: 'ctx' }],
      true,
      false,
      false,
      { invoice_id: 'group' }
    );

    const call = lastCall(calls);
    expect(call.method).toBe('POST');
    expect(call.path).toBe('/v1/extractions');
    expect(call.body).toEqual({
      document: { filename: 'uploaded_file.plain', url: TEXT_DATA_URI },
      json_schema: { type: 'object', properties: { invoice_id: { type: 'string' } } },
      model: 'retab-micro',
      image_resolution_dpi: 96,
      instructions: 'be precise',
      n_consensus: 3,
      metadata: { tenant: 'acme' },
      additional_messages: [{ role: 'system', content: 'ctx' }],
      bust_cache: true,
      stream: false,
      background: false,
      chunking_keys: { invoice_id: 'group' },
    });
  });

  test('extractions.create coerces a Buffer document to a base64 data: URL MIMEData', async () => {
    const { client, calls } = clientCapturing({ id: 'extr_2', file: FILE_REF, output: {} });

    await client.extractions.create(PDF_BUFFER, { type: 'object' });

    const call = lastCall(calls);
    const document = (call.body as Record<string, unknown>).document as {
      filename: string;
      url: string;
    };
    expect(document.filename).toBe('uploaded_file.pdf');
    expect(document.url).toBe(`data:application/pdf;base64,${PDF_BUFFER.toString('base64')}`);
    // Optional params the caller omitted must serialize as null (JSON drops undefined).
    expect((call.body as Record<string, unknown>).json_schema).toEqual({ type: 'object' });
  });

  test('parses.create maps tableParsingFormat/imageResolutionDpi/bustCache', async () => {
    const { client, calls } = clientCapturing({
      id: 'parse_1',
      file: FILE_REF,
      output: { text: 'hi' },
    });

    await client.parses.create(TEXT_DATA_URI, 'retab-micro', 'html', 96, 'go', true, false);

    const call = lastCall(calls);
    expect(call.method).toBe('POST');
    expect(call.path).toBe('/v1/parses');
    expect(call.body).toEqual({
      document: { filename: 'uploaded_file.plain', url: TEXT_DATA_URI },
      model: 'retab-micro',
      table_parsing_format: 'html',
      image_resolution_dpi: 96,
      instructions: 'go',
      bust_cache: true,
      background: false,
    });
  });

  test('classifications.create maps firstNPages/nConsensus and serializes categories', async () => {
    const { client, calls } = clientCapturing({ id: 'cls_1', file: FILE_REF, categories: [] });

    await client.classifications.create(
      TEXT_DATA_URI,
      [
        { name: 'invoice', description: 'a billing invoice' },
        { name: 'receipt', description: 'a payment receipt' },
      ],
      'retab-micro',
      5,
      'classify it',
      2,
      true,
      false
    );

    const call = lastCall(calls);
    expect(call.method).toBe('POST');
    expect(call.path).toBe('/v1/classifications');
    expect(call.body).toEqual({
      document: { filename: 'uploaded_file.plain', url: TEXT_DATA_URI },
      categories: [
        { name: 'invoice', description: 'a billing invoice' },
        { name: 'receipt', description: 'a payment receipt' },
      ],
      model: 'retab-micro',
      first_n_pages: 5,
      instructions: 'classify it',
      n_consensus: 2,
      bust_cache: true,
      background: false,
    });
  });

  test('splits.create maps nConsensus and serializes subdocuments', async () => {
    const { client, calls } = clientCapturing({ id: 'split_1', file: FILE_REF, subdocuments: [] });

    await client.splits.create(
      PDF_BUFFER,
      [
        { name: 'header', description: 'header sections' },
        { name: 'body', description: 'the body' },
      ],
      'retab-micro',
      'split it',
      2,
      false,
      false
    );

    const call = lastCall(calls);
    expect(call.method).toBe('POST');
    expect(call.path).toBe('/v1/splits');
    const body = call.body as Record<string, unknown>;
    expect(body.subdocuments).toEqual([
      { name: 'header', description: 'header sections' },
      { name: 'body', description: 'the body' },
    ]);
    expect(body.model).toBe('retab-micro');
    expect(body.instructions).toBe('split it');
    expect(body.n_consensus).toBe(2);
    expect(body.bust_cache).toBe(false);
    expect(body.background).toBe(false);
    expect((body.document as { filename: string }).filename).toBe('uploaded_file.pdf');
  });

  test('partitions.create maps allowOverlap/nConsensus/bustCache', async () => {
    const { client, calls } = clientCapturing({ id: 'part_1', file: FILE_REF });

    await client.partitions.create(
      TEXT_DATA_URI,
      'line_items',
      'group by line item',
      'retab-micro',
      3,
      true,
      true,
      false
    );

    const call = lastCall(calls);
    expect(call.method).toBe('POST');
    expect(call.path).toBe('/v1/partitions');
    expect(call.body).toEqual({
      document: { filename: 'uploaded_file.plain', url: TEXT_DATA_URI },
      key: 'line_items',
      instructions: 'group by line item',
      model: 'retab-micro',
      n_consensus: 3,
      allow_overlap: true,
      bust_cache: true,
      background: false,
    });
  });

  test('edits.create maps templateId/bustCache and coerces an optional document', async () => {
    const { client, calls } = clientCapturing({
      id: 'edit_1',
      file: FILE_REF,
      config: {},
    });

    await client.edits.create(
      'fill the form',
      PDF_BUFFER,
      'tmpl_42',
      'retab-micro',
      undefined,
      true,
      false
    );

    const call = lastCall(calls);
    expect(call.method).toBe('POST');
    expect(call.path).toBe('/v1/edits');
    const body = call.body as Record<string, unknown>;
    expect(body.instructions).toBe('fill the form');
    expect(body.template_id).toBe('tmpl_42');
    expect(body.model).toBe('retab-micro');
    expect(body.bust_cache).toBe(true);
    expect(body.background).toBe(false);
    expect((body.document as { filename: string }).filename).toBe('uploaded_file.pdf');
  });

  test('edits.create omits the document entirely when none is provided', async () => {
    const { client, calls } = clientCapturing({
      id: 'edit_2',
      file: FILE_REF,
      config: {},
    });

    await client.edits.create('fill from template', undefined, 'tmpl_99');

    const call = lastCall(calls);
    const body = call.body as Record<string, unknown>;
    // undefined document is dropped by JSON.stringify — must not appear on the wire.
    expect('document' in body).toBe(false);
    expect(body.template_id).toBe('tmpl_99');
  });
});

describe('document resources — list query mapping', () => {
  const envelope = { data: [], list_metadata: { before: null, after: null } };

  test('extractions.list snake_cases filters and paging', async () => {
    const { client, calls } = clientCapturing(envelope);

    await client.extractions.list({
      filename: 'inv.pdf',
      filenameRegex: 'inv.*',
      filenameContains: 'inv',
      documentType: ['invoice', 'receipt'],
      status: 'completed',
      fromDate: '2026-01-01',
      toDate: '2026-02-01',
      metadata: '{"k":"v"}',
      limit: 25,
      before: 'b1',
      after: 'a1',
      order: 'asc',
    });

    const call = lastCall(calls);
    expect(call.method).toBe('GET');
    expect(call.path).toBe('/v1/extractions');
    // Array params repeat the key (?document_type=invoice&document_type=receipt).
    expect(call.url.searchParams.getAll('document_type')).toEqual(['invoice', 'receipt']);
    expect(call.query).toMatchObject({
      filename: 'inv.pdf',
      filename_regex: 'inv.*',
      filename_contains: 'inv',
      status: 'completed',
      from_date: '2026-01-01',
      to_date: '2026-02-01',
      metadata: '{"k":"v"}',
      limit: '25',
      before: 'b1',
      after: 'a1',
      order: 'asc',
    });
  });

  test('parses.list maps fromDate/toDate', async () => {
    const { client, calls } = clientCapturing(envelope);
    await client.parses.list({ filename: 'a.pdf', fromDate: '2026-01-01', toDate: '2026-02-01' });
    const call = lastCall(calls);
    expect(call.path).toBe('/v1/parses');
    expect(call.query).toMatchObject({
      filename: 'a.pdf',
      from_date: '2026-01-01',
      to_date: '2026-02-01',
    });
  });

  test('edits.list maps templateId -> template_id', async () => {
    const { client, calls } = clientCapturing(envelope);
    await client.edits.list({ templateId: 'tmpl_7', filename: 'form.pdf' });
    const call = lastCall(calls);
    expect(call.path).toBe('/v1/edits');
    expect(call.query).toMatchObject({ template_id: 'tmpl_7', filename: 'form.pdf' });
  });

  test('classifications/splits/partitions list map status + dates to the right paths', async () => {
    const checks: Array<[() => Promise<unknown>, string]> = [];
    const c1 = clientCapturing(envelope);
    checks.push([
      () => c1.client.classifications.list({ status: 'completed', fromDate: '2026-01-01' }),
      '/v1/classifications',
    ]);
    const c2 = clientCapturing(envelope);
    checks.push([
      () => c2.client.splits.list({ status: 'completed', toDate: '2026-02-01' }),
      '/v1/splits',
    ]);
    const c3 = clientCapturing(envelope);
    checks.push([
      () => c3.client.partitions.list({ status: 'completed', fromDate: '2026-01-01' }),
      '/v1/partitions',
    ]);

    await checks[0][0]();
    expect(lastCall(c1.calls).path).toBe('/v1/classifications');
    expect(lastCall(c1.calls).query).toMatchObject({
      status: 'completed',
      from_date: '2026-01-01',
    });

    await checks[1][0]();
    expect(lastCall(c2.calls).path).toBe('/v1/splits');
    expect(lastCall(c2.calls).query).toMatchObject({ status: 'completed', to_date: '2026-02-01' });

    await checks[2][0]();
    expect(lastCall(c3.calls).path).toBe('/v1/partitions');
    expect(lastCall(c3.calls).query).toMatchObject({
      status: 'completed',
      from_date: '2026-01-01',
    });
  });
});
