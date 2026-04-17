from __future__ import annotations

from datetime import datetime
from io import IOBase
from pathlib import Path
from typing import Any

import PIL.Image
from pydantic import HttpUrl

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.edits import (
    Edit,
    EditConfig,
    EditTemplate,
    EditTemplateRequest,
    FormField,
    UpdateEditTemplateRequest,
)
from ....types.mime import MIMEData
from ....types.pagination import PaginatedList, PaginationOrder
from ....types.standards import UNSET, PreparedRequest, _Unset
from ....utils.mime import prepare_mime_document


class EditTemplatesMixin:
    def _prepare_create(
        self,
        name: str,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        form_fields: list[FormField] | list[dict[str, Any]],
    ) -> PreparedRequest:
        mime_document = prepare_mime_document(document)
        field_objects = [FormField(**f) if isinstance(f, dict) else f for f in form_fields]
        payload = EditTemplateRequest(
            name=name,
            document=mime_document,
            form_fields=field_objects,
        )
        return PreparedRequest(
            method="POST",
            url="/edits/templates",
            data=payload.model_dump(mode="json", exclude_unset=True),
        )

    def _prepare_get(self, template_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/edits/templates/{template_id}")

    def _prepare_list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: PaginationOrder = "desc",
        name: str | None = None,
    ) -> PreparedRequest:
        params = {
            "before": before,
            "after": after,
            "limit": limit,
            "order": order,
            "name": name,
        }
        params = {k: v for k, v in params.items() if v is not None}
        return PreparedRequest(method="GET", url="/edits/templates", params=params)

    def _prepare_update(
        self,
        template_id: str,
        name: str | None | _Unset = UNSET,
        form_fields: list[FormField] | list[dict[str, Any]] | None | _Unset = UNSET,
    ) -> PreparedRequest:
        update_dict: dict[str, Any] = {}
        if name is not UNSET:
            update_dict["name"] = name
        if form_fields is not UNSET:
            update_dict["form_fields"] = (
                None
                if form_fields is None
                else [FormField(**f) if isinstance(f, dict) else f for f in form_fields]
            )
        payload = UpdateEditTemplateRequest(**update_dict)
        return PreparedRequest(
            method="PATCH",
            url=f"/edits/templates/{template_id}",
            data=payload.model_dump(mode="json", exclude_unset=True),
        )

    def _prepare_delete(self, template_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/edits/templates/{template_id}")

    def _prepare_fill(
        self,
        template_id: str,
        instructions: str,
        model: str | _Unset = UNSET,
        color: str | _Unset = UNSET,
        bust_cache: bool = False,
        **extra_body: Any,
    ) -> PreparedRequest:
        body: dict[str, Any] = {
            "template_id": template_id,
            "instructions": instructions,
        }
        if model is not UNSET:
            body["model"] = model
        if color is not UNSET:
            body["config"] = EditConfig(color=color).model_dump(mode="json")
        if bust_cache:
            body["bust_cache"] = True
        if extra_body:
            body.update(extra_body)
        return PreparedRequest(method="POST", url="/edits/templates/fill", data=body)


class EditTemplates(SyncAPIResource, EditTemplatesMixin):
    def create(
        self,
        name: str,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        form_fields: list[FormField] | list[dict[str, Any]],
    ) -> EditTemplate:
        request = self._prepare_create(name=name, document=document, form_fields=form_fields)
        return EditTemplate.model_validate(self._client._prepared_request(request))

    def get(self, template_id: str) -> EditTemplate:
        request = self._prepare_get(template_id)
        return EditTemplate.model_validate(self._client._prepared_request(request))

    def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: PaginationOrder = "desc",
        name: str | None = None,
    ) -> PaginatedList:
        request = self._prepare_list(before=before, after=after, limit=limit, order=order, name=name)
        response = self._client._prepared_request(request)
        result = PaginatedList(**response)

        def fetch_next(after: str) -> PaginatedList:
            return self.list(before=before, after=after, limit=limit, order=order, name=name)

        result._fetch_next_page = fetch_next
        return result

    def update(
        self,
        template_id: str,
        name: str | None | _Unset = UNSET,
        form_fields: list[FormField] | list[dict[str, Any]] | None | _Unset = UNSET,
    ) -> EditTemplate:
        request = self._prepare_update(template_id=template_id, name=name, form_fields=form_fields)
        return EditTemplate.model_validate(self._client._prepared_request(request))

    def delete(self, template_id: str) -> None:
        request = self._prepare_delete(template_id)
        self._client._prepared_request(request)

    def fill(
        self,
        template_id: str,
        instructions: str,
        model: str | _Unset = UNSET,
        color: str | _Unset = UNSET,
        bust_cache: bool = False,
        **extra_body: Any,
    ) -> Edit:
        request = self._prepare_fill(
            template_id=template_id,
            instructions=instructions,
            model=model,
            color=color,
            bust_cache=bust_cache,
            **extra_body,
        )
        return Edit.model_validate(self._client._prepared_request(request))


class AsyncEditTemplates(AsyncAPIResource, EditTemplatesMixin):
    async def create(
        self,
        name: str,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        form_fields: list[FormField] | list[dict[str, Any]],
    ) -> EditTemplate:
        request = self._prepare_create(name=name, document=document, form_fields=form_fields)
        return EditTemplate.model_validate(await self._client._prepared_request(request))

    async def get(self, template_id: str) -> EditTemplate:
        request = self._prepare_get(template_id)
        return EditTemplate.model_validate(await self._client._prepared_request(request))

    async def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: PaginationOrder = "desc",
        name: str | None = None,
    ) -> PaginatedList:
        request = self._prepare_list(before=before, after=after, limit=limit, order=order, name=name)
        response = await self._client._prepared_request(request)
        return PaginatedList(**response)

    async def update(
        self,
        template_id: str,
        name: str | None | _Unset = UNSET,
        form_fields: list[FormField] | list[dict[str, Any]] | None | _Unset = UNSET,
    ) -> EditTemplate:
        request = self._prepare_update(template_id=template_id, name=name, form_fields=form_fields)
        return EditTemplate.model_validate(await self._client._prepared_request(request))

    async def delete(self, template_id: str) -> None:
        request = self._prepare_delete(template_id)
        await self._client._prepared_request(request)

    async def fill(
        self,
        template_id: str,
        instructions: str,
        model: str | _Unset = UNSET,
        color: str | _Unset = UNSET,
        bust_cache: bool = False,
        **extra_body: Any,
    ) -> Edit:
        request = self._prepare_fill(
            template_id=template_id,
            instructions=instructions,
            model=model,
            color=color,
            bust_cache=bust_cache,
            **extra_body,
        )
        return Edit.model_validate(await self._client._prepared_request(request))
