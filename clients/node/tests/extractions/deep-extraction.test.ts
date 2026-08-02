import { describe, expect, test } from 'bun:test';

import { Retab } from '../../src/index.js';

// Wire-shape tests for the deep_extraction extraction knob.
//
// deep_extraction opts an extraction into the server's conversational
// long-array strategy. Its failure modes on the client side are all SILENT: a
// dropped field, a camelCase key the API ignores, or an unset default that
// serializes as `false` each produce a request that succeeds and quietly
// returns the truncated single-shot output the flag exists to fix. So these
// assert the request body actually put on the wire.

type CapturedRequest = { url: string; body: Record<string, unknown> };

// captureRequest builds a client over an injected fetch (the SDK's supported
// seam) and returns what the client actually sent. Going through the real
// client rather than asserting on a hand-built body is the point: the generated
// method maps its camelCase argument onto a snake_case wire key, and that
// mapping is what can silently go missing.
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
          file: { id: 'file_1', filename: 'ledger.pdf', mime_type: 'application/pdf' },
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

const DOCUMENT = { filename: 'ledger.pdf', url: 'data:application/pdf;base64,QUJD' };
const JSON_SCHEMA = { type: 'object', properties: { rows: { type: 'array' } } };

describe('deep_extraction request shape', () => {
  test('sends deep_extraction when enabled', async () => {
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

    expect(captured).toHaveLength(1);
    expect(captured[0]!.body.deep_extraction).toBe(true);
  });

  test('omits deep_extraction when unset', async () => {
    // Absent, not false: the default request body must stay identical to what
    // it was before this field existed, so the server takes its "absent means
    // default" branch for every caller who never asked for the feature.
    const { client, captured } = captureRequest();

    await client.extractions.create(DOCUMENT, JSON_SCHEMA, 'retab-large');

    expect(captured).toHaveLength(1);
    expect('deep_extraction' in captured[0]!.body).toBe(false);
  });

  test('sends an explicit false so an opt-out is distinguishable from unset', async () => {
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
      false
    );

    expect(captured[0]!.body.deep_extraction).toBe(false);
  });

  test('uses the server field name, not the camelCase argument name', async () => {
    // The method argument is deepExtraction; the API reads deep_extraction. A
    // mapping slip would send a key the server ignores and run a shallow
    // extraction with no error anywhere.
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

    expect('deepExtraction' in captured[0]!.body).toBe(false);
    expect(captured[0]!.body.deep_extraction).toBe(true);
  });

  test('does not send the tuning knobs', async () => {
    // deep_extraction is a single opaque switch: window size and per-window row
    // holdback stay at the server's defaults. An explicit holdback_rows of 0
    // would DISABLE the boundary-row holdback rather than default it to 1.
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

    for (const knob of [
      'window_size',
      'holdback_rows',
      'long_list_strategy',
      'long_list_target_field',
    ]) {
      expect(knob in captured[0]!.body).toBe(false);
    }
  });

  test('sends deep_extraction on the streaming create too', async () => {
    // The streaming create builds its own request body; the server dispatches
    // the conversational strategy on its streaming path, so a field wired into
    // only the non-streaming create is a real hole.
    const { client, captured } = captureRequest();

    await client.extractions.create_stream(
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

    expect(captured).toHaveLength(1);
    expect(captured[0]!.url).toContain('/v1/extractions/stream');
    expect(captured[0]!.body.deep_extraction).toBe(true);
  });
});
