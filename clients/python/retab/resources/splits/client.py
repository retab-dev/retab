from __future__ import annotations

from datetime import datetime
from io import IOBase
from pathlib import Path
from typing import Any

import PIL.Image
from pydantic import HttpUrl

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.mime import MIMEData
from ...types.pagination import PaginatedList, PaginationOrder
from ...types.splits import Split, SplitRequest, Subdocument
from ...types.standards import UNSET, PreparedRequest, _Unset
from ...utils.mime import prepare_mime_document


class SplitsMixin:
    def _prepare_create(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        subdocuments: list[Subdocument] | list[dict[str, Any]],
        model: str,
        n_consensus: int | _Unset = UNSET,
        bust_cache: bool = False,
        **extra_body: Any,
    ) -> PreparedRequest:
        request_dict: dict[str, Any] = {
            "document": prepare_mime_document(document),
            "subdocuments": [
                Subdocument(**subdocument) if isinstance(subdocument, dict) else subdocument
                for subdocument in subdocuments
            ],
            "model": model,
        }
        if n_consensus is not UNSET:
            request_dict["n_consensus"] = n_consensus
        if bust_cache:
            request_dict["bust_cache"] = True
        if extra_body:
            request_dict.update(extra_body)

        split_request = SplitRequest(**request_dict)
        return PreparedRequest(
            method="POST",
            url="/splits",
            data=split_request.model_dump(mode="json", exclude_unset=True),
        )

    def _prepare_get(self, split_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/splits/{split_id}")

    def _prepare_list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: PaginationOrder = "desc",
        from_date: datetime | None = None,
        to_date: datetime | None = None,
    ) -> PreparedRequest:
        params = {
            "before": before,
            "after": after,
            "limit": limit,
            "order": order,
            "from_date": from_date.isoformat() if from_date else None,
            "to_date": to_date.isoformat() if to_date else None,
        }
        params = {key: value for key, value in params.items() if value is not None}
        return PreparedRequest(method="GET", url="/splits", params=params)

    def _prepare_delete(self, split_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/splits/{split_id}")


class Splits(SyncAPIResource, SplitsMixin):
    def create(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        subdocuments: list[Subdocument] | list[dict[str, Any]],
        model: str,
        n_consensus: int | _Unset = UNSET,
        bust_cache: bool = False,
        **extra_body: Any,
    ) -> Split:
        request = self._prepare_create(
            document=document,
            subdocuments=subdocuments,
            model=model,
            n_consensus=n_consensus,
            bust_cache=bust_cache,
            **extra_body,
        )
        return Split.model_validate(self._client._prepared_request(request))

    def get(self, split_id: str) -> Split:
        request = self._prepare_get(split_id)
        return Split.model_validate(self._client._prepared_request(request))

    def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: PaginationOrder = "desc",
        from_date: datetime | None = None,
        to_date: datetime | None = None,
    ) -> PaginatedList:
        request = self._prepare_list(
            before=before,
            after=after,
            limit=limit,
            order=order,
            from_date=from_date,
            to_date=to_date,
        )
        response = self._client._prepared_request(request)
        result = PaginatedList(**response)

        def fetch_next(after_cursor: str) -> PaginatedList:
            return self.list(
                before=before,
                after=after_cursor,
                limit=limit,
                order=order,
                from_date=from_date,
                to_date=to_date,
            )

        result._fetch_next_page = fetch_next
        return result

    def delete(self, split_id: str) -> None:
        request = self._prepare_delete(split_id)
        self._client._prepared_request(request)


class AsyncSplits(AsyncAPIResource, SplitsMixin):
    async def create(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        subdocuments: list[Subdocument] | list[dict[str, Any]],
        model: str,
        n_consensus: int | _Unset = UNSET,
        bust_cache: bool = False,
        **extra_body: Any,
    ) -> Split:
        request = self._prepare_create(
            document=document,
            subdocuments=subdocuments,
            model=model,
            n_consensus=n_consensus,
            bust_cache=bust_cache,
            **extra_body,
        )
        return Split.model_validate(await self._client._prepared_request(request))

    async def get(self, split_id: str) -> Split:
        request = self._prepare_get(split_id)
        return Split.model_validate(await self._client._prepared_request(request))

    async def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: PaginationOrder = "desc",
        from_date: datetime | None = None,
        to_date: datetime | None = None,
    ) -> PaginatedList:
        request = self._prepare_list(
            before=before,
            after=after,
            limit=limit,
            order=order,
            from_date=from_date,
            to_date=to_date,
        )
        response = await self._client._prepared_request(request)
        return PaginatedList(**response)

    async def delete(self, split_id: str) -> None:
        request = self._prepare_delete(split_id)
        await self._client._prepared_request(request)
