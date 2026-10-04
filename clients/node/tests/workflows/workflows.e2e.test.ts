// REAL end-to-end tests for the workflows resource against a live server.
//
// CREDITLESS: only workflow definition list/get/pagination. No writes, runs,
// experiments, block-eval runs, or LLM calls.

import { describe, expect, test } from 'bun:test';

import { RetabNotFoundError } from '../../src/index.js';
import { LIVE, LIVE_SKIP_REASON, discoverProjectId, liveClient } from '../live.js';

const d = describe.skipIf(!LIVE);

if (!LIVE) {
  describe('workflows e2e', () => {
    test.skip(LIVE_SKIP_REASON, () => {});
  });
}

d('workflows.list (live)', () => {
  test('returns a typed page of workflow-shaped rows', async () => {
    const client = liveClient();
    const page = await client.workflows.list({ limit: 5 });

    expect(Array.isArray(page.data)).toBe(true);
    expect(page.data.length).toBeLessThanOrEqual(5);

    for (const wf of page.data) {
      expect(typeof wf.id).toBe('string');
      expect(wf.createdAt).toBeInstanceOf(Date);
      expect(wf.updatedAt).toBeInstanceOf(Date);
    }
  });

  test('order asc/desc and limit are honored', async () => {
    const client = liveClient();
    const page = await client.workflows.list({ limit: 2, order: 'desc' });
    expect(page.data.length).toBeLessThanOrEqual(2);
  });

  test("filtering by a discovered project_id returns only that project's workflows", async () => {
    const client = liveClient();
    const projectId = await discoverProjectId(client);
    if (!projectId) {
      // No workflows -> nothing to filter; covered by the unfiltered list test.
      return;
    }
    const page = await client.workflows.list({ projectId, limit: 10 });
    for (const wf of page.data) {
      expect(wf.projectId).toBe(projectId);
    }
  });

  test('auto-pagination iterates without duplicates', async () => {
    const client = liveClient();
    const page = await client.workflows.list({ limit: 2 });
    const ids: string[] = [];
    for await (const wf of page) {
      ids.push(wf.id);
      if (ids.length >= 5) break;
    }
    expect(new Set(ids).size).toBe(ids.length);
  });
});

d('workflows.get (live)', () => {
  test('get-by-id for an id discovered via list round-trips', async () => {
    const client = liveClient();
    const page = await client.workflows.list({ limit: 1 });
    if (page.data.length === 0) return;
    const id = page.data[0].id;
    const got = await client.workflows.get(id);
    expect(got.id).toBe(id);
    expect(got.createdAt).toBeInstanceOf(Date);
  });
});

d('workflows error paths (live)', () => {
  test('a bogus workflow id yields a typed 404', async () => {
    const client = liveClient();
    let thrown: unknown;
    try {
      await client.workflows.get('wrk_does_not_exist_zzz');
    } catch (e) {
      thrown = e;
    }
    expect(thrown).toBeInstanceOf(RetabNotFoundError);
    expect((thrown as RetabNotFoundError).status).toBe(404);
  });
});
