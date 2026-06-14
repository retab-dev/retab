// REAL end-to-end tests for the parses resource against a live server.
//
// CREDITLESS: list/get/paginate/filter + error paths ONLY. create()/cancel()
// run OCR/parse work and are never called here. get-by-id is discovered from an
// existing list row.
//
// Defensive row assertions: staging carries legacy rows, so we only assert the
// always-present `id` and, when present, a known `status`.

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
  describe('parses e2e', () => {
    test.skip(LIVE_SKIP_REASON, () => {});
  });
}

d('parses.list (live, read-only)', () => {
  test('returns a typed page with id-bearing rows', async () => {
    const client = liveClient();
    const page = await client.parses.list({ limit: 5 });

    expect(Array.isArray(page.data)).toBe(true);
    expect(page.data.length).toBeLessThanOrEqual(5);
    expect(typeof page.hasMore()).toBe('boolean');

    for (const p of page.data) {
      expect(typeof p.id).toBe('string');
      expect(p.id.length).toBeGreaterThan(0);
      if (p.status != null) {
        expect(STATUSES.has(p.status)).toBe(true);
      }
    }
  });

  test('respects the limit parameter', async () => {
    const client = liveClient();
    const page = await client.parses.list({ limit: 1 });
    expect(page.data.length).toBeLessThanOrEqual(1);
  });

  test('order=asc and order=desc both return typed pages', async () => {
    const client = liveClient();
    const asc = await client.parses.list({ limit: 3, order: 'asc' });
    const desc = await client.parses.list({ limit: 3, order: 'desc' });
    expect(Array.isArray(asc.data)).toBe(true);
    expect(Array.isArray(desc.data)).toBe(true);
  });

  test('YYYY-MM-DD from_date and to_date filters are accepted together', async () => {
    const client = liveClient();
    const page = await client.parses.list({
      fromDate: '2020-01-01',
      toDate: '2999-01-01',
      limit: 3,
    });
    expect(Array.isArray(page.data)).toBe(true);
  });

  test('a filename filter is accepted (may return an empty page)', async () => {
    const client = liveClient();
    const page = await client.parses.list({
      filename: 'node-e2e-no-such-file.zzz',
      limit: 3,
    });
    expect(Array.isArray(page.data)).toBe(true);
  });

  test('after-cursor paging advances past the first page when more exist', async () => {
    const client = liveClient();
    const first = await client.parses.list({ limit: 1 });
    const cursor = first.list_metadata.after;
    if (!first.hasMore() || !cursor) {
      expect(Array.isArray(first.data)).toBe(true);
      return;
    }
    const second = await client.parses.list({ limit: 1, after: cursor });
    expect(Array.isArray(second.data)).toBe(true);
    if (first.data[0] && second.data[0]) {
      expect(second.data[0].id).not.toBe(first.data[0].id);
    }
  });

  test('auto-pagination iterates lazily without duplicates', async () => {
    const client = liveClient();
    const page = await client.parses.list({ limit: 2 });
    const ids: string[] = [];
    for await (const p of page) {
      ids.push(p.id);
      if (ids.length >= 5) break;
    }
    expect(new Set(ids).size).toBe(ids.length);
  });
});

d('parses.get (live, read-only)', () => {
  test('get-by-id for an id discovered via list round-trips', async () => {
    const client = liveClient();
    const page = await client.parses.list({ limit: 1 });
    if (page.data.length === 0) return;
    const id = page.data[0].id;
    const got = await client.parses.get(id);
    expect(got.id).toBe(id);
  });

  test('get with includeOutput=false still round-trips the id', async () => {
    const client = liveClient();
    const page = await client.parses.list({ limit: 1 });
    if (page.data.length === 0) return;
    const id = page.data[0].id;
    const got = await client.parses.get(id, { includeOutput: false });
    expect(got.id).toBe(id);
  });
});

d('parses error paths (live)', () => {
  test('a bogus parse id yields a typed 404', async () => {
    const client = liveClient();
    let thrown: unknown;
    try {
      await client.parses.get('parse_does_not_exist_zzz');
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
      await client.parses.list({ fromDate: 'not-a-date' });
    } catch (e) {
      status = (e as { status?: number }).status;
    }
    expect(status).toBe(400);
  });
});
