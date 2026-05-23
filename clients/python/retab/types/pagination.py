from typing import Any, AsyncIterator, Awaitable, Callable, Generic, Iterator, Literal, TypeAlias, TypeVar

from pydantic import ConfigDict, PrivateAttr
from retab.types.base import RetabBaseModel


T = TypeVar("T")


class ListMetadata(RetabBaseModel):
    """Boundary resource IDs for page navigation."""

    before: str | None
    after: str | None


class PaginatedList(RetabBaseModel, Generic[T]):
    """Sync paginated envelope."""

    model_config = ConfigDict(arbitrary_types_allowed=True)

    data: list[T]
    list_metadata: ListMetadata

    _fetch_next_page: Callable[..., "PaginatedList[T]"] | None = PrivateAttr(default=None)

    def __iter__(self) -> Iterator[T]:  # type: ignore[override]
        return self.auto_paging_iter()

    def __len__(self) -> int:
        return len(self.data)

    def __getitem__(self, index: int) -> T:
        return self.data[index]

    @property
    def has_more(self) -> bool:
        return self.list_metadata.after is not None

    def auto_paging_iter(self) -> Iterator[T]:
        page: PaginatedList[T] = self
        while True:
            yield from page.data
            if not page.has_more or page._fetch_next_page is None:
                break
            page = page._fetch_next_page(after=page.list_metadata.after)


class AsyncPaginatedList(RetabBaseModel, Generic[T]):
    """Async paginated envelope."""

    model_config = ConfigDict(arbitrary_types_allowed=True)

    data: list[T]
    list_metadata: ListMetadata

    _fetch_next_page: Callable[..., Awaitable["AsyncPaginatedList[T]"]] | None = PrivateAttr(default=None)

    def __aiter__(self) -> AsyncIterator[T]:
        return self.auto_paging_iter()

    def __len__(self) -> int:
        return len(self.data)

    def __getitem__(self, index: int) -> T:
        return self.data[index]

    @property
    def has_more(self) -> bool:
        return self.list_metadata.after is not None

    async def auto_paging_iter(self) -> AsyncIterator[T]:
        page: AsyncPaginatedList[T] = self
        while True:
            for item in page.data:
                yield item
            if not page.has_more or page._fetch_next_page is None:
                break
            page = await page._fetch_next_page(after=page.list_metadata.after)


PaginationOrder: TypeAlias = Literal["asc", "desc"]


def _validate_page_items(raw_items: Any, model: type[Any]) -> list[Any]:
    validator = getattr(model, "model_validate", None)
    if validator is None:
        return list(raw_items)
    return [validator(item) if isinstance(item, dict) else item for item in raw_items]
