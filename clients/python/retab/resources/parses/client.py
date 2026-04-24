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
from ...types.parses import Parse, ParseRequest, TableParsingFormat
from ...types.standards import UNSET, PreparedRequest, _Unset
from ...utils.mime import prepare_mime_document


class ParsesMixin:
    def _prepare_create(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        model: str,
        table_parsing_format: TableParsingFormat | _Unset = UNSET,
        image_resolution_dpi: int | _Unset = UNSET,
        instructions: str | None = None,
        bust_cache: bool = False,
        **extra_body: Any,
    ) -> PreparedRequest:
        mime_document = prepare_mime_document(document)

        request_dict: dict[str, Any] = {
            "document": mime_document,
            "model": model,
        }
        if table_parsing_format is not UNSET:
            request_dict["table_parsing_format"] = table_parsing_format
        if image_resolution_dpi is not UNSET:
            request_dict["image_resolution_dpi"] = image_resolution_dpi
        if instructions is not None:
            request_dict["instructions"] = instructions
        if bust_cache:
            request_dict["bust_cache"] = True
        if extra_body:
            request_dict.update(extra_body)

        parse_request = ParseRequest(**request_dict)
        return PreparedRequest(
            method="POST",
            url="/parses",
            data=parse_request.model_dump(mode="json", exclude_unset=True),
        )

    def _prepare_get(self, parse_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/parses/{parse_id}")

    def _prepare_list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: PaginationOrder = "desc",
        filename: str | None = None,
        from_date: datetime | None = None,
        to_date: datetime | None = None,
    ) -> PreparedRequest:
        params = {
            "before": before,
            "after": after,
            "limit": limit,
            "order": order,
            "filename": filename,
            "from_date": from_date.isoformat() if from_date else None,
            "to_date": to_date.isoformat() if to_date else None,
        }
        params = {k: v for k, v in params.items() if v is not None}
        return PreparedRequest(method="GET", url="/parses", params=params)

    def _prepare_delete(self, parse_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/parses/{parse_id}")


class Parses(SyncAPIResource, ParsesMixin):

    def create(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        model: str,
        table_parsing_format: TableParsingFormat | _Unset = UNSET,
        image_resolution_dpi: int | _Unset = UNSET,
        instructions: str | None = None,
        bust_cache: bool = False,
        **extra_body: Any,
    ) -> Parse:
        request = self._prepare_create(
            document=document,
            model=model,
            table_parsing_format=table_parsing_format,
            image_resolution_dpi=image_resolution_dpi,
            instructions=instructions,
            bust_cache=bust_cache,
            **extra_body,
        )
        response = self._client._prepared_request(request)
        return Parse.model_validate(response)

    def get(self, parse_id: str) -> Parse:
        request = self._prepare_get(parse_id)
        return Parse.model_validate(self._client._prepared_request(request))

    def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: PaginationOrder = "desc",
        filename: str | None = None,
        from_date: datetime | None = None,
        to_date: datetime | None = None,
    ) -> PaginatedList:
        request = self._prepare_list(
            before=before, after=after, limit=limit, order=order,
            filename=filename, from_date=from_date, to_date=to_date,
        )
        response = self._client._prepared_request(request)
        result = PaginatedList(**response)

        def fetch_next(after: str) -> PaginatedList:
            return self.list(
                before=before, after=after, limit=limit, order=order,
                filename=filename, from_date=from_date, to_date=to_date,
            )

        result._fetch_next_page = fetch_next
        return result

    def delete(self, parse_id: str) -> None:
        request = self._prepare_delete(parse_id)
        self._client._prepared_request(request)


class AsyncParses(AsyncAPIResource, ParsesMixin):

    async def create(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        model: str,
        table_parsing_format: TableParsingFormat | _Unset = UNSET,
        image_resolution_dpi: int | _Unset = UNSET,
        instructions: str | None = None,
        bust_cache: bool = False,
        **extra_body: Any,
    ) -> Parse:
        request = self._prepare_create(
            document=document,
            model=model,
            table_parsing_format=table_parsing_format,
            image_resolution_dpi=image_resolution_dpi,
            instructions=instructions,
            bust_cache=bust_cache,
            **extra_body,
        )
        response = await self._client._prepared_request(request)
        return Parse.model_validate(response)

    async def get(self, parse_id: str) -> Parse:
        request = self._prepare_get(parse_id)
        return Parse.model_validate(await self._client._prepared_request(request))

    async def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: PaginationOrder = "desc",
        filename: str | None = None,
        from_date: datetime | None = None,
        to_date: datetime | None = None,
    ) -> PaginatedList:
        request = self._prepare_list(
            before=before, after=after, limit=limit, order=order,
            filename=filename, from_date=from_date, to_date=to_date,
        )
        response = await self._client._prepared_request(request)
        return PaginatedList(**response)

    async def delete(self, parse_id: str) -> None:
        request = self._prepare_delete(parse_id)
        await self._client._prepared_request(request)
