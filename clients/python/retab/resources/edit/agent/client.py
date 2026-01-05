"""
Agent Edit SDK client - Wrapper for agent-based document editing functionality.
"""

from io import IOBase
from pathlib import Path
from typing import Any

import PIL.Image
from pydantic import HttpUrl

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....utils.mime import prepare_mime_document
from ....types.documents.edit import (
    EditConfig,
    EditRequest,
    EditResponse,
)
from ....types.mime import MIMEData
from ....types.standards import PreparedRequest, FieldUnset


class BaseAgentMixin:
    """Shared methods for preparing agent edit API requests."""

    def _prepare_fill(
        self,
        instructions: str,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        model: str = FieldUnset,
        color: str = FieldUnset,
        **extra_body: Any,
    ) -> PreparedRequest:
        request_dict: dict[str, Any] = {
            "instructions": instructions,
        }

        if document is not None:
            mime_document = prepare_mime_document(document)
            request_dict["document"] = mime_document

        if model is not FieldUnset:
            request_dict["model"] = model

        if color is not FieldUnset:
            request_dict["config"] = EditConfig(color=color)

        # Merge any extra fields provided by the caller
        if extra_body:
            request_dict.update(extra_body)

        edit_request = EditRequest(**request_dict)
        return PreparedRequest(
            method="POST",
            url="/v1/edit/agent/fill",
            data=edit_request.model_dump(mode="json", exclude_unset=True),
        )


class Agent(SyncAPIResource, BaseAgentMixin):
    """Agent Edit API wrapper for synchronous usage."""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)

    def fill(
        self,
        instructions: str,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        model: str = FieldUnset,
        color: str = FieldUnset,
        **extra_body: Any,
    ) -> EditResponse:
        """
        Edit a document by inferring form fields and filling them with provided instructions.

        This method performs:
        1. Detection to identify form field bounding boxes
        2. LLM inference to name and describe detected fields
        3. LLM-based form filling using the provided instructions
        4. Returns the filled document with form field values populated

        Args:
            instructions: Instructions describing how to fill the form fields.
            document: The document to edit. Can be a file path (Path or str), file-like object,
                MIMEData, PIL Image, or URL.
            model: The LLM model to use for inference. Defaults to "retab-small".
            color: Hex color code for filled text (e.g. "#000080"). Defaults to dark blue.

        Returns:
            EditResponse: Response containing:
                - form_data: List of form fields with filled values
                - filled_document: Document with filled form values (MIMEData)

        Raises:
            HTTPException: If the request fails.

        Supported document formats:
            - PDF: Native form field detection and filling
            - DOCX/DOC: Native editing to preserve styles and formatting
            - PPTX/PPT: Native editing for presentations
            - XLSX/XLS: Native editing for spreadsheets
        """
        request = self._prepare_fill(
            instructions=instructions,
            document=document,
            model=model,
            color=color,
            **extra_body,
        )
        response = self._client._prepared_request(request)
        return EditResponse.model_validate(response)


class AsyncAgent(AsyncAPIResource, BaseAgentMixin):
    """Agent Edit API wrapper for asynchronous usage."""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)

    async def fill(
        self,
        instructions: str,
        document: Path | str | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        model: str = FieldUnset,
        color: str = FieldUnset,
        **extra_body: Any,
    ) -> EditResponse:
        """
        Edit a document by inferring form fields and filling them with provided instructions asynchronously.

        This method performs:
        1. Detection to identify form field bounding boxes
        2. LLM inference to name and describe detected fields
        3. LLM-based form filling using the provided instructions
        4. Returns the filled document with form field values populated

        Args:
            instructions: Instructions describing how to fill the form fields.
            document: The document to edit. Can be a file path (Path or str), file-like object,
                MIMEData, PIL Image, or URL.
            model: The LLM model to use for inference. Defaults to "retab-small".
            color: Hex color code for filled text (e.g. "#000080"). Defaults to dark blue.

        Returns:
            EditResponse: Response containing:
                - form_data: List of form fields with filled values
                - filled_document: Document with filled form values (MIMEData)

        Raises:
            HTTPException: If the request fails.

        Supported document formats:
            - PDF: Native form field detection and filling
            - DOCX/DOC: Native editing to preserve styles and formatting
            - PPTX/PPT: Native editing for presentations
            - XLSX/XLS: Native editing for spreadsheets
        """
        request = self._prepare_fill(
            instructions=instructions,
            document=document,
            model=model,
            color=color,
            **extra_body,
        )
        response = await self._client._prepared_request(request)
        return EditResponse.model_validate(response)

