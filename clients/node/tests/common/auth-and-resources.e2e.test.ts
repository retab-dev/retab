// REAL end-to-end tests for cross-cutting auth + creditless config resources
// (secrets, tables) against a live server.
//
// CREDITLESS: list/get config + an auth (401) probe with a junk key. Secrets are
// read-only here (list + get metadata) — we deliberately do NOT create/delete
// org secrets because they are shared org config and concurrent agents may rely
// on them. Tables are list/shape-only (table create ingests + bills a file).

import { describe, expect, test } from 'bun:test';

import { RetabAuthenticationError, RetabNotFoundError } from '../../src/index.js';
import { LIVE, LIVE_SKIP_REASON, bogusKeyClient, discoverProjectId, liveClient } from '../live.js';

const d = describe.skipIf(!LIVE);

if (!LIVE) {
  describe('auth + resources e2e', () => {
    test.skip(LIVE_SKIP_REASON, () => {});
  });
}

d('authentication (live)', () => {
  test('a junk API key yields a typed 401 on a list call', async () => {
    const client = bogusKeyClient();
    let thrown: unknown;
    try {
      await client.files.list({ limit: 1 });
    } catch (e) {
      thrown = e;
    }
    expect(thrown).toBeInstanceOf(RetabAuthenticationError);
    expect((thrown as RetabAuthenticationError).status).toBe(401);
  });

  test('a valid key authenticates and a list call succeeds', async () => {
    const client = liveClient();
    const page = await client.files.list({ limit: 1 });
    expect(Array.isArray(page.data)).toBe(true);
  });
});

d('secrets.list_secrets (live, read-only)', () => {
  test('returns a typed { secrets: [...] } envelope with secret-shaped rows', async () => {
    const client = liveClient();
    const res = await client.secrets.list_secrets();
    expect(Array.isArray(res.secrets)).toBe(true);
    for (const s of res.secrets) {
      expect(typeof s.name).toBe('string');
      expect(s.createdAt).toBeInstanceOf(Date);
      expect(s.updatedAt).toBeInstanceOf(Date);
    }
  });

  test('get_secret by a discovered name returns matching metadata', async () => {
    const client = liveClient();
    const res = await client.secrets.list_secrets();
    if (res.secrets.length === 0) return;
    const name = res.secrets[0].name;
    const got = await client.secrets.get_secret(name);
    expect(got.secret.name).toBe(name);
    expect(got.secret.createdAt).toBeInstanceOf(Date);
  });

  test('a bogus secret name yields a typed 404', async () => {
    const client = liveClient();
    let thrown: unknown;
    try {
      await client.secrets.get_secret('NODE_E2E_SECRET_DOES_NOT_EXIST_ZZZ');
    } catch (e) {
      thrown = e;
    }
    expect(thrown).toBeInstanceOf(RetabNotFoundError);
    expect((thrown as RetabNotFoundError).status).toBe(404);
  });
});

d('tables.list (live, read-only)', () => {
  test('returns a typed { tables: [...] } envelope for a discovered project', async () => {
    const client = liveClient();
    const projectId = await discoverProjectId(client);
    if (!projectId) return; // tables.list requires a project_id
    const res = await client.tables.list({ projectId });
    expect(Array.isArray(res.tables)).toBe(true);
    for (const t of res.tables) {
      expect(typeof t.id).toBe('string');
    }
  });

  test('tables.list without a project_id is rejected with a 400', async () => {
    const client = liveClient();
    let status: number | undefined;
    try {
      await client.tables.list({});
    } catch (e) {
      status = (e as { status?: number }).status;
    }
    expect(status).toBe(400);
  });

  test('a bogus table id yields a typed 404', async () => {
    const client = liveClient();
    let thrown: unknown;
    try {
      await client.tables.get('tbl_does_not_exist_zzz');
    } catch (e) {
      thrown = e;
    }
    expect(thrown).toBeInstanceOf(RetabNotFoundError);
    expect((thrown as RetabNotFoundError).status).toBe(404);
  });
});
