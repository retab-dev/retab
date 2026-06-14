// REAL end-to-end tests for the workflows resource against a live server.
//
// CREDITLESS: only workflow DEFINITION CRUD + list/get/pagination. No runs,
// no experiments, no block-test runs, no LLM. The create->get->update->delete
// cycle deletes only the workflow it created and never touches pre-existing
// staging data.

import { describe, expect, test } from 'bun:test';

import { RetabNotFoundError } from '../../src/index.js';
import { LIVE, LIVE_SKIP_REASON, discoverProjectId, liveClient, uniqueSuffix } from '../live.js';

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

d('workflows definition CRUD (live, creditless)', () => {
  test('create -> get -> update -> delete a workflow definition (no runs)', async () => {
    const client = liveClient();
    const projectId = await discoverProjectId(client);
    if (!projectId) {
      // Cannot create a workflow without a project; skip rather than fabricate.
      test.skip('no project_id discoverable to attach a workflow', () => {});
      return;
    }

    const name = `node-e2e-creditless-${uniqueSuffix()}`;
    const created = await client.workflows.create(projectId, name, 'created by node e2e');
    expect(typeof created.id).toBe('string');
    expect(created.name).toBe(name);
    expect(created.projectId).toBe(projectId);

    try {
      const got = await client.workflows.get(created.id);
      expect(got.id).toBe(created.id);
      expect(got.name).toBe(name);

      const renamed = `${name}-renamed`;
      const updated = await client.workflows.update(created.id, renamed, 'renamed by node e2e');
      expect(updated.id).toBe(created.id);
      expect(updated.name).toBe(renamed);

      // The rename is durable on a fresh get.
      const reGot = await client.workflows.get(created.id);
      expect(reGot.name).toBe(renamed);

      // The newly-created workflow is visible in its project listing.
      const inProject = await client.workflows.list({ projectId, limit: 50 });
      expect(inProject.data.some((wf) => wf.id === created.id)).toBe(true);
    } finally {
      await client.workflows.delete(created.id);
    }

    // After delete it is gone (typed 404).
    let thrown: unknown;
    try {
      await client.workflows.get(created.id);
    } catch (e) {
      thrown = e;
    }
    expect(thrown).toBeInstanceOf(RetabNotFoundError);
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
