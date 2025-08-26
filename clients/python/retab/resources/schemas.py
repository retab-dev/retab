from io import IOBase
from pathlib import Path
from typing import Sequence, Any

import PIL.Image
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort

from .._resource import AsyncAPIResource, SyncAPIResource
from ..utils.ai_models import assert_valid_model_schema_generation
from ..utils.mime import prepare_mime_document_list
from ..types.mime import MIMEData
from ..types.modalities import Modality
from ..types.schemas.generate import GenerateSchemaRequest
from ..types.standards import PreparedRequest


class SchemasMixin:
    def prepare_generate(
        self,
        documents: Sequence[Path | str | bytes | MIMEData | IOBase | PIL.Image.Image],
        instructions: str | None = None,
        model: str = "gpt-5-mini",
        temperature: float = 0,
        modality: Modality = "native",
        reasoning_effort: ChatCompletionReasoningEffort = "minimal",
    ) -> PreparedRequest:
        assert_valid_model_schema_generation(model)
        mime_documents = prepare_mime_document_list(documents)
        request = GenerateSchemaRequest(
            documents=mime_documents,
            instructions=instructions if instructions else None,
            model=model,
            temperature=temperature,
            modality=modality,
            reasoning_effort=reasoning_effort,
        )
        return PreparedRequest(method="POST", url="/v1/schemas/generate", data=request.model_dump())


class Schemas(SyncAPIResource, SchemasMixin):
    
    def generate(
        self,
        documents: Sequence[Path | str | bytes | MIMEData | IOBase | PIL.Image.Image],
        instructions: str | None = None,
        model: str = "gpt-5-mini",
        temperature: float = 0,
        modality: Modality = "native",
    ) -> dict[str, Any]:
        """
        Generate a complete JSON schema by analyzing the provided documents.

        The generated schema includes X-Prompts for enhanced LLM interactions:
        - X-SystemPrompt: Defines high-level instructions and context for consistent LLM behavior
        - X-ReasoningPrompt: Creates auxiliary reasoning fields for complex data processing

        Args:
            documents: List of documents (as MIMEData) to analyze

        Returns:
            dict[str, Any]: Generated JSON schema with X-Prompts based on document analysis

        Raises:
            HTTPException if the request fails
        """

        prepared_request = self.prepare_generate(documents, instructions, model, temperature, modality)
        response = self._client._prepared_request(prepared_request)
        return response

class AsyncSchemas(AsyncAPIResource, SchemasMixin):
    
    async def generate(
        self,
        documents: Sequence[Path | str | bytes | MIMEData | IOBase | PIL.Image.Image],
        instructions: str | None = None,
        model: str = "gpt-5-mini",
        temperature: float = 0.0,
        modality: Modality = "native",
    ) -> dict[str, Any]:
        """
        Generate a complete JSON schema by analyzing the provided documents.

        The generated schema includes X-Prompts for enhanced LLM interactions:
        - X-SystemPrompt: Defines high-level instructions and context for consistent LLM behavior
        - X-FieldPrompt: Enhances standard description fields with specific extraction guidance
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
            modality=modality,
        )
        response = await self._client._prepared_request(prepared_request)
        return response

