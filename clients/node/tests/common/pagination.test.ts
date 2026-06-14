// Offline unit tests for src/_pagination.ts + RetabBase._fetchPage.
// Drives client.files.list and client.workflows.runs.list with canned multi-page
// Responses via injected fetch to exercise autoPagingIter cursor walking, the
// loop guard, before-drop on forward paging, empty-page termination, and that
// _fetchPage preserves the original query/headers while only changing `after`.

import { describe, expect, test } from 'bun:test';

import { Retab } from '../../src/index.js';
import { PaginatedList } from '../../src/_pagination.js';

function fileWire(id: string) {
  return { object: 'file', id, filename: `${id}.pdf`, mime_type: 'application/pdf' };
}

function runWire(id: string) {
  return {
    id,
    workflow_id: 'wf_1',
    workflow_version_id: 'ver_1',
    trigger: { type: 'manual' },
    lifecycle: { status: 'pending' },
    timing: { created_at: '2026-05-28T00:00:00Z' },
  };
}

type Recorded = { url: URL };

/**
 * Build a client whose injected fetch serves a sequence of envelopes keyed by
 * the `after` cursor on the request. The first request has no `after`.
 */
function clientServingPages(
  pagesByAfter: Record<string, { data: unknown[]; after: string | null; before?: string | null }>
): { client: Retab; recorded: Recorded[] } {
  const recorded: Recorded[] = [];
  const client = new Retab({
    apiKey: 'test_key',
    fetch: async (input) => {
      const url = new URL(String(input));
      recorded.push({ url });
      const after = url.searchParams.get('after') ?? '__first__';
      const page = pagesByAfter[after];
      if (!page) throw new Error(`no canned page for after=${after}`);
      return new Response(
        JSON.stringify({
          data: page.data,
          list_metadata: { before: page.before ?? null, after: page.after },
        }),
        { status: 200, headers: { 'content-type': 'application/json' } }
      );
    },
  });
  return { client, recorded };
}

describe('autoPagingIter walks the after cursor across pages (files.list)', () => {
  test('iterates two pages then terminates on null after', async () => {
    const { client, recorded } = clientServingPages({
      __first__: { data: [fileWire('file_a'), fileWire('file_b')], after: 'cursor_1' },
      cursor_1: { data: [fileWire('file_c')], after: null },
    });

    const page = await client.files.list({ limit: 2 });
    const ids: string[] = [];
    for await (const f of page) ids.push(f.id);

    expect(ids).toEqual(['file_a', 'file_b', 'file_c']);
    // Two HTTP requests: the first page, then one follow-up for cursor_1.
    expect(recorded.length).toBe(2);
    expect(recorded[0].url.searchParams.has('after')).toBe(false);
    expect(recorded[1].url.searchParams.get('after')).toBe('cursor_1');
  });

  test('single page with null after makes exactly one request', async () => {
    const { client, recorded } = clientServingPages({
      __first__: { data: [fileWire('only')], after: null },
    });
    const page = await client.files.list();
    const ids: string[] = [];
    for await (const f of page) ids.push(f.id);
    expect(ids).toEqual(['only']);
    expect(recorded.length).toBe(1);
  });
});

describe('autoPagingIter loop guard + empty-page termination', () => {
  test('a repeated after cursor breaks the loop (seenAfterCursors guard)', async () => {
    const { client, recorded } = clientServingPages({
      // First page points to cursor_loop; the cursor_loop page points to itself.
      __first__: { data: [fileWire('p0')], after: 'cursor_loop' },
      cursor_loop: { data: [fileWire('p1')], after: 'cursor_loop' },
    });

    const page = await client.files.list();
    const ids: string[] = [];
    for await (const f of page) ids.push(f.id);

    // p0 (first page) + p1 (cursor_loop fetched once); the second sight of
    // cursor_loop is rejected by the guard before another fetch.
    expect(ids).toEqual(['p0', 'p1']);
    expect(recorded.length).toBe(2);
    expect(recorded[1].url.searchParams.get('after')).toBe('cursor_loop');
  });

  test('an empty follow-up page terminates iteration', async () => {
    const { client, recorded } = clientServingPages({
      __first__: { data: [fileWire('x')], after: 'cursor_empty' },
      // Backend returns more=after but an empty data array on the next page.
      cursor_empty: { data: [], after: 'cursor_should_not_be_followed' },
    });

    const page = await client.files.list();
    const ids: string[] = [];
    for await (const f of page) ids.push(f.id);

    expect(ids).toEqual(['x']);
    // first page + the empty cursor_empty page; iteration stops on empty data
    // so cursor_should_not_be_followed is never requested.
    expect(recorded.length).toBe(2);
    expect(recorded[1].url.searchParams.get('after')).toBe('cursor_empty');
  });

  test('iteration stops cleanly when no next-page closure is wired', async () => {
    const page = new PaginatedList<string>({
      data: ['one', 'two'],
      // after is set, but no fetchNextPage closure -> must not throw, stop after page.
      list_metadata: { before: null, after: 'cursor_dangling' },
    });
    const items: string[] = [];
    for await (const item of page) items.push(item);
    expect(items).toEqual(['one', 'two']);
    expect(page.hasMore()).toBe(true);
  });
});

describe('_fetchPage preserves original query/headers, only changing after', () => {
  test('forward paging carries filters and drops before', async () => {
    const recorded: URL[] = [];
    const client = new Retab({
      apiKey: 'test_key',
      fetch: async (input) => {
        const url = new URL(String(input));
        recorded.push(url);
        const after = url.searchParams.get('after');
        const next = after === null ? 'cursor_1' : null;
        return new Response(
          JSON.stringify({
            data: [fileWire(after ?? 'first')],
            list_metadata: { before: null, after: next },
          }),
          { status: 200, headers: { 'content-type': 'application/json' } }
        );
      },
    });

    const page = await client.files.list({
      filename: 'invoice.pdf',
      mimeType: 'application/pdf',
      limit: 10,
      order: 'desc',
      before: 'cursor_before',
    });
    const ids: string[] = [];
    for await (const f of page) ids.push(f.id);

    expect(recorded.length).toBe(2);

    // First request: original filters + before present, no after.
    const first = recorded[0].searchParams;
    expect(first.get('filename')).toBe('invoice.pdf');
    expect(first.get('mime_type')).toBe('application/pdf');
    expect(first.get('limit')).toBe('10');
    expect(first.get('order')).toBe('desc');
    expect(first.get('before')).toBe('cursor_before');
    expect(first.has('after')).toBe(false);

    // Second request: same filters, after set, before dropped.
    const second = recorded[1].searchParams;
    expect(second.get('filename')).toBe('invoice.pdf');
    expect(second.get('mime_type')).toBe('application/pdf');
    expect(second.get('limit')).toBe('10');
    expect(second.get('order')).toBe('desc');
    expect(second.get('after')).toBe('cursor_1');
    expect(second.has('before')).toBe(false);
  });
});

describe('workflows.runs.list paginates the same way', () => {
  test('walks two pages of workflow runs by cursor', async () => {
    const recorded: URL[] = [];
    const client = new Retab({
      apiKey: 'test_key',
      fetch: async (input) => {
        const url = new URL(String(input));
        recorded.push(url);
        const after = url.searchParams.get('after');
        if (after === null) {
          return new Response(
            JSON.stringify({
              data: [runWire('run_1'), runWire('run_2')],
              list_metadata: { before: null, after: 'cursor_runs_1' },
            }),
            { status: 200, headers: { 'content-type': 'application/json' } }
          );
        }
        return new Response(
          JSON.stringify({
            data: [runWire('run_3')],
            list_metadata: { before: null, after: null },
          }),
          { status: 200, headers: { 'content-type': 'application/json' } }
        );
      },
    });

    const page = await client.workflows.runs.list({ workflowId: 'wf_1', limit: 2 });
    const ids: string[] = [];
    for await (const r of page) ids.push(r.id);

    expect(ids).toEqual(['run_1', 'run_2', 'run_3']);
    expect(recorded.length).toBe(2);
    expect(recorded[0].searchParams.get('workflow_id')).toBe('wf_1');
    // Filter must survive into the follow-up page request.
    expect(recorded[1].searchParams.get('workflow_id')).toBe('wf_1');
    expect(recorded[1].searchParams.get('after')).toBe('cursor_runs_1');
  });
});
