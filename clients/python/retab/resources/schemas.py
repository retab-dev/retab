from io import IOBase
from pathlib import Path
from typing import Sequence, Any

import PIL.Image
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort

from .._resource import AsyncAPIResource, SyncAPIResource
from ..utils.mime import prepare_mime_document_list
from ..types.mime import MIMEData
from ..types.schemas.generate import GenerateSchemaRequest
from ..types.browser_canvas import BrowserCanvas
from ..types.standards import PreparedRequest


class SchemasMixin:
    def prepare_generate(
        self,
        documents: Sequence[Path | str | bytes | MIMEData | IOBase | PIL.Image.Image],
        instructions: str | None = None,
        model: str = "gpt-5-mini",
        temperature: float = 0.0,
        reasoning_effort: ChatCompletionReasoningEffort = "minimal",
        image_resolution_dpi: int = 96,
        browser_canvas: BrowserCanvas = "A4",
        **extra_body: Any,
    ) -> PreparedRequest:
        mime_documents = prepare_mime_document_list(documents)
        # Build known body and merge in any extras
        body: dict[str, Any] = {
            "documents": mime_documents,
            "instructions": instructions if instructions else None,
            "model": model,
            "temperature": temperature,
            "reasoning_effort": reasoning_effort,
            "image_resolution_dpi": image_resolution_dpi,
            "browser_canvas": browser_canvas,
        }
        if extra_body:
            body.update(extra_body)

        request = GenerateSchemaRequest(**body)
        return PreparedRequest(method="POST", url="/v1/schemas/generate", data=request.model_dump())


class Schemas(SyncAPIResource, SchemasMixin):
    
    def generate(
        self,
        documents: Sequence[Path | str | bytes | MIMEData | IOBase | PIL.Image.Image],
        instructions: str | None = None,
        model: str = "gpt-5-mini",
        temperature: float = 0.0,
        reasoning_effort: ChatCompletionReasoningEffort = "minimal",
        image_resolution_dpi: int = 96,
        browser_canvas: BrowserCanvas = "A4",
        **extra_body: Any,
    ) -> dict[str, Any]:
        """
        Generate a complete JSON schema by analyzing the provided documents.

        The generated schema can include X-Prompts for enhanced LLM interactions:
        - X-ReasoningPrompt: Creates auxiliary reasoning fields for complex data processing

        Args:
            documents: List of documents (as MIMEData) to analyze

        Returns:
            dict[str, Any]: Generated JSON schema with X-Prompts based on document analysis

        Raises:
            HTTPException if the request fails
        """
        prepared_request = self.prepare_generate(
            documents=documents,
            instructions=instructions,
            model=model,
            temperature=temperature,
            reasoning_effort=reasoning_effort,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            **extra_body,
        )
        response = self._client._prepared_request(prepared_request)
        return response

class AsyncSchemas(AsyncAPIResource, SchemasMixin):
    
    async def generate(
        self,
        documents: Sequence[Path | str | bytes | MIMEData | IOBase | PIL.Image.Image],
        instructions: str | None = None,
        model: str = "gpt-4o-mini",
        temperature: float = 0.0,
        reasoning_effort: ChatCompletionReasoningEffort = "minimal",
        image_resolution_dpi: int = 96,
        browser_canvas: BrowserCanvas = "A4",
        **extra_body: Any,
    ) -> dict[str, Any]:
        """
        Generate a complete JSON schema by analyzing the provided documents.

        The generated schema can include X-Prompts for enhanced LLM interactions:
        - X-ReasoningPrompt: Creates auxiliary reasoning fields for complex data processing

        Args:
            documents: List of documents (as MIMEData) to analyze

        Returns:
            dict[str, Any]: Generated JSON schema with X-Prompts based on document analysis

        Raises
            HTTPException if the request fails
        """
        prepared_request = self.prepare_generate(
            documents=documents,
            instructions=instructions,
            model=model,
            temperature=temperature,
            reasoning_effort=reasoning_effort,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            **extra_body,
        )
        response = await self._client._prepared_request(prepared_request)
        return response

