from typing import Any, Callable, Generic, Iterator, List, Literal, TypeVar

from pydantic import BaseModel, ConfigDict, PrivateAttr


T = TypeVar("T")


class ListMetadata(BaseModel):
    before: str | None
    after: str | None


class PaginatedList(BaseModel, Generic[T]):
    model_config = ConfigDict(arbitrary_types_allowed=True)

    data: list[T]
    list_metadata: ListMetadata

    _fetch_next_page: Callable[..., "PaginatedList[T]"] | None = PrivateAttr(default=None)

    def __iter__(self) -> Iterator[T]:  # type: ignore[override]
        return iter(self.data)

    def __len__(self) -> int:
        return len(self.data)

    def __getitem__(self, index: int) -> T:
        return self.data[index]

    @property
    def has_more(self) -> bool:
        """Whether there are more pages available after this one."""
        return self.list_metadata.after is not None

    def auto_paging_iter(self) -> Iterator[T]:
        """Iterate through all items across all pages automatically.

        Yields items from the current page, then fetches subsequent pages
        until no more are available.
        """
        page = self
        while True:
            yield from page.data
            if not page.has_more or page._fetch_next_page is None:
                break
            page = page._fetch_next_page(after=page.list_metadata.after)


type PaginationOrder = Literal["asc", "desc"]