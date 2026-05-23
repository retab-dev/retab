"""Paginated list envelopes for `.list()` endpoints.

Mirrors the WorkOS `SyncPage` / `AsyncPage` split:

* `PaginatedList[T]` — sync. `__iter__` returns `auto_paging_iter()`, so
  ``for item in page:`` walks every page transparently. The fetch
  closure returns `PaginatedList[T]` synchronously.
* `AsyncPaginatedList[T]` — async. `__aiter__` returns the async
  `auto_paging_iter()`, so ``async for item in page:`` walks every
  page. The fetch closure returns ``Awaitable[AsyncPaginatedList[T]]``.

The wire shape (``{"data": [...], "list_metadata": {...}}``) is
identical on both — only the iteration protocol and the fetch closure's
return type differ. Constructors keep accepting ``Cls(**response)`` so
direct dict-construction by callers still works.
"""

from typing import (
    Any,
    AsyncIterator,
    Awaitable,
    Callable,
    Generic,
    Iterator,
    Literal,
    TypeVar,
)

from pydantic import BaseModel, ConfigDict, PrivateAttr
from retab.types.base import RetabBaseModel


T = TypeVar("T")


class ListMetadata(RetabBaseModel):
    """Boundary resource IDs for page navigation."""

    before: str | None
    after: str | None


class PaginatedList(RetabBaseModel, Generic[T]):
    """Sync paginated envelope.

    Iterating with ``for item in page:`` transparently fetches subsequent
    pages via the wired ``_fetch_next_page`` closure. Indexing (`page[i]`)
    and `len(page)` operate on the current page only.
    """

    model_config = ConfigDict(arbitrary_types_allowed=True)

    data: list[T]
    list_metadata: ListMetadata

    _fetch_next_page: Callable[..., "PaginatedList[T]"] | None = PrivateAttr(default=None)

    def __iter__(self) -> Iterator[T]:  # type: ignore[override]
        # Walks every page. Matches WorkOS SyncPage.__iter__.
        return self.auto_paging_iter()

    def __len__(self) -> int:
        return len(self.data)

    def __getitem__(self, index: int) -> T:
        return self.data[index]

    @property
    def has_more(self) -> bool:
        """Whether there are more pages available after this page's last resource ID."""
        return self.list_metadata.after is not None

    def auto_paging_iter(self) -> Iterator[T]:
        """Iterate through all items across all pages automatically.

        Yields items from the current page, then fetches subsequent pages
        until no more are available.
        """
        page: PaginatedList[T] = self
        while True:
            yield from page.data
            if not page.has_more or page._fetch_next_page is None:
                break
            page = page._fetch_next_page(after=page.list_metadata.after)


class AsyncPaginatedList(RetabBaseModel, Generic[T]):
    """Async paginated envelope.

    Iterating with ``async for item in page:`` transparently awaits the
    next page via the wired ``_fetch_next_page`` closure (an async
    callable). Indexing and ``len`` still operate on the current page.
    """

    model_config = ConfigDict(arbitrary_types_allowed=True)

    data: list[T]
    list_metadata: ListMetadata

    _fetch_next_page: Callable[..., Awaitable["AsyncPaginatedList[T]"]] | None = PrivateAttr(default=None)

    def __aiter__(self) -> AsyncIterator[T]:
        # Walks every page. Matches WorkOS AsyncPage.__aiter__.
        return self.auto_paging_iter()

    def __len__(self) -> int:
        return len(self.data)

    def __getitem__(self, index: int) -> T:
        return self.data[index]

    @property
    def has_more(self) -> bool:
        """Whether there are more pages available after this page's last resource ID."""
        return self.list_metadata.after is not None

    async def auto_paging_iter(self) -> AsyncIterator[T]:
        """Iterate through all items across all pages automatically (async).

        Yields items from the current page, then awaits subsequent pages
        until no more are available.
        """
        page: AsyncPaginatedList[T] = self
        while True:
            for item in page.data:
                yield item
            if not page.has_more or page._fetch_next_page is None:
                break
            page = await page._fetch_next_page(after=page.list_metadata.after)


PaginationOrder = Literal["asc", "desc"]


def _validate_page_items(raw_items: Any, model: type[Any] | None) -> list[Any]:
    """Re-validate raw items into ``model`` instances when possible.

    Keeps the centralized item-coercion logic in one place, so the
    per-resource list methods don't repeat the ``[Cls.model_validate(...)
    for item in ...]`` boilerplate.

    ``model`` is typed as ``type[Any]`` so callers can pass any pydantic
    model class without re-binding their own TypeVars; the runtime check
    below asserts it actually exposes ``model_validate``. Passing
    ``None`` leaves items as-is. Items that are not dicts
    (already-validated models, or other shapes) pass through untouched.
    """
    if model is None:
        return list(raw_items)
    validator = getattr(model, "model_validate", None)
    if validator is None:
        return list(raw_items)
    return [validator(item) if isinstance(item, dict) else item for item in raw_items]
