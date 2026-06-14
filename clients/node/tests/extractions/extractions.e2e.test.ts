// REAL end-to-end tests for the extractions resource against a live server.
//
// CREDITLESS: list/get/paginate/filter + error paths ONLY. We NEVER call
// create()/cancel() — those run LLM inference and bill credits. Every get-by-id
// is discovered from an existing list row, so nothing new is processed.
//
// Staging has legacy rows that violate the current contract (e.g. null fields),
// so row assertions are defensive: we only assert the always-present `id` and,
// when present, that `status` is one of the known lifecycle values.

import { describe, expect, test } from 'bun:test';

import { RetabNotFoundError } from '../../src/index.js';
import { LIVE, LIVE_SKIP_REASON, liveClient } from '../live.js';

const d = describe.skipIf(!LIVE);

const STATUSES = new Set([
  'pending',
  'queued',
  'in_progress',
  'completed',
  'failed',
  'cancelled',
]);

if (!LIVE) {
  describe('extractions e2e', () => {
    test.skip(LIVE_SKIP_REASON, () => {});
  });
}

d('extractions.list (live, read-only)', () => {
  test('returns a typed page with id-bearing rows', async () => {
    const client = liveClient();
    const page = await client.extractions.list({ limit: 5 });

    expect(Array.isArray(page.data)).toBe(true);
    expect(page.data.length).toBeLessThanOrEqual(5);
    expect(typeof page.hasMore()).toBe('boolean');

    for (const e of page.data) {
      expect(typeof e.id).toBe('string');
      expect(e.id.length).toBeGreaterThan(0);
      // status is optional on the wire; if present it must be a known value.
      if (e.status != null) {
        expect(STATUSES.has(e.status)).toBe(true);
      }
    }
  });

  test('respects the limit parameter', async () => {
    const client = liveClient();
    const page = await client.extractions.list({ limit: 1 });
    expect(page.data.length).toBeLessThanOrEqual(1);
  });

  test('order=asc and order=desc both return typed pages', async () => {
    const client = liveClient();
    const asc = await client.extractions.list({ limit: 3, order: 'asc' });
    const desc = await client.extractions.list({ limit: 3, order: 'desc' });
    expect(Array.isArray(asc.data)).toBe(true);
    expect(Array.isArray(desc.data)).toBe(true);
  });

  test('a YYYY-MM-DD from_date filter is accepted', async () => {
    const client = liveClient();
    const page = await client.extractions.list({ fromDate: '2020-01-01', limit: 3 });
    expect(Array.isArray(page.data)).toBe(true);
  });

  test('a status filter is accepted and rows match when returned', async () => {
    const client = liveClient();
    const page = await client.extractions.list({ status: 'completed', limit: 5 });
    expect(Array.isArray(page.data)).toBe(true);
    for (const e of page.data) {
      if (e.status != null) expect(e.status).toBe('completed');
    }
  });

  test('after-cursor paging advances past the first page when more exist', async () => {
    const client = liveClient();
    const first = await client.extractions.list({ limit: 1 });
    const cursor = first.list_metadata.after;
    if (!first.hasMore() || !cursor) {
      expect(Array.isArray(first.data)).toBe(true);
      return;
    }
    const second = await client.extractions.list({ limit: 1, after: cursor });
    expect(Array.isArray(second.data)).toBe(true);
    if (first.data[0] && second.data[0]) {
      expect(second.data[0].id).not.toBe(first.data[0].id);
    }
  });

  test('auto-pagination iterates lazily without duplicates', async () => {
    const client = liveClient();
    const page = await client.extractions.list({ limit: 2 });
    const ids: string[] = [];
    for await (const e of page) {
      ids.push(e.id);
      if (ids.length >= 5) break; // bound the walk; stays fast + creditless
    }
    expect(new Set(ids).size).toBe(ids.length);
  });
});

d('extractions.get (live, read-only)', () => {
  test('get-by-id for an id discovered via list round-trips', async () => {
    const client = liveClient();
    const page = await client.extractions.list({ limit: 1 });
    if (page.data.length === 0) return; // empty org: nothing to get
    const id = page.data[0].id;
    const got = await client.extractions.get(id);
    expect(got.id).toBe(id);
  });
});

d('extractions error paths (live)', () => {
  test('a bogus extraction id yields a typed 404', async () => {
    const client = liveClient();
    let thrown: unknown;
    try {
      await client.extractions.get('extr_does_not_exist_zzz');
    } catch (e) {
      thrown = e;
    }
    expect(thrown).toBeInstanceOf(RetabNotFoundError);
    expect((thrown as RetabNotFoundError).status).toBe(404);
  });

  test('a malformed from_date yields a 400', async () => {
    const client = liveClient();
    let status: number | undefined;
    try {
      await client.extractions.list({ fromDate: 'not-a-date' });
    } catch (e) {
      status = (e as { status?: number }).status;
    }
    expect(status).toBe(400);
  });
});
