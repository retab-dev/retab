/**
 * Runtime-introspection regression test for the SDK pagination contract.
 *
 * See [.notes/blueprints/sdk-pagination-contract.md] for the full contract.
 * In short: every `.list()` method on a public resource MUST route through
 * `AbstractClient._fetchPage<T>` so the returned `PaginatedList<T>` ships
 * with a wired-up `_fetchPage` closure. Without that closure, `for await`
 * iteration silently stops after the first page — `hasMore()` reports
 * `true` but auto-pagination yields nothing.
 *
 * This test proves the closure is wired by stubbing `_fetch`, queuing two
 * fixture envelopes whose `list_metadata.after` chain is "cursor-2" → null,
 * then iterating the page with `for await` and asserting that `_fetch` was
 * called more than once. If a list method bypasses `_fetchPage` and
 * constructs `PaginatedList` by hand, iteration stops after one fetch and
 * the assertion fails — naming the offending resource path.
 *
 * If you add a new resource, add it to `KNOWN_LIST_METHODS` below. The
 * Python and Go SDKs ship the same regression test against their own
 * resource registries — keep all three lists in sync.
 */
import { describe, expect, test } from 'bun:test';

import { Retab } from '../src/index.js';
import { AbstractClient, CompositionClient } from '../src/client.js';

type RecordedRequest = {
  url: string;
  method: string;
  params?: Record<string, unknown>;
  headers?: Record<string, unknown>;
  body?: Record<string, unknown>;
};

type PageFixture = {
  data: unknown[];
  list_metadata: { before: string | null; after: string | null };
};

/**
 * Test transport that hands back a queue of `{data, list_metadata}`
 * envelopes — one envelope per `_fetch` call — and records every request
 * it sees so the test can assert how many pages the closure actually
 * fetched.
 *
 * The default fixtures encode a two-page walk:
 *   page 1 → `{data: [], list_metadata: {before: null, after: "cursor-2"}}`
 *   page 2 → `{data: [], list_metadata: {before: null, after: null}}`
 *
 * The page bodies are empty arrays on purpose — `autoPagingIter` walks
 * `hasMore()` regardless of data length, so empty data still exercises the
 * follow-up fetch when the closure is correctly wired. That lets the test
 * cover every resource with the same fixture, instead of hand-rolling
 * minimally-valid items for each Zod schema.
 */
class ScriptedFetcher extends AbstractClient {
  public requests: RecordedRequest[] = [];
  private pages: PageFixture[];

  constructor(pages?: PageFixture[]) {
    super();
    this.pages = pages
      ? [...pages]
      : [
          { data: [], list_metadata: { before: null, after: 'cursor-2' } },
          { data: [], list_metadata: { before: null, after: null } },
        ];
  }

  protected async _fetch(params: RecordedRequest): Promise<Response> {
    this.requests.push(params);
    // If the test calls us more times than the queue accounts for, hand
    // back a terminal page rather than throwing — the test asserts on
    // `requests.length`, not on queue exhaustion.
    const next = this.pages.shift() ?? {
      data: [],
      list_metadata: { before: null, after: null },
    };
    return new Response(JSON.stringify(next), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    });
  }
}

/**
 * Build a `Retab` client whose underlying transport is a `ScriptedFetcher`.
 *
 * The public `Retab` constructor only takes options (`apiKey`, `baseUrl`),
 * not a custom transport — so we instantiate the real class once for shape
 * discovery, then mutate its protected `_client` reference to the scripted
 * fetcher. This is the same trick `pagination.test.ts` plays via `APIV1`,
 * but routed through the public surface so the discovery walk sees the
 * exact `Retab` instance a user would.
 */
function buildClientWithScriptedFetcher(): { client: Retab; fetcher: ScriptedFetcher } {
  const fetcher = new ScriptedFetcher();
  const client = new Retab({ apiKey: 'sentinel-pagination-contract', baseUrl: 'http://localhost' });
  (client as unknown as { _client: AbstractClient })._client = fetcher;
  return { client, fetcher };
}

/**
 * The full enumeration of resource list methods we expect to delegate to
 * `_fetchPage`. Each entry names the dotted path off the `Retab` instance
 * and supplies a minimal argument vector so the method can be invoked
 * without hitting "required field missing" exceptions.
 *
 * Sentinel string args ("sentinel_for_contract_test") are accepted because
 * the scripted fetcher ignores URL path content — it only cares about how
 * many times `_fetch` was invoked.
 *
 * TODO: when you add a new resource list method, add it here. The Python
 * and Go SDKs carry the same registry — keep all three in sync per the
 * pagination contract blueprint.
 */
type ListEntry = {
  path: string;
  invoke: (client: Retab) => Promise<{ data: unknown[] } & AsyncIterable<unknown>>;
};

const SENTINEL = 'sentinel_for_contract_test';

const KNOWN_LIST_METHODS: ListEntry[] = [
  // Top-level resources.
  { path: 'classifications.list', invoke: (c) => c.classifications.list() },
  { path: 'edits.list', invoke: (c) => c.edits.list() },
  { path: 'edits.templates.list', invoke: (c) => c.edits.templates.list() },
  { path: 'extractions.list', invoke: (c) => c.extractions.list() },
  { path: 'files.list', invoke: (c) => c.files.list() },
  { path: 'jobs.list', invoke: (c) => c.jobs.list() },
  { path: 'parses.list', invoke: (c) => c.parses.list() },
  { path: 'partitions.list', invoke: (c) => c.partitions.list() },
  { path: 'splits.list', invoke: (c) => c.splits.list() },
  // Workflows + nested sub-resources.
  { path: 'workflows.list', invoke: (c) => c.workflows.list() },
  { path: 'workflows.runs.list', invoke: (c) => c.workflows.runs.list() },
  {
    path: 'workflows.reviews.list',
    invoke: (c) => c.workflows.reviews.list(),
  },
  {
    path: 'workflows.reviews.versions.list',
    invoke: (c) => c.workflows.reviews.versions.list({ reviewId: SENTINEL }),
  },
  {
    path: 'workflows.blocks.list',
    invoke: (c) => c.workflows.blocks.list(SENTINEL),
  },
  {
    path: 'workflows.blocks.executions.list',
    invoke: (c) =>
      c.workflows.blocks.executions.list({ runId: SENTINEL, blockId: SENTINEL }),
  },
  {
    path: 'workflows.edges.list',
    invoke: (c) => c.workflows.edges.list(SENTINEL),
  },
  {
    path: 'workflows.artifacts.list',
    invoke: (c) => c.workflows.artifacts.list({ runId: SENTINEL }),
  },
  {
    path: 'workflows.steps.list',
    invoke: (c) => c.workflows.steps.list(SENTINEL),
  },
  {
    path: 'workflows.tests.list',
    invoke: (c) => c.workflows.tests.list({}),
  },
  {
    path: 'workflows.tests.runs.list',
    invoke: (c) => c.workflows.tests.runs.list({}),
  },
  {
    path: 'workflows.tests.results.list',
    invoke: (c) => c.workflows.tests.results.list({ runId: SENTINEL }),
  },
  {
    path: 'workflows.experiments.list',
    invoke: (c) => c.workflows.experiments.list(SENTINEL),
  },
  {
    path: 'workflows.experiments.runs.list',
    invoke: (c) => c.workflows.experiments.runs.list({}),
  },
  {
    path: 'workflows.experiments.results.list',
    invoke: (c) => c.workflows.experiments.results.list({ runId: SENTINEL }),
  },
];

/**
 * Allowlist for list methods that legitimately bypass the central
 * `_fetchPage` helper. Per the blueprint's "Acceptable exceptions"
 * section, the Node SDK has no entries today — every Node list method
 * routes through `AbstractClient._fetchPage`. The set exists so the test
 * surfaces a clear "add to allowlist" path if a future refactor needs a
 * custom decoder (e.g., dual-shape envelopes like Go's artifacts/blocks).
 */
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const KNOWN_BYPASS: ReadonlySet<string> = new Set();

/**
 * Walk the live `Retab` instance and collect every dotted path that ends
 * at a `CompositionClient` whose prototype exposes a `list` method. The
 * walk runs purely for parity with the hardcoded `KNOWN_LIST_METHODS`
 * registry above: if the walk discovers a `.list()` we don't know about,
 * the test fails and tells you to extend the registry.
 */
function discoverListPaths(
  target: object,
  pathPrefix = '',
  seen = new Set<unknown>(),
  depth = 0
): string[] {
  if (depth > 3 || seen.has(target)) {
    return [];
  }
  seen.add(target);

  const paths: string[] = [];
  for (const name of Object.keys(target)) {
    if (name.startsWith('_')) {
      continue;
    }
    const value = (target as Record<string, unknown>)[name];
    if (!(value instanceof CompositionClient)) {
      continue;
    }
    const resourcePath = pathPrefix ? `${pathPrefix}.${name}` : name;
    // Walk the prototype chain so inherited `list` methods are picked up.
    let proto: object | null = Object.getPrototypeOf(value);
    let hasList = false;
    while (proto && proto !== Object.prototype) {
      if (typeof (proto as Record<string, unknown>).list === 'function') {
        hasList = true;
        break;
      }
      proto = Object.getPrototypeOf(proto);
    }
    if (hasList) {
      paths.push(`${resourcePath}.list`);
    }
    paths.push(...discoverListPaths(value, resourcePath, seen, depth + 1));
  }
  return paths;
}

describe('SDK pagination contract', () => {
  test('discovered list methods match the registry', () => {
    const { client } = buildClientWithScriptedFetcher();
    const discovered = discoverListPaths(client as unknown as object).sort();
    const expected = KNOWN_LIST_METHODS.map((entry) => entry.path).sort();
    expect(
      discovered,
      'KNOWN_LIST_METHODS is out of sync with the live client surface — ' +
        'add or remove entries to match the discovered set'
    ).toEqual(expected);
  });

  test.each(KNOWN_LIST_METHODS.map((entry) => [entry.path, entry] as const))(
    'expects %s to wire _fetchPage closure',
    async (_path, entry) => {
      const { client, fetcher } = buildClientWithScriptedFetcher();
      const page = await entry.invoke(client);

      // Drain the auto-paging iterator. With the queued fixtures the
      // first page reports `after: "cursor-2"`; if the closure is wired
      // up correctly the iterator follows that cursor and triggers a
      // second `_fetch` call. If the closure is missing, iteration stops
      // after the first page and `requests.length === 1`.
      for await (const _ of page) {
        // body intentionally empty — both fixture pages return `data: []`
      }

      expect(
        fetcher.requests.length,
        `${entry.path} did not trigger a follow-up fetch — its PaginatedList ` +
          `must have come from _fetchPage with a wired closure. Re-route this ` +
          `method through AbstractClient._fetchPage<T> per the SDK pagination contract.`
      ).toBeGreaterThan(1);

      // Sanity: the follow-up request must include the cursor from page 1,
      // which is exactly what `_fetchPage`'s next-page closure layers on.
      expect(
        fetcher.requests[1]?.params,
        `${entry.path} follow-up request did not include the after cursor ` +
          `from list_metadata — the next-page closure did not preserve params correctly.`
      ).toMatchObject({ after: 'cursor-2' });
    }
  );
});
