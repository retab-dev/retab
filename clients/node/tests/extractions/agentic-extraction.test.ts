import { describe, expect, test } from 'bun:test';

import { Retab } from '../../src/index.js';

// Wire-shape tests for the agentic_extraction extraction field.
//
// agentic_extraction asks the server to run agentic extraction on a long document
// (it implies deep extraction). Its client-side failure modes are silent: a
// dropped field, a camelCase key the API ignores, or an unset default that
// serializes as `false` each give a request that succeeds and quietly runs
// something else. So these assert the request body actually put on the wire.
//
// The generated create/create_stream methods are positional in the request
// schema's property order. agentic_extraction is the LAST property, after
// sources, so it is the 13th argument and every earlier argument keeps its
// position.

type CapturedRequest = { url: string; body: Record<string, unknown> };

// captureRequest builds a client over an injected fetch (the SDK's supported
// seam) and returns what the client actually sent.
function captureRequest(): { client: Retab; captured: CapturedRequest[] } {
  const captured: CapturedRequest[] = [];
  const client = new Retab({
    apiKey: 'test_key',
    fetch: (async (input: unknown, init?: { body?: unknown }) => {
      captured.push({
        url: String(input),
        body: init?.body ? (JSON.parse(String(init.body)) as Record<string, unknown>) : {},
      });
      return new Response(
        JSON.stringify({
          id: 'extr_test',
          file: { id: 'file_1', filename: 'statement.pdf', mime_type: 'application/pdf' },
          model: 'retab-large',
          json_schema: { type: 'object' },
          output: {},
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } }
      );
    }) as typeof fetch,
  });

  return { client, captured };
}

const DOCUMENT = { filename: 'statement.pdf', url: 'data:application/pdf;base64,QUJD' };
const JSON_SCHEMA = {
  type: 'object',
  properties: { transactions: { type: 'array', items: { type: 'object' } } },
};

type CreateArgs = Parameters<Retab['extractions']['create']>;

// createArgs returns the positional arguments for create/create_stream with
// agenticExtraction in the 13th slot (after sources). Built as an array so
// the test fails at runtime, not only at type-check time, if the slot moves.
function createArgs(agenticExtraction: boolean | undefined): CreateArgs {
  const args: unknown[] = [
    DOCUMENT, // document
    JSON_SCHEMA, // jsonSchema
    'retab-large', // model
    undefined, // instructions
    undefined, // nConsensus
    undefined, // metadata
    undefined, // additionalMessages
    undefined, // bustCache
    undefined, // stream
    undefined, // background
    undefined, // deepExtraction
    undefined, // sources
    agenticExtraction, // agenticExtraction
  ];
  return args as unknown as CreateArgs;
}

describe('agentic_extraction request shape', () => {
  test('sends agentic_extraction when enabled', async () => {
    const { client, captured } = captureRequest();

    await client.extractions.create(...createArgs(true));

    expect(captured).toHaveLength(1);
    expect(captured[0]!.url).toContain('/v1/extractions');
    expect(captured[0]!.body.agentic_extraction).toBe(true);
  });

  test('sends an explicit false so an opt-out is distinguishable from unset', async () => {
    const { client, captured } = captureRequest();

    await client.extractions.create(...createArgs(false));

    expect('agentic_extraction' in captured[0]!.body).toBe(true);
    expect(captured[0]!.body.agentic_extraction).toBe(false);
  });

  test('omits agentic_extraction when unset', async () => {
    // Absent, not false: the default request body stays identical to what it
    // was before this field existed.
    const { client, captured } = captureRequest();

    await client.extractions.create(DOCUMENT, JSON_SCHEMA, 'retab-large');

    expect(captured).toHaveLength(1);
    expect('agentic_extraction' in captured[0]!.body).toBe(false);
  });

  test('uses the server field name, not the camelCase argument name', async () => {
    const { client, captured } = captureRequest();

    await client.extractions.create(...createArgs(true));

    expect('agenticExtraction' in captured[0]!.body).toBe(false);
    expect(captured[0]!.body.agentic_extraction).toBe(true);
  });

  test('appending the argument last leaves the earlier positions untouched', async () => {
    // deepExtraction stays the 11th argument: a caller who already passes it
    // positionally must not have it reinterpreted as agentic extraction.
    const { client, captured } = captureRequest();

    await client.extractions.create(
      DOCUMENT,
      JSON_SCHEMA,
      'retab-large',
      undefined,
      undefined,
      undefined,
      undefined,
      undefined,
      undefined,
      undefined,
      true
    );

    expect(captured[0]!.body.deep_extraction).toBe(true);
    expect('agentic_extraction' in captured[0]!.body).toBe(false);
  });

  test('does not set deep_extraction on the client', async () => {
    // The server applies "agentic implies deep"; the SDK sends what it was given.
    const { client, captured } = captureRequest();

    await client.extractions.create(...createArgs(true));

    expect('deep_extraction' in captured[0]!.body).toBe(false);
    for (const knob of [
      'long_list_strategy',
      'window_size',
      'holdback_rows',
      'long_list_target_field',
    ]) {
      expect(knob in captured[0]!.body).toBe(false);
    }
  });

  test('sends agentic_extraction on the streaming create too', async () => {
    // The server answers 422 for agentic extraction on the streaming route; the
    // SDK must still send it so the caller gets that error rather than a
    // silently ordinary streamed extraction.
    const { client, captured } = captureRequest();

    await client.extractions.create_stream(...createArgs(true));

    expect(captured).toHaveLength(1);
    expect(captured[0]!.url).toContain('/v1/extractions/stream');
    expect(captured[0]!.body.agentic_extraction).toBe(true);
  });

  test('omits agentic_extraction on the streaming create when unset', async () => {
    const { client, captured } = captureRequest();

    await client.extractions.create_stream(DOCUMENT, JSON_SCHEMA, 'retab-large');

    expect('agentic_extraction' in captured[0]!.body).toBe(false);
  });
});
