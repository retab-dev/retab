"""
Edit Templates SDK client - Wrapper for template generation and filling.
"""

from io import IOBase
from pathlib import Path
from typing import Any

import PIL.Image
from pydantic import HttpUrl

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.documents.edit import EditConfig, EditResponse, InferFormSchemaRequest, InferFormSchemaResponse
from ....types.edit.templates import FillTemplateRequest
from ....types.mime import MIMEData
from ....types.standards import PreparedRequest, UNSET, _Unset
from ....utils.mime import prepare_mime_document


class BaseTemplatesMixin:
    """Shared methods for template generation and filling."""

    def _prepare_generate(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        model: str | _Unset = UNSET,
        instructions: str | None | _Unset = UNSET,
        **extra_body: Any,
    ) -> PreparedRequest:
        mime_document = prepare_mime_document(document)

        request_dict: dict[str, Any] = {
            "document": mime_document,
        }

        if model is not UNSET:
            request_dict["model"] = model
        if instructions is not UNSET:
            request_dict["instructions"] = instructions
        if extra_body:
            request_dict.update(extra_body)

        infer_request = InferFormSchemaRequest(**request_dict)
        return PreparedRequest(
            method="POST",
            url="/edit/templates/generate",
            data=infer_request.model_dump(mode="json", exclude_unset=True),
        )

    def _prepare_fill(
        self,
        template_id: str,
        instructions: str,
        model: str | _Unset = UNSET,
        color: str | _Unset = UNSET,
        **extra_body: Any,
    ) -> PreparedRequest:
        request_dict: dict[str, Any] = {
            "template_id": template_id,
            "instructions": instructions,
        }

        if model is not UNSET:
            request_dict["model"] = model
        if color is not UNSET:
            request_dict["config"] = EditConfig(color=color)
        if extra_body:
            request_dict.update(extra_body)

        fill_request = FillTemplateRequest(**request_dict)
        return PreparedRequest(
            method="POST",
            url="/edit/templates/fill",
            data=fill_request.model_dump(mode="json", exclude_unset=True),
        )


class Templates(SyncAPIResource, BaseTemplatesMixin):
    """Templates API wrapper for synchronous usage."""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)

    def generate(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        model: str | _Unset = UNSET,
        instructions: str | None | _Unset = UNSET,
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
        model: str | _Unset = UNSET,
        color: str | _Unset = UNSET,
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

    async def generate(
        self,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl,
        model: str | _Unset = UNSET,
        instructions: str | None | _Unset = UNSET,
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
        model: str | _Unset = UNSET,
        color: str | _Unset = UNSET,
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
