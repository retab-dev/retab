// REAL end-to-end cross-cutting EDGE/NEGATIVE tests against a live server.
//
// These exercise the list-query contract boundaries that the per-resource e2e
// suites don't: bad enums, pathological limits, malformed cursors, inverted /
// future date ranges, unicode filters, empty-result envelope shape, 401 across
// MULTIPLE resources, and the typed-error contract (status, code/message body).
//
// CREDITLESS: list/get + error paths ONLY. No create/process/LLM calls.
//
// NOTE on tolerance: backends differ on whether a bad query param is a hard 400
// or is silently clamped/ignored. Where behavior is genuinely
// implementation-defined we accept "either a typed error OR a clean typed page"
// and assert only that the SDK never throws an *untyped* error or returns a
// malformed envelope. Cases that DO have a firm contract (404, 401, malformed
// from_date -> 400) are asserted strictly in the per-resource suites and here.

import { describe, expect, test } from 'bun:test';

import { RetabAuthenticationError, RetabError } from '../../src/index.js';
import { LIVE, LIVE_SKIP_REASON, bogusKeyClient, liveClient } from '../live.js';

const d = describe.skipIf(!LIVE);

if (!LIVE) {
  describe('edge-cases e2e', () => {
    test.skip(LIVE_SKIP_REASON, () => {});
  });
}

/** Run a list call, returning either the page or the thrown error. */
async function attempt<T>(fn: () => Promise<T>): Promise<{ page?: T; err?: unknown }> {
  try {
    return { page: await fn() };
  } catch (err) {
    return { err };
  }
}

d('list query: invalid order enum', () => {
  test('an invalid order value is either rejected (400) or ignored, never an untyped throw', async () => {
    const client = liveClient();
    const { page, err } = await attempt(() =>
      client.files.list({ order: 'sideways' as unknown as 'asc' })
    );
    if (err) {
      // A thrown error must be a typed RetabError carrying a numeric status.
      expect(err).toBeInstanceOf(RetabError);
      expect(typeof (err as RetabError).status).toBe('number');
      expect((err as RetabError).status).toBe(400);
    } else {
      expect(Array.isArray(page!.data)).toBe(true);
    }
  });
});

d('list query: pathological limits', () => {
  test('limit=0 returns a typed page or a typed 400', async () => {
    const client = liveClient();
    const { page, err } = await attempt(() => client.files.list({ limit: 0 }));
    if (err) {
      expect(err).toBeInstanceOf(RetabError);
      expect((err as RetabError).status).toBe(400);
    } else {
      expect(Array.isArray(page!.data)).toBe(true);
    }
  });

  test('a negative limit returns a typed page or a typed 400', async () => {
    const client = liveClient();
    const { page, err } = await attempt(() => client.files.list({ limit: -5 }));
    if (err) {
      expect(err).toBeInstanceOf(RetabError);
      expect((err as RetabError).status).toBe(400);
    } else {
      expect(Array.isArray(page!.data)).toBe(true);
    }
  });

  test('a huge limit is clamped/served as a typed page or rejected with a typed 400', async () => {
    const client = liveClient();
    const { page, err } = await attempt(() => client.files.list({ limit: 1_000_000 }));
    if (err) {
      expect(err).toBeInstanceOf(RetabError);
      expect((err as RetabError).status).toBe(400);
    } else {
      expect(Array.isArray(page!.data)).toBe(true);
    }
  });
});

d('list query: malformed after-cursor', () => {
  test('a junk cursor yields a typed error or a clean typed page, never an untyped throw', async () => {
    const client = liveClient();
    const { page, err } = await attempt(() =>
      client.files.list({ limit: 2, after: 'not-a-real-cursor-zzz' })
    );
    if (err) {
      expect(err).toBeInstanceOf(RetabError);
      expect(typeof (err as RetabError).status).toBe('number');
    } else {
      expect(Array.isArray(page!.data)).toBe(true);
    }
  });
});

d('list query: date-range edge cases', () => {
  test('to_date < from_date yields an empty (or clamped) typed page, not an error', async () => {
    const client = liveClient();
    const page = await client.files.list({
      fromDate: '2999-01-01',
      toDate: '2000-01-01',
      limit: 5,
    });
    expect(Array.isArray(page.data)).toBe(true);
    // An inverted range cannot match anything; expect an empty page.
    expect(page.data.length).toBe(0);
  });

  test('a far-future from_date yields an empty typed page', async () => {
    const client = liveClient();
    const page = await client.files.list({ fromDate: '2999-01-01', limit: 5 });
    expect(Array.isArray(page.data)).toBe(true);
    expect(page.data.length).toBe(0);
  });
});

d('list query: unicode filename filter', () => {
  test('a unicode filename filter returns a clean typed (likely empty) page', async () => {
    const client = liveClient();
    const page = await client.files.list({ filename: 'fïçhïér-日本語-🚀.pdf', limit: 5 });
    expect(Array.isArray(page.data)).toBe(true);
    // We don't assert emptiness (some org could legitimately have such a name);
    // we only assert the query round-trips and any rows are well-formed.
    for (const f of page.data) {
      expect(typeof f.id).toBe('string');
    }
  });
});

d('empty-result envelope shape', () => {
  test('a guaranteed-empty filter still returns a well-formed list envelope', async () => {
    const client = liveClient();
    const page = await client.files.list({
      filename: 'definitely-no-such-file-node-e2e-zzz.bin',
      limit: 5,
    });
    expect(Array.isArray(page.data)).toBe(true);
    expect(page.data.length).toBe(0);
    // Envelope contract: list_metadata present with before/after keys; an empty
    // page must not advertise more pages.
    expect(page.list_metadata).toBeDefined();
    expect('after' in page.list_metadata).toBe(true);
    expect('before' in page.list_metadata).toBe(true);
    expect(page.hasMore()).toBe(false);
    expect(page.listMetadata).toBe(page.list_metadata); // camelCase alias, same ref
  });
});

d('401 across multiple resources', () => {
  test('a junk key yields a typed 401 on every listable resource', async () => {
    const client = bogusKeyClient();
    const calls: Array<() => Promise<unknown>> = [
      () => client.files.list({ limit: 1 }),
      () => client.extractions.list({ limit: 1 }),
      () => client.parses.list({ limit: 1 }),
      () => client.classifications.list({ limit: 1 }),
      () => client.splits.list({ limit: 1 }),
      () => client.workflows.list({ limit: 1 }),
      () => client.secrets.list_secrets(),
    ];
    for (const call of calls) {
      const { err } = await attempt(call);
      expect(err).toBeInstanceOf(RetabAuthenticationError);
      expect((err as RetabAuthenticationError).status).toBe(401);
    }
  });
});

d('typed-error contract', () => {
  test('a 401 error carries status + a non-empty response body string', async () => {
    const client = bogusKeyClient();
    const { err } = await attempt(() => client.files.list({ limit: 1 }));
    expect(err).toBeInstanceOf(RetabError);
    const e = err as RetabError;
    expect(e.status).toBe(401);
    expect(typeof e.responseBody).toBe('string');
    expect(typeof e.message).toBe('string');
    expect(e.message.length).toBeGreaterThan(0);
    expect(e.name).toBe('RetabAuthenticationError');
  });

  test('a 404 error body is a parseable JSON error envelope when present', async () => {
    const client = liveClient();
    const { err } = await attempt(() => client.extractions.get('extr_no_such_id_zzz'));
    expect(err).toBeInstanceOf(RetabError);
    const e = err as RetabError;
    expect(e.status).toBe(404);
    // The body is JSON-ish for our gateway errors; if it parses, it should carry
    // a human-readable message/detail field. We tolerate a non-JSON body too
    // (proxy/edge errors) and only assert it is a string in that case.
    if (e.responseBody.trim().startsWith('{')) {
      const parsed = JSON.parse(e.responseBody) as Record<string, unknown>;
      const hasMessage =
        typeof parsed.detail === 'string' ||
        typeof parsed.message === 'string' ||
        typeof parsed.error === 'string';
      expect(hasMessage).toBe(true);
    } else {
      expect(typeof e.responseBody).toBe('string');
    }
  });
});
