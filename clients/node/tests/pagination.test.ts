/**
 * Auto-pagination tests for the Node SDK.
 *
 * The new `PaginatedList<T>` class wraps a single page of results plus an
 * optional `_fetchPage(after)` closure that the resource client wires up.
 * Consumers can:
 *
 *   - call `result.data` for the current page (back-compat with the old
 *     return shape),
 *   - iterate `for await (const item of result)` to walk every page
 *     transparently, and
 *   - check `result.hasMore()` to know whether another page would be
 *     fetched.
 *
 * These tests stub the underlying transport (`_fetch`) so we can assert the
 * iterator's behaviour without hitting a real backend.
 */
import { describe, expect, test } from 'bun:test';

import APIV1 from '../src/api/client.js';
import { AbstractClient } from '../src/client.js';
import { PaginatedList } from '../src/api/_pagination.js';

type RecordedRequest = {
  url: string;
  method: string;
  params?: Record<string, unknown>;
  headers?: Record<string, unknown>;
  body?: Record<string, unknown>;
};

type PageFixture = {
  data: Array<Record<string, unknown>>;
  list_metadata: { before: string | null; after: string | null };
};

/**
 * Test fetcher that replays a queue of pre-built pages, one per call, and
 * records every request it sees. Once the queue is exhausted further calls
 * throw so the test surfaces unexpected re-fetches.
 */
class ScriptedFetcher extends AbstractClient {
  public requests: RecordedRequest[] = [];
  private pages: PageFixture[];

  constructor(pages: PageFixture[]) {
    super();
    this.pages = [...pages];
  }

  protected async _fetch(params: RecordedRequest): Promise<Response> {
    this.requests.push(params);
    const next = this.pages.shift();
    if (!next) {
      throw new Error(`ScriptedFetcher exhausted at request ${this.requests.length}`);
    }
    return new Response(JSON.stringify(next), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    });
  }
}

function makeExtraction(id: string): Record<string, unknown> {
  // Minimal ExtractionV2 shape — provides exactly the fields the Zod
  // schema requires plus the typed-shape fields the iterator/typing test
  // looks for. The schema is `.strip()` so any extra wire field is dropped
  // silently.
  return {
    id,
    file: { id: 'file_x', filename: 'invoice.pdf', mime_type: 'application/pdf' },
    model: 'retab-small',
    json_schema: { type: 'object' },
    output: { json: { foo: id } },
    metadata: {},
  };
}

describe('PaginatedList auto-pagination', () => {
  test('for-await yields items from every page (3 pages -> 6 items)', async () => {
    const fetcher = new ScriptedFetcher([
      {
        data: [makeExtraction('ex_1'), makeExtraction('ex_2')],
        list_metadata: { before: null, after: 'ex_2' },
      },
      {
        data: [makeExtraction('ex_3'), makeExtraction('ex_4')],
        list_metadata: { before: 'ex_2', after: 'ex_4' },
      },
      {
        data: [makeExtraction('ex_5'), makeExtraction('ex_6')],
        list_metadata: { before: 'ex_4', after: null },
      },
    ]);
    const api = new APIV1(fetcher);

    const firstPage = await api.extractions.list({ limit: 2 });

    // The first page exposes `.data` directly (back-compat).
    expect(firstPage.data.map((row) => row.id)).toEqual(['ex_1', 'ex_2']);
    expect(firstPage.hasMore()).toBe(true);
    expect(firstPage.list_metadata).toEqual({ before: null, after: 'ex_2' });

    // Walking the auto-paginating iterator fetches the remaining two pages.
    const collected: string[] = [];
    for await (const item of firstPage) {
      collected.push(item.id as string);
    }
    expect(collected).toEqual(['ex_1', 'ex_2', 'ex_3', 'ex_4', 'ex_5', 'ex_6']);

    // Three GET /extractions calls in total — one per page.
    expect(fetcher.requests).toHaveLength(3);
    for (const req of fetcher.requests) {
      expect(req.url).toBe('/extractions');
      expect(req.method).toBe('GET');
    }
    // The follow-up pages set the `after` cursor on the next request.
    expect(fetcher.requests[1].params).toMatchObject({ after: 'ex_2' });
    expect(fetcher.requests[2].params).toMatchObject({ after: 'ex_4' });
  });

  test('iteration stops cleanly when list_metadata.after is null', async () => {
    const fetcher = new ScriptedFetcher([
      {
        data: [makeExtraction('only_1')],
        list_metadata: { before: null, after: null },
      },
    ]);
    const api = new APIV1(fetcher);

    const page = await api.extractions.list({ limit: 10 });
    expect(page.hasMore()).toBe(false);

    const collected: string[] = [];
    for await (const item of page) {
      collected.push(item.id as string);
    }
    expect(collected).toEqual(['only_1']);
    // No follow-up fetches were made.
    expect(fetcher.requests).toHaveLength(1);
  });

  test('manually-constructed PaginatedList with no fetcher iterates only the current page', async () => {
    // Used as a building block in user code: hand-roll a page from in-memory
    // data without wiring up a fetcher. Iteration should yield exactly the
    // items the caller passed in and stop, regardless of what
    // `list_metadata.after` says.
    const page = new PaginatedList<{ id: string }>({
      data: [{ id: 'manual_1' }, { id: 'manual_2' }],
      list_metadata: { before: null, after: 'manual_2' },
    });

    expect(page.hasMore()).toBe(true);
    expect(page.list_metadata.after).toBe('manual_2');

    const collected: string[] = [];
    for await (const item of page) {
      collected.push(item.id);
    }
    expect(collected).toEqual(['manual_1', 'manual_2']);
  });

  test('list_metadata snake-case alias and listMetadata camelCase alias both work', async () => {
    const fetcher = new ScriptedFetcher([
      {
        data: [makeExtraction('ex_1')],
        list_metadata: { before: null, after: null },
      },
    ]);
    const api = new APIV1(fetcher);
    const page = await api.extractions.list();

    expect(page.list_metadata).toEqual({ before: null, after: null });
    expect(page.listMetadata).toEqual({ before: null, after: null });
    // Same object reference under both aliases.
    expect(page.listMetadata).toBe(page.list_metadata);
  });

  test('typed item shape: page.data[0] has the resource fields the schema parses', async () => {
    const fetcher = new ScriptedFetcher([
      {
        data: [makeExtraction('ex_1'), makeExtraction('ex_2')],
        list_metadata: { before: null, after: null },
      },
    ]);
    const api = new APIV1(fetcher);
    const page = await api.extractions.list();

    // After typed-item wiring, page.data[0] is an ExtractionV2 — so the
    // top-level fields the Zod schema describes are accessible without a
    // cast and at runtime.
    const first = page.data[0];
    expect(first.id).toBe('ex_1');
    expect(first.model).toBe('retab-small');
    expect(first.output).toBeDefined();
    expect(first.file.id).toBe('file_x');
  });

  test('files.list returns a PaginatedList with hasMore() honoring the cursor', async () => {
    // Sanity-check that the helper is wired through every resource — files
    // happens to use its own inline ZFile schema.
    const fetcher = new ScriptedFetcher([
      {
        data: [
          {
            object: 'file',
            id: 'file_1',
            filename: 'a.pdf',
            created_at: '2026-01-01T00:00:00Z',
            updated_at: '2026-01-01T00:00:00Z',
          },
        ],
        list_metadata: { before: null, after: 'file_1' },
      },
      {
        data: [
          {
            object: 'file',
            id: 'file_2',
            filename: 'b.pdf',
            created_at: '2026-01-01T00:00:00Z',
            updated_at: '2026-01-01T00:00:00Z',
          },
        ],
        list_metadata: { before: 'file_1', after: null },
      },
    ]);
    const api = new APIV1(fetcher);

    const page = await api.files.list({ limit: 1 });
    expect(page.hasMore()).toBe(true);

    const ids: string[] = [];
    for await (const file of page) {
      ids.push(file.id);
    }
    expect(ids).toEqual(['file_1', 'file_2']);
    expect(fetcher.requests).toHaveLength(2);
    expect(fetcher.requests[1].params).toMatchObject({ after: 'file_1' });
  });
});
