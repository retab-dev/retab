"""Regression tests for async auto-pagination on `.list()` endpoints.

Before this fix, only the sync `Resource.list()` methods wired
`_fetch_next_page`, so the async ones could iterate the current page but
could not walk past it. `async for item in result:` either silently
returned only page 1 or raised because `__aiter__` was missing.

These tests pin the async-side contract: the resource's async `.list()`
must return an `AsyncPaginatedList[T]` that:

* yields all items across pages via `auto_paging_iter()`
* supports `async for item in page:` directly
* re-calls the same prepared-request closure with `after=` swapped to the
  previous page's terminal cursor
"""

from __future__ import annotations

from typing import Any
from unittest.mock import AsyncMock, MagicMock

import pytest

from retab.resources.extractions import AsyncExtractions
from retab.types.extractions import Extraction
from retab.types.pagination import AsyncPaginatedList

# Whole module is creditless (live list/get/pagination only; no credits).
pytestmark = pytest.mark.creditless


_FILE_REF = {"id": "file_abc", "filename": "doc.pdf", "mime_type": "application/pdf"}


def _extraction(id_: str) -> dict[str, Any]:
    return {
        "id": id_,
        "file": _FILE_REF,
        "model": "retab-small",
        "json_schema": {"type": "object", "properties": {}},
        "output": {"total": 100},
    }


def _page(items: list[dict[str, Any]], *, after: str | None) -> dict[str, Any]:
    return {"data": items, "list_metadata": {"before": None, "after": after}}


def _mock_async_client(*, prepared_request: AsyncMock) -> MagicMock:
    """Mock only the network primitive (``_prepared_request``) on the
    client. The page helper now lives on ``AsyncAPIResource`` itself,
    so the resource → page-helper → network chain stays end-to-end.
    """
    client = MagicMock()
    client._prepared_request = prepared_request
    return client


@pytest.mark.asyncio
async def test_async_list_returns_async_paginated_list() -> None:
    """The async resource must return an `AsyncPaginatedList`, distinct
    from the sync `PaginatedList`. This is the type-level signal that the
    fetch closure is awaitable.
    """
    client = _mock_async_client(
        prepared_request=AsyncMock(
            return_value=_page([_extraction("ext_1")], after=None),
        ),
    )

    result = await AsyncExtractions(client=client).list()

    assert isinstance(result, AsyncPaginatedList)
    assert len(result) == 1
    assert isinstance(result[0], Extraction)


@pytest.mark.asyncio
async def test_async_auto_pagination_walks_to_second_page() -> None:
    """`async for item in page:` must yield items from page 1 AND page 2
    (and stop on the page with `after is None`). This is the core bug:
    before the fix, the closure simply wasn't wired on the async side.
    """
    # Page 1 ends with after=ext_2; page 2 ends with after=None.
    client = _mock_async_client(
        prepared_request=AsyncMock(
            side_effect=[
                _page(
                    [_extraction("ext_1"), _extraction("ext_2")],
                    after="ext_2",
                ),
                _page(
                    [_extraction("ext_3"), _extraction("ext_4")],
                    after=None,
                ),
            ]
        ),
    )

    page = await AsyncExtractions(client=client).list()

    collected = [item async for item in page]

    assert [item.id for item in collected] == ["ext_1", "ext_2", "ext_3", "ext_4"]
    # The closure must have been called twice — once for the initial page,
    # once for the follow-up with after=ext_2.
    assert client._prepared_request.await_count == 2
    second_request = client._prepared_request.await_args_list[1].args[0]
    assert second_request.params.get("after") == "ext_2"


@pytest.mark.asyncio
async def test_async_auto_paging_iter_explicit_helper() -> None:
    """`page.auto_paging_iter()` should be callable explicitly (mirrors
    the sync helper) and yield items the same way as `__aiter__`.
    """
    client = _mock_async_client(
        prepared_request=AsyncMock(
            side_effect=[
                _page([_extraction("ext_1")], after="ext_1"),
                _page([_extraction("ext_2")], after=None),
            ]
        ),
    )

    page = await AsyncExtractions(client=client).list()

    collected = []
    async for item in page.auto_paging_iter():
        collected.append(item.id)

    assert collected == ["ext_1", "ext_2"]
