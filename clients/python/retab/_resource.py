"""Base classes for resource clients.

Provides `SyncAPIResource` and `AsyncAPIResource`. Every resource (e.g.
`Extractions`, `WorkflowRuns`) inherits from one of these.

Includes `request_page` / `arequest_page` — the canonical helpers for
list endpoints. They run a `PreparedRequest`, re-validate items against
the typed model, and return a page (`PaginatedList[T]` /
`AsyncPaginatedList[T]`) with the next-page closure already wired.

Modelled on WorkOS's `request_page` pattern: centralising closure
construction here means every resource's `.list()` collapses to one
line and the async side can't silently forget to wire pagination.

The optional ``transform`` callback lets resources apply client-side
post-processing (e.g. ``WorkflowSteps.list(block_ids=...)``) to every
page, not just the first one. Without it, auto-paging would silently
return UNFILTERED data on subsequent pages because the filter wouldn't
be in the request's ``params`` and the closure would re-fetch the
unfiltered server response.
"""

from __future__ import annotations

import asyncio
import time
from typing import TYPE_CHECKING, Any, Callable, Type, TypeVar

from .types.pagination import (
    AsyncPaginatedList,
    ListMetadata,
    PaginatedList,
    _validate_page_items,
)
from .types.standards import PreparedRequest


if TYPE_CHECKING:
    from .client import AsyncRetab, Retab


T = TypeVar("T")
PageTransform = Callable[[list[Any]], list[Any]]


def _swap_after_cursor(params: dict[str, Any] | None, after: str) -> dict[str, Any]:
    """Return the request's params dict with ``after`` set and ``before`` dropped.

    Forward-paging always uses ``after``; carrying a stale ``before`` from
    the caller's first page would force the API to interpret the next
    request as a backward jump.
    """
    next_params: dict[str, Any] = dict(params or {})
    next_params["after"] = after
    next_params.pop("before", None)
    return next_params


def _build_next_request(request: PreparedRequest, after: str) -> PreparedRequest:
    return PreparedRequest(
        method=request.method,
        url=request.url,
        data=request.data,
        params=_swap_after_cursor(request.params, after),
        form_data=request.form_data,
        files=request.files,
        idempotency_key=None,  # next page is a fresh request
        raise_for_status=request.raise_for_status,
    )


def _apply_transform(items: list[Any], transform: PageTransform | None) -> list[Any]:
    if transform is None:
        return items
    return list(transform(items))


def _new_sync_page(
    *,
    response: Any,
    model: Type[T],
    client: "Retab",
    request: PreparedRequest,
    transform: PageTransform | None = None,
) -> PaginatedList[T]:
    """Build a `PaginatedList[T]` from a wire response and wire the closure.

    The closure captures `client`, `request`, `model`, AND `transform`,
    then re-invokes this helper for each subsequent page with the
    ``after`` cursor swapped. So filter params, ordering, limits, AND
    any client-side post-processing remain consistent across pages.
    """
    raw = response if isinstance(response, dict) else {}
    items = _apply_transform(_validate_page_items(raw.get("data") or [], model), transform)
    meta_dict = raw.get("list_metadata") or {"before": None, "after": None}
    page: PaginatedList[T] = PaginatedList(
        data=items,
        list_metadata=ListMetadata(**meta_dict),
    )

    def _fetch_next(after: str) -> PaginatedList[T]:
        next_request = _build_next_request(request, after)
        next_response = client._prepared_request(next_request)
        return _new_sync_page(
            response=next_response,
            model=model,
            client=client,
            request=next_request,
            transform=transform,
        )

    page._fetch_next_page = _fetch_next
    return page


async def _new_async_page(
    *,
    response: Any,
    model: Type[T],
    client: "AsyncRetab",
    request: PreparedRequest,
    transform: PageTransform | None = None,
) -> AsyncPaginatedList[T]:
    """Build an `AsyncPaginatedList[T]` from a wire response and wire the closure."""
    raw = response if isinstance(response, dict) else {}
    items = _apply_transform(_validate_page_items(raw.get("data") or [], model), transform)
    meta_dict = raw.get("list_metadata") or {"before": None, "after": None}
    page: AsyncPaginatedList[T] = AsyncPaginatedList(
        data=items,
        list_metadata=ListMetadata(**meta_dict),
    )

    async def _fetch_next(after: str) -> AsyncPaginatedList[T]:
        next_request = _build_next_request(request, after)
        next_response = await client._prepared_request(next_request)
        return await _new_async_page(
            response=next_response,
            model=model,
            client=client,
            request=next_request,
            transform=transform,
        )

    page._fetch_next_page = _fetch_next
    return page


class SyncAPIResource:
    _client: "Retab"

    def __init__(self, client: "Retab") -> None:
        self._client = client

    def _sleep(self, seconds: float) -> None:
        time.sleep(seconds)

    def request_page(
        self,
        request: PreparedRequest,
        *,
        model: Type[T],
        transform: PageTransform | None = None,
    ) -> PaginatedList[T]:
        """Run a list-style `PreparedRequest` and return a wired-up page.

        ``model.model_validate(item)`` is called on every dict item in the
        response's ``data`` array (the helper safely passes through
        already-validated items). The returned ``PaginatedList[T]`` has
        its ``_fetch_next_page`` closure already attached, so callers
        get transparent auto-pagination via ``for item in page:``.

        Pass ``transform`` when the resource applies client-side
        post-processing (e.g. ``WorkflowSteps.list(block_ids=...)``).
        The callback receives the validated items list and returns a new
        list; it's applied to EVERY page (initial + each closure call)
        so ``auto_paging_iter`` stays consistent.
        """
        response = self._client._prepared_request(request)
        return _new_sync_page(
            response=response,
            model=model,
            client=self._client,
            request=request,
            transform=transform,
        )


class AsyncAPIResource:
    _client: "AsyncRetab"

    def __init__(self, client: "AsyncRetab") -> None:
        self._client = client

    async def _sleep(self, seconds: float) -> None:
        await asyncio.sleep(seconds)

    async def request_page(
        self,
        request: PreparedRequest,
        *,
        model: Type[T],
        transform: PageTransform | None = None,
    ) -> AsyncPaginatedList[T]:
        """Async variant of ``SyncAPIResource.request_page``.

        Returns an ``AsyncPaginatedList[T]`` whose closure awaits each
        subsequent page, so callers can use ``async for item in page:``.
        Honours ``transform`` the same way as the sync version.
        """
        response = await self._client._prepared_request(request)
        return await _new_async_page(
            response=response,
            model=model,
            client=self._client,
            request=request,
            transform=transform,
        )
