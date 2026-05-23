/**
 * Typed, auto-paginating list result for the Retab Node SDK.
 *
 * Modelled on the WorkOS Python SDK's `AsyncPage` pattern: the resource
 * client's `_fetchPage(itemSchema, opts)` helper returns a `PaginatedList<T>`
 * that holds the current page (`data`, `list_metadata`) AND a private
 * `_fetchPage` closure that knows how to ask for the NEXT page from the same
 * URL with the same params plus the updated `after` cursor.
 *
 * That gives callers three ergonomic options on top of the same object:
 *
 *   1. Synchronous, one-page-at-a-time:
 *      ```ts
 *      const page = await client.extractions.list({ limit: 50 });
 *      for (const ex of page.data) { ... }
 *      if (page.hasMore()) {
 *          const next = await client.extractions.list({ after: page.list_metadata.after });
 *      }
 *      ```
 *
 *   2. Auto-pagination with `for await`:
 *      ```ts
 *      for await (const ex of await client.extractions.list({ limit: 50 })) {
 *          // walks every page transparently
 *      }
 *      ```
 *
 *   3. Manual iterator handle for advanced flow control:
 *      ```ts
 *      const iter = (await client.extractions.list()).autoPagingIter();
 *      const first = await iter.next();
 *      ```
 *
 * Both snake_case (`list_metadata`) and camelCase (`listMetadata`)
 * accessors are exposed. The snake_case spelling is the canonical wire
 * shape across the Retab API and matches every existing test assertion;
 * the camelCase alias is provided for callers that prefer JS conventions.
 */

export type ListMetadata = {
  before: string | null;
  after: string | null;
};

export type PaginatedListInit<T> = {
  data: T[];
  list_metadata: ListMetadata;
  /**
   * Optional next-page fetcher. The resource client wires this up so
   * `autoPagingIter()` can walk subsequent pages with the same filters and
   * limit, only changing the `after` cursor. Leaving it `undefined` keeps
   * `PaginatedList` usable for in-memory single-page values.
   */
  fetchNextPage?: (after: string) => Promise<PaginatedList<T>>;
};

export class PaginatedList<T> {
  public readonly data: T[];
  public readonly list_metadata: ListMetadata;
  private readonly _fetchPage?: (after: string) => Promise<PaginatedList<T>>;

  constructor(init: PaginatedListInit<T>) {
    this.data = init.data;
    this.list_metadata = init.list_metadata;
    this._fetchPage = init.fetchNextPage;
  }

  /**
   * CamelCase alias for `list_metadata`. Same reference, no copy — both
   * properties resolve to the exact object the API returned.
   */
  get listMetadata(): ListMetadata {
    return this.list_metadata;
  }

  /**
   * `true` when the backend returned a non-null `after` cursor, meaning
   * another page exists. Note: this is purely a wire-level check — it does
   * NOT also confirm a `_fetchPage` closure is wired up.
   */
  hasMore(): boolean {
    return this.list_metadata.after != null;
  }

  /**
   * Yield every item across the current page and all subsequent pages.
   *
   * If no `_fetchPage` closure was provided (e.g. the page was constructed
   * by hand), iteration stops cleanly after the current page's items —
   * regardless of what `list_metadata.after` says — because there is no
   * way to fetch more.
   */
  async *autoPagingIter(): AsyncGenerator<T> {
    // Walk the page we already have first.
    for (const item of this.data) {
      yield item;
    }

    // Only attempt to fetch more pages if we know how to.
    if (!this._fetchPage) {
      return;
    }

    // `current` advances along the page chain.
    // eslint-disable-next-line @typescript-eslint/no-this-alias
    let current: PaginatedList<T> = this;
    while (current.hasMore() && current._fetchPage) {
      const afterCursor = current.list_metadata.after;
      if (afterCursor == null) {
        // Defensive: `hasMore()` already guarded against this, but the
        // null-narrowing the type system needs lives here.
        break;
      }
      current = await current._fetchPage(afterCursor);
      for (const item of current.data) {
        yield item;
      }
    }
  }

  [Symbol.asyncIterator](): AsyncGenerator<T> {
    return this.autoPagingIter();
  }
}
