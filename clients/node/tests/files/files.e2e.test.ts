// REAL end-to-end tests for the files resource against a live Retab server.
//
// These are NOT the offline contract tests in files.test.ts — they construct a
// real client (RETAB_API_BASE_URL + RETAB_API_KEY) and round-trip live calls.
//
// CREDITLESS: list/filter/paginate/get plus the storage-only upload lifecycle
// (create_upload -> signed PUT -> complete_upload). No parse/OCR/extraction is
// ever triggered, so no credits are consumed.

import { createHash } from 'crypto';

import { describe, expect, test } from 'bun:test';

import { RetabNotFoundError } from '../../src/index.js';
import { LIVE, LIVE_SKIP_REASON, liveClient, uploadTinyPdf } from '../live.js';

const d = describe.skipIf(!LIVE);

if (!LIVE) {
  describe('files e2e', () => {
    test.skip(LIVE_SKIP_REASON, () => {});
  });
}

d('files.list (live)', () => {
  test('returns a typed page with file-shaped rows', async () => {
    const client = liveClient();
    const page = await client.files.list({ limit: 5 });

    expect(Array.isArray(page.data)).toBe(true);
    expect(page.data.length).toBeLessThanOrEqual(5);
    expect(typeof page.hasMore()).toBe('boolean');

    for (const f of page.data) {
      expect(f.object).toBe('file');
      expect(typeof f.id).toBe('string');
      expect(typeof f.filename).toBe('string');
    }
  });

  test('respects the limit parameter', async () => {
    const client = liveClient();
    const page = await client.files.list({ limit: 1 });
    expect(page.data.length).toBeLessThanOrEqual(1);
  });

  test('order=asc and order=desc both return typed pages', async () => {
    const client = liveClient();
    const asc = await client.files.list({ limit: 3, order: 'asc' });
    const desc = await client.files.list({ limit: 3, order: 'desc' });
    expect(Array.isArray(asc.data)).toBe(true);
    expect(Array.isArray(desc.data)).toBe(true);
  });

  test('a YYYY-MM-DD from_date filter is accepted and returns a typed page', async () => {
    const client = liveClient();
    const page = await client.files.list({ fromDate: '2020-01-01', limit: 3 });
    expect(Array.isArray(page.data)).toBe(true);
  });

  test('after-cursor paging advances past the first page when more exist', async () => {
    const client = liveClient();
    const first = await client.files.list({ limit: 1 });
    const cursor = first.list_metadata.after;
    if (!first.hasMore() || !cursor) {
      // Not enough data to page; nothing to assert beyond a clean first page.
      expect(Array.isArray(first.data)).toBe(true);
      return;
    }
    const second = await client.files.list({ limit: 1, after: cursor });
    expect(Array.isArray(second.data)).toBe(true);
    if (first.data[0] && second.data[0]) {
      expect(second.data[0].id).not.toBe(first.data[0].id);
    }
  });

  test('auto-pagination iterates lazily across the after cursor', async () => {
    const client = liveClient();
    const page = await client.files.list({ limit: 2 });
    const ids: string[] = [];
    for await (const f of page) {
      ids.push(f.id);
      if (ids.length >= 5) break; // bound the walk so the test stays fast + creditless
    }
    expect(new Set(ids).size).toBe(ids.length); // no duplicates across pages
  });
});

d('files.get (live)', () => {
  test('get-by-id for an id discovered via list round-trips', async () => {
    const client = liveClient();
    const page = await client.files.list({ limit: 1 });
    if (page.data.length === 0) {
      test.skip('no files in org to get-by-id', () => {});
      return;
    }
    const id = page.data[0].id;
    const got = await client.files.get(id);
    expect(got.id).toBe(id);
    expect(typeof got.filename).toBe('string');
    // NOTE: unlike list rows, GET /v1/files/{id} omits the `object` discriminator
    // on the wire, so it deserializes to undefined here (see report).
  });

  test('a download link can be minted for an existing file', async () => {
    const client = liveClient();
    const page = await client.files.list({ limit: 1 });
    if (page.data.length === 0) return;
    const link = await client.files.get_download_link(page.data[0].id);
    // The FileLink field is `downloadUrl` (wire `download_url`), not `url`.
    expect(typeof link.downloadUrl).toBe('string');
    expect(link.downloadUrl.length).toBeGreaterThan(0);
    expect(typeof link.expiresIn).toBe('string');
  });
});

d('files upload lifecycle (live, storage-only)', () => {
  test('create_upload -> signed PUT -> complete_upload -> get round-trips', async () => {
    const client = liveClient();
    const { fileId, filename } = await uploadTinyPdf(client);

    expect(typeof fileId).toBe('string');
    expect(fileId.length).toBeGreaterThan(0);

    const got = await client.files.get(fileId);
    expect(got.id).toBe(fileId);
    expect(got.filename).toBe(filename);

    // The freshly-uploaded file is also retrievable as a download link.
    const link = await client.files.get_download_link(fileId);
    expect(typeof link.downloadUrl).toBe('string');
    // NOTE: the files resource exposes no delete; uploaded test blobs cannot be
    // cleaned up via the SDK (see report — this is a known gap).
  });

  test('create_upload returns a signed-URL envelope (fileId + http PUT url + headers)', async () => {
    const client = liveClient();
    const filename = `node-e2e-shape-${Date.now()}.pdf`;
    const bytes = Buffer.from('%PDF-1.4\n', 'utf8');
    const sha256 = createHash('sha256').update(bytes).digest('hex');

    const created = await client.files.create_upload(
      filename,
      bytes.length,
      'application/pdf',
      sha256
    );
    expect(typeof created.fileId).toBe('string');
    expect(created.fileId.length).toBeGreaterThan(0);
    expect(typeof created.uploadUrl).toBe('string');
    expect(created.uploadUrl.startsWith('http')).toBe(true);
    expect(typeof created.uploadMethod).toBe('string');
    expect(typeof created.uploadHeaders).toBe('object');
    // NOTE: mimeData is nullable on create_upload and is null at this stage even
    // when a sha256 is supplied; it is only guaranteed by complete_upload.
    expect(created.mimeData === null || typeof created.mimeData === 'object').toBe(true);
  });
});

d('files error paths (live)', () => {
  test('a bogus file id yields a typed 404', async () => {
    const client = liveClient();
    let thrown: unknown;
    try {
      await client.files.get('file_does_not_exist_zzz');
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
      // `not-a-date` is not a valid YYYY-MM-DD / RFC3339 value.
      await client.files.list({ fromDate: 'not-a-date' as unknown as string });
    } catch (e) {
      status = (e as { status?: number }).status;
    }
    expect(status).toBe(400);
  });
});
