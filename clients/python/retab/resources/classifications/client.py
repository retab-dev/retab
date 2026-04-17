from __future__ import annotations

from datetime import datetime
from io import IOBase
from pathlib import Path
from typing import Any

import PIL.Image
from pydantic import HttpUrl

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.classifications import Category, Classification, ClassificationRequest
from ...types.mime import MIMEData
from ...types.pagination import PaginatedList, PaginationOrder
from ...types.standards import UNSET, PreparedRequest, _Unset
from ...utils.mime import prepare_mime_document


class ClassificationsMixin:
    def _prepare_create(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        categories: list[Category] | list[dict[str, str]],
        model: str,
        n_consensus: int | _Unset = UNSET,
        bust_cache: bool = False,
        **extra_body: Any,
    ) -> PreparedRequest:
        mime_document = prepare_mime_document(document)

        category_objects = [
            Category(**cat) if isinstance(cat, dict) else cat
            for cat in categories
        ]

        request_dict: dict[str, Any] = {
            "document": mime_document,
            "categories": category_objects,
            "model": model,
        }
        if n_consensus is not UNSET:
            request_dict["n_consensus"] = n_consensus
        if bust_cache:
            request_dict["bust_cache"] = True
        if extra_body:
            request_dict.update(extra_body)

        classify_request = ClassificationRequest(**request_dict)
        return PreparedRequest(
            method="POST",
            url="/classifications",
            data=classify_request.model_dump(mode="json", exclude_unset=True),
        )

    def _prepare_get(self, classification_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/classifications/{classification_id}")

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
        params = {k: v for k, v in params.items() if v is not None}
        return PreparedRequest(method="GET", url="/classifications", params=params)

    def _prepare_delete(self, classification_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/classifications/{classification_id}")


class Classifications(SyncAPIResource, ClassificationsMixin):

    def create(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        categories: list[Category] | list[dict[str, str]],
        model: str,
        n_consensus: int | _Unset = UNSET,
        bust_cache: bool = False,
        **extra_body: Any,
    ) -> Classification:
        request = self._prepare_create(
            document=document,
            categories=categories,
            model=model,
            n_consensus=n_consensus,
            bust_cache=bust_cache,
            **extra_body,
        )
        response = self._client._prepared_request(request)
        return Classification.model_validate(response)

    def get(self, classification_id: str) -> Classification:
        request = self._prepare_get(classification_id)
        return Classification.model_validate(self._client._prepared_request(request))

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
            before=before, after=after, limit=limit, order=order,
            from_date=from_date, to_date=to_date,
        )
        response = self._client._prepared_request(request)
        result = PaginatedList(**response)

        def fetch_next(after: str) -> PaginatedList:
            return self.list(
                before=before, after=after, limit=limit, order=order,
                from_date=from_date, to_date=to_date,
            )

        result._fetch_next_page = fetch_next
        return result

    def delete(self, classification_id: str) -> None:
        request = self._prepare_delete(classification_id)
        self._client._prepared_request(request)


class AsyncClassifications(AsyncAPIResource, ClassificationsMixin):

    async def create(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        categories: list[Category] | list[dict[str, str]],
        model: str,
        n_consensus: int | _Unset = UNSET,
        bust_cache: bool = False,
        **extra_body: Any,
    ) -> Classification:
        request = self._prepare_create(
            document=document,
            categories=categories,
            model=model,
            n_consensus=n_consensus,
            bust_cache=bust_cache,
            **extra_body,
        )
        response = await self._client._prepared_request(request)
        return Classification.model_validate(response)

    async def get(self, classification_id: str) -> Classification:
        request = self._prepare_get(classification_id)
        return Classification.model_validate(await self._client._prepared_request(request))

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
            before=before, after=after, limit=limit, order=order,
            from_date=from_date, to_date=to_date,
        )
        response = await self._client._prepared_request(request)
        return PaginatedList(**response)

    async def delete(self, classification_id: str) -> None:
        request = self._prepare_delete(classification_id)
        await self._client._prepared_request(request)
