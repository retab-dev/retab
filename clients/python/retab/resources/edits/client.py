from __future__ import annotations

from datetime import datetime
from io import IOBase
from pathlib import Path
from typing import Any

import PIL.Image
from pydantic import HttpUrl

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.edits import Edit, EditConfig, EditRequest
from ...types.mime import MIMEData
from ...types.pagination import PaginatedList, PaginationOrder
from ...types.standards import UNSET, PreparedRequest, _Unset
from ...utils.mime import prepare_mime_document
from .templates import AsyncEditTemplates, EditTemplates


class EditsMixin:
    def _prepare_create(
        self,
        instructions: str,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        template_id: str | None = None,
        model: str | _Unset = UNSET,
        color: str | _Unset = UNSET,
        bust_cache: bool = False,
        **extra_body: Any,
    ) -> PreparedRequest:
        request_dict: dict[str, Any] = {"instructions": instructions}

        if document is not None:
            request_dict["document"] = prepare_mime_document(document)
        if template_id is not None:
            request_dict["template_id"] = template_id
        if model is not UNSET:
            request_dict["model"] = model
        if color is not UNSET:
            request_dict["config"] = EditConfig(color=color)
        if bust_cache:
            request_dict["bust_cache"] = True
        if extra_body:
            request_dict.update(extra_body)

        edit_request = EditRequest(**request_dict)
        return PreparedRequest(
            method="POST",
            url="/edits",
            data=edit_request.model_dump(mode="json", exclude_unset=True),
        )

    def _prepare_get(self, edit_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/edits/{edit_id}")

    def _prepare_list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: PaginationOrder = "desc",
        filename: str | None = None,
        template_id: str | None = None,
        from_date: datetime | None = None,
        to_date: datetime | None = None,
    ) -> PreparedRequest:
        params = {
            "before": before,
            "after": after,
            "limit": limit,
            "order": order,
            "filename": filename,
            "template_id": template_id,
            "from_date": from_date.isoformat() if from_date else None,
            "to_date": to_date.isoformat() if to_date else None,
        }
        params = {k: v for k, v in params.items() if v is not None}
        return PreparedRequest(method="GET", url="/edits", params=params)

    def _prepare_delete(self, edit_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/edits/{edit_id}")


class Edits(SyncAPIResource, EditsMixin):
    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.templates = EditTemplates(client=client)

    def create(
        self,
        instructions: str,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        template_id: str | None = None,
        model: str | _Unset = UNSET,
        color: str | _Unset = UNSET,
        bust_cache: bool = False,
        **extra_body: Any,
    ) -> Edit:
        request = self._prepare_create(
            instructions=instructions,
            document=document,
            template_id=template_id,
            model=model,
            color=color,
            bust_cache=bust_cache,
            **extra_body,
        )
        return Edit.model_validate(self._client._prepared_request(request))

    def get(self, edit_id: str) -> Edit:
        request = self._prepare_get(edit_id)
        return Edit.model_validate(self._client._prepared_request(request))

    def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: PaginationOrder = "desc",
        filename: str | None = None,
        template_id: str | None = None,
        from_date: datetime | None = None,
        to_date: datetime | None = None,
    ) -> PaginatedList:
        request = self._prepare_list(
            before=before, after=after, limit=limit, order=order,
            filename=filename, template_id=template_id,
            from_date=from_date, to_date=to_date,
        )
        response = self._client._prepared_request(request)
        result = PaginatedList(**response)

        def fetch_next(after: str) -> PaginatedList:
            return self.list(
                before=before, after=after, limit=limit, order=order,
                filename=filename, template_id=template_id,
                from_date=from_date, to_date=to_date,
            )

        result._fetch_next_page = fetch_next
        return result

    def delete(self, edit_id: str) -> None:
        request = self._prepare_delete(edit_id)
        self._client._prepared_request(request)


class AsyncEdits(AsyncAPIResource, EditsMixin):
    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.templates = AsyncEditTemplates(client=client)

    async def create(
        self,
        instructions: str,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        template_id: str | None = None,
        model: str | _Unset = UNSET,
        color: str | _Unset = UNSET,
        bust_cache: bool = False,
        **extra_body: Any,
    ) -> Edit:
        request = self._prepare_create(
            instructions=instructions,
            document=document,
            template_id=template_id,
            model=model,
            color=color,
            bust_cache=bust_cache,
            **extra_body,
        )
        return Edit.model_validate(await self._client._prepared_request(request))

    async def get(self, edit_id: str) -> Edit:
        request = self._prepare_get(edit_id)
        return Edit.model_validate(await self._client._prepared_request(request))

    async def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: PaginationOrder = "desc",
        filename: str | None = None,
        template_id: str | None = None,
        from_date: datetime | None = None,
        to_date: datetime | None = None,
    ) -> PaginatedList:
        request = self._prepare_list(
            before=before, after=after, limit=limit, order=order,
            filename=filename, template_id=template_id,
            from_date=from_date, to_date=to_date,
        )
        response = await self._client._prepared_request(request)
        return PaginatedList(**response)

    async def delete(self, edit_id: str) -> None:
        request = self._prepare_delete(edit_id)
        await self._client._prepared_request(request)
