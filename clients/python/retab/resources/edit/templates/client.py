"""
Edit Templates SDK client - Wrapper for edit template management.
"""

from io import IOBase
from pathlib import Path
from typing import Any, Literal, List

import PIL.Image
from pydantic import HttpUrl

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....utils.mime import prepare_mime_document
from ....types.documents.edit import (
    EditConfig,
    FormField,
    InferFormSchemaRequest,
    InferFormSchemaResponse,
    EditResponse,
)
from ....types.edit.templates import EditTemplate, FillTemplateRequest
from ....types.mime import MIMEData
from ....types.standards import PreparedRequest, FieldUnset
from ....types.pagination import PaginatedList


class BaseTemplatesMixin:
    """Shared methods for preparing template API requests."""

    def _prepare_list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] = "desc",
        filename: str | None = None,
        mime_type: str | None = None,
        **extra_params: Any,
    ) -> PreparedRequest:
        params: dict[str, Any] = {
            "limit": limit,
            "order": order,
        }
        if before:
            params["before"] = before
        if after:
            params["after"] = after
        if filename:
            params["filename"] = filename
        if mime_type:
            params["mime_type"] = mime_type
        if extra_params:
            params.update(extra_params)

        return PreparedRequest(
            method="GET",
            url="/v1/edit/templates",
            params=params,
        )

    def _prepare_get(self, template_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url=f"/v1/edit/templates/{template_id}",
        )

    def _prepare_create(
        self,
        name: str,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        form_fields: list[FormField],
        **extra_body: Any,
    ) -> PreparedRequest:
        mime_document = prepare_mime_document(document)
        
        request_dict: dict[str, Any] = {
            "name": name,
            "document": mime_document,
            "form_fields": [f.model_dump() if hasattr(f, 'model_dump') else f for f in form_fields],
        }
        if extra_body:
            request_dict.update(extra_body)

        return PreparedRequest(
            method="POST",
            url="/v1/edit/templates",
            data=request_dict,
        )

    def _prepare_update(
        self,
        template_id: str,
        name: str | None = None,
        form_fields: list[FormField] | None = None,
        **extra_body: Any,
    ) -> PreparedRequest:
        request_dict: dict[str, Any] = {}
        if name is not None:
            request_dict["name"] = name
        if form_fields is not None:
            request_dict["form_fields"] = [f.model_dump() if hasattr(f, 'model_dump') else f for f in form_fields]
        if extra_body:
            request_dict.update(extra_body)

        return PreparedRequest(
            method="PATCH",
            url=f"/v1/edit/templates/{template_id}",
            data=request_dict,
        )

    def _prepare_delete(self, template_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="DELETE",
            url=f"/v1/edit/templates/{template_id}",
        )

    def _prepare_duplicate(
        self,
        template_id: str,
        name: str | None = None,
        **extra_body: Any,
    ) -> PreparedRequest:
        request_dict: dict[str, Any] = {}
        if name is not None:
            request_dict["name"] = name
        if extra_body:
            request_dict.update(extra_body)

        return PreparedRequest(
            method="POST",
            url=f"/v1/edit/templates/{template_id}/duplicate",
            data=request_dict,
        )

    def _prepare_generate(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        model: str = FieldUnset,
        instructions: str | None = FieldUnset,
        **extra_body: Any,
    ) -> PreparedRequest:
        mime_document = prepare_mime_document(document)

        request_dict: dict[str, Any] = {
            "document": mime_document,
        }

        if model is not FieldUnset:
            request_dict["model"] = model
        if instructions is not FieldUnset:
            request_dict["instructions"] = instructions
        if extra_body:
            request_dict.update(extra_body)

        infer_request = InferFormSchemaRequest(**request_dict)
        return PreparedRequest(
            method="POST",
            url="/v1/edit/templates/generate",
            data=infer_request.model_dump(mode="json", exclude_unset=True),
        )

    def _prepare_fill(
        self,
        template_id: str,
        instructions: str,
        model: str = FieldUnset,
        color: str = FieldUnset,
        **extra_body: Any,
    ) -> PreparedRequest:
        request_dict: dict[str, Any] = {
            "template_id": template_id,
            "instructions": instructions,
        }

        if model is not FieldUnset:
            request_dict["model"] = model
        if color is not FieldUnset:
            request_dict["config"] = EditConfig(color=color)
        if extra_body:
            request_dict.update(extra_body)

        fill_request = FillTemplateRequest(**request_dict)
        return PreparedRequest(
            method="POST",
            url="/v1/edit/templates/fill",
            data=fill_request.model_dump(mode="json", exclude_unset=True),
        )


class Templates(SyncAPIResource, BaseTemplatesMixin):
    """Templates API wrapper for synchronous usage."""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)

    def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] = "desc",
        filename: str | None = None,
        mime_type: str | None = None,
        **extra_params: Any,
    ) -> PaginatedList:
        """
        List edit templates with pagination and optional filtering.

        Args:
            before: Cursor for backward pagination
            after: Cursor for forward pagination
            limit: Number of items per page (1-100, default 10)
            order: Sort order ("asc" or "desc", default "desc")
            filename: Filter by filename (partial match)
            mime_type: Filter by MIME type

        Returns:
            PaginatedList: Paginated list of templates (data contains EditTemplate objects)
        """
        request = self._prepare_list(
            before=before,
            after=after,
            limit=limit,
            order=order,
            filename=filename,
            mime_type=mime_type,
            **extra_params,
        )
        response = self._client._prepared_request(request)
        return PaginatedList.model_validate(response)

    def get(self, template_id: str) -> EditTemplate:
        """
        Get a specific edit template by ID.

        Args:
            template_id: The template ID to retrieve

        Returns:
            EditTemplate: The template details
        """
        request = self._prepare_get(template_id)
        response = self._client._prepared_request(request)
        return EditTemplate.model_validate(response)

    def create(
        self,
        name: str,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        form_fields: List[FormField],
        **extra_body: Any,
    ) -> EditTemplate:
        """
        Create a new edit template.

        Args:
            name: Name of the template
            document: The PDF document to use as the template
            form_fields: List of form fields in the template

        Returns:
            EditTemplate: The created template
        """
        request = self._prepare_create(
            name=name,
            document=document,
            form_fields=form_fields,
            **extra_body,
        )
        response = self._client._prepared_request(request)
        return EditTemplate.model_validate(response)

    def update(
        self,
        template_id: str,
        name: str | None = None,
        form_fields: List[FormField] | None = None,
        **extra_body: Any,
    ) -> EditTemplate:
        """
        Update an existing edit template.

        Args:
            template_id: The template ID to update
            name: New name for the template (optional)
            form_fields: New form fields (optional)

        Returns:
            EditTemplate: The updated template
        """
        request = self._prepare_update(
            template_id=template_id,
            name=name,
            form_fields=form_fields,
            **extra_body,
        )
        response = self._client._prepared_request(request)
        return EditTemplate.model_validate(response)

    def delete(self, template_id: str) -> None:
        """
        Delete an edit template.

        Args:
            template_id: The template ID to delete
        """
        request = self._prepare_delete(template_id)
        self._client._prepared_request(request)

    def duplicate(
        self,
        template_id: str,
        name: str | None = None,
        **extra_body: Any,
    ) -> EditTemplate:
        """
        Duplicate an existing edit template.

        Args:
            template_id: The template ID to duplicate
            name: Name for the duplicated template (defaults to "<original> (copy)")

        Returns:
            EditTemplate: The duplicated template
        """
        request = self._prepare_duplicate(
            template_id=template_id,
            name=name,
            **extra_body,
        )
        response = self._client._prepared_request(request)
        return EditTemplate.model_validate(response)

    def generate(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        model: str = FieldUnset,
        instructions: str | None = FieldUnset,
        **extra_body: Any,
    ) -> InferFormSchemaResponse:
        """
        Infer form schema from a PDF document.

        This method combines computer vision for precise bounding box detection
        with LLM for semantic field naming (key, description) and type classification.

        Args:
            document: The PDF document to analyze
            model: The LLM model to use for field naming (default: "retab-small")
            instructions: Optional instructions to guide form field detection

        Returns:
            InferFormSchemaResponse: Response containing:
                - form_schema: The detected form schema
                - annotated_pdf: PDF with bounding boxes for visual verification
                - detection_count: Number of fields detected

        Note:
            Only PDF documents are supported for form schema inference.
        """
        request = self._prepare_generate(
            document=document,
            model=model,
            instructions=instructions,
            **extra_body,
        )
        response = self._client._prepared_request(request)
        return InferFormSchemaResponse.model_validate(response)

    def fill(
        self,
        template_id: str,
        instructions: str,
        model: str = FieldUnset,
        color: str = FieldUnset,
        **extra_body: Any,
    ) -> EditResponse:
        """
        Fill a PDF form using a pre-defined template.

        This method uses a template's pre-defined form fields to fill a PDF form,
        skipping the field detection step for faster processing.

        Args:
            template_id: The template ID to use for filling
            instructions: Instructions describing how to fill the form fields
            model: The LLM model to use for inference (default: "retab-small")
            color: Hex color code for filled text (e.g. "#000080"). Defaults to dark blue.

        Returns:
            EditResponse: Response containing:
                - form_data: List of form fields with filled values
                - filled_document: The filled PDF document as MIMEData

        Use cases:
            - Batch processing of the same form with different data
            - Faster form filling when field detection is already done
            - Consistent field mapping across multiple fills
        """
        request = self._prepare_fill(
            template_id=template_id,
            instructions=instructions,
            model=model,
            color=color,
            **extra_body,
        )
        response = self._client._prepared_request(request)
        return EditResponse.model_validate(response)


class AsyncTemplates(AsyncAPIResource, BaseTemplatesMixin):
    """Templates API wrapper for asynchronous usage."""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)

    async def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] = "desc",
        filename: str | None = None,
        mime_type: str | None = None,
        **extra_params: Any,
    ) -> PaginatedList:
        """
        List edit templates with pagination and optional filtering.

        Args:
            before: Cursor for backward pagination
            after: Cursor for forward pagination
            limit: Number of items per page (1-100, default 10)
            order: Sort order ("asc" or "desc", default "desc")
            filename: Filter by filename (partial match)
            mime_type: Filter by MIME type

        Returns:
            PaginatedList: Paginated list of templates (data contains EditTemplate objects)
        """
        request = self._prepare_list(
            before=before,
            after=after,
            limit=limit,
            order=order,
            filename=filename,
            mime_type=mime_type,
            **extra_params,
        )
        response = await self._client._prepared_request(request)
        return PaginatedList.model_validate(response)

    async def get(self, template_id: str) -> EditTemplate:
        """
        Get a specific edit template by ID.

        Args:
            template_id: The template ID to retrieve

        Returns:
            EditTemplate: The template details
        """
        request = self._prepare_get(template_id)
        response = await self._client._prepared_request(request)
        return EditTemplate.model_validate(response)

    async def create(
        self,
        name: str,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        form_fields: List[FormField],
        **extra_body: Any,
    ) -> EditTemplate:
        """
        Create a new edit template.

        Args:
            name: Name of the template
            document: The PDF document to use as the template
            form_fields: List of form fields in the template

        Returns:
            EditTemplate: The created template
        """
        request = self._prepare_create(
            name=name,
            document=document,
            form_fields=form_fields,
            **extra_body,
        )
        response = await self._client._prepared_request(request)
        return EditTemplate.model_validate(response)

    async def update(
        self,
        template_id: str,
        name: str | None = None,
        form_fields: List[FormField] | None = None,
        **extra_body: Any,
    ) -> EditTemplate:
        """
        Update an existing edit template.

        Args:
            template_id: The template ID to update
            name: New name for the template (optional)
            form_fields: New form fields (optional)

        Returns:
            EditTemplate: The updated template
        """
        request = self._prepare_update(
            template_id=template_id,
            name=name,
            form_fields=form_fields,
            **extra_body,
        )
        response = await self._client._prepared_request(request)
        return EditTemplate.model_validate(response)

    async def delete(self, template_id: str) -> None:
        """
        Delete an edit template.

        Args:
            template_id: The template ID to delete
        """
        request = self._prepare_delete(template_id)
        await self._client._prepared_request(request)

    async def duplicate(
        self,
        template_id: str,
        name: str | None = None,
        **extra_body: Any,
    ) -> EditTemplate:
        """
        Duplicate an existing edit template.

        Args:
            template_id: The template ID to duplicate
            name: Name for the duplicated template (defaults to "<original> (copy)")

        Returns:
            EditTemplate: The duplicated template
        """
        request = self._prepare_duplicate(
            template_id=template_id,
            name=name,
            **extra_body,
        )
        response = await self._client._prepared_request(request)
        return EditTemplate.model_validate(response)

    async def generate(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        model: str = FieldUnset,
        instructions: str | None = FieldUnset,
        **extra_body: Any,
    ) -> InferFormSchemaResponse:
        """
        Infer form schema from a PDF document asynchronously.

        This method combines computer vision for precise bounding box detection
        with LLM for semantic field naming (key, description) and type classification.

        Args:
            document: The PDF document to analyze
            model: The LLM model to use for field naming (default: "retab-small")
            instructions: Optional instructions to guide form field detection

        Returns:
            InferFormSchemaResponse: Response containing:
                - form_schema: The detected form schema
                - annotated_pdf: PDF with bounding boxes for visual verification
                - detection_count: Number of fields detected

        Note:
            Only PDF documents are supported for form schema inference.
        """
        request = self._prepare_generate(
            document=document,
            model=model,
            instructions=instructions,
            **extra_body,
        )
        response = await self._client._prepared_request(request)
        return InferFormSchemaResponse.model_validate(response)

    async def fill(
        self,
        template_id: str,
        instructions: str,
        model: str = FieldUnset,
        color: str = FieldUnset,
        **extra_body: Any,
    ) -> EditResponse:
        """
        Fill a PDF form using a pre-defined template asynchronously.

        This method uses a template's pre-defined form fields to fill a PDF form,
        skipping the field detection step for faster processing.

        Args:
            template_id: The template ID to use for filling
            instructions: Instructions describing how to fill the form fields
            model: The LLM model to use for inference (default: "retab-small")
            color: Hex color code for filled text (e.g. "#000080"). Defaults to dark blue.

        Returns:
            EditResponse: Response containing:
                - form_data: List of form fields with filled values
                - filled_document: The filled PDF document as MIMEData

        Use cases:
            - Batch processing of the same form with different data
            - Faster form filling when field detection is already done
            - Consistent field mapping across multiple fills
        """
        request = self._prepare_fill(
            template_id=template_id,
            instructions=instructions,
            model=model,
            color=color,
            **extra_body,
        )
        response = await self._client._prepared_request(request)
        return EditResponse.model_validate(response)

