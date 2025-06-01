from io import IOBase
from pathlib import Path
from typing import Any, List, Optional, Sequence

import PIL.Image
from pydantic import BaseModel

from .._resource import AsyncAPIResource, SyncAPIResource
from .._utils.ai_models import assert_valid_model_schema_generation
from .._utils.json_schema import load_json_schema
from .._utils.mime import prepare_mime_document_list
from ..types.modalities import Modality
from ..types.mime import MIMEData
from ..types.schemas.generate import GenerateSchemaRequest
from ..types.schemas.enhance import EnhanceSchemaConfig, EnhanceSchemaConfigDict, EnhanceSchemaRequest
from ..types.schemas.evaluate import EvaluateSchemaRequest, EvaluateSchemaResponse
from ..types.schemas.object import Schema
from ..types.standards import PreparedRequest
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort


class SchemasMixin:
    def prepare_generate(
        self,
        documents: Sequence[Path | str | bytes | MIMEData | IOBase | PIL.Image.Image],
        instructions: str | None = None,
        model: str = "gpt-4o-2024-11-20",
        temperature: float = 0,
        modality: Modality = "native",
    ) -> PreparedRequest:
        assert_valid_model_schema_generation(model)
        mime_documents = prepare_mime_document_list(documents)
        data = {
            "documents": [doc.model_dump() for doc in mime_documents],
            "instructions": instructions if instructions else None,
            "model": model,
            "temperature": temperature,
            "modality": modality,
        }
        GenerateSchemaRequest.model_validate(data)
        return PreparedRequest(method="POST", url="/v1/schemas/generate", data=data)

    def prepare_evaluate(
        self,
        documents: Sequence[Path | str | bytes | MIMEData | IOBase | PIL.Image.Image],
        json_schema: dict[str, Any],
        ground_truths: list[dict[str, Any]] | None = None,
        model: str = "gpt-4o-mini",
        temperature: float = 0.0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        modality: Modality = "native",
        image_resolution_dpi: int = 96,
        browser_canvas: str = "A4",
        n_consensus: int = 1,
    ) -> PreparedRequest:
        # Assert that if ground_truths is not None, it has the same length as documents
        if ground_truths is not None and len(documents) != len(ground_truths):
            raise ValueError("Number of documents must match number of ground truths")

        mime_documents = prepare_mime_document_list(documents)
        data = {
            "documents": [doc.model_dump() for doc in mime_documents],
            "ground_truths": ground_truths,
            "model": model,
            "temperature": temperature,
            "reasoning_effort": reasoning_effort,
            "modality": modality,
            "image_resolution_dpi": image_resolution_dpi,
            "browser_canvas": browser_canvas,
            "n_consensus": n_consensus,
            "json_schema": json_schema,
        }
        EvaluateSchemaRequest.model_validate(data)
        return PreparedRequest(method="POST", url="/v1/schemas/evaluate", data=data)

    def prepare_enhance(
        self,
        json_schema: dict[str, Any] | Path | str,
        documents: Sequence[Path | str | bytes | MIMEData | IOBase | PIL.Image.Image],
        ground_truths: list[dict[str, Any]] | None,
        instructions: str | None,
        model: str,
        temperature: float,
        modality: Modality,
        flat_likelihoods: list[dict[str, float]] | dict[str, float] | None,
        tools_config: EnhanceSchemaConfig,
    ) -> PreparedRequest:
        assert_valid_model_schema_generation(model)
        mime_documents = prepare_mime_document_list(documents)
        loaded_json_schema = load_json_schema(json_schema)
        data = {
            "json_schema": loaded_json_schema,
            "documents": [doc.model_dump() for doc in mime_documents],
            "ground_truths": ground_truths,
            "instructions": instructions if instructions else None,
            "model": model,
            "temperature": temperature,
            "modality": modality,
            "flat_likelihoods": flat_likelihoods,
            "tools_config": tools_config.model_dump(),
        }
        EnhanceSchemaRequest.model_validate(data)
        return PreparedRequest(method="POST", url="/v1/schemas/enhance", data=data)

    def prepare_get(self, schema_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/schemas/{schema_id}")


class Schemas(SyncAPIResource, SchemasMixin):
    def load(self, json_schema: dict[str, Any] | Path | str | None = None, pydantic_model: type[BaseModel] | None = None) -> Schema:
        """Load a schema from a JSON schema.

        Args:
            json_schema: The JSON schema to load
        """
        if json_schema:
            return Schema(json_schema=load_json_schema(json_schema))
        elif pydantic_model:
            return Schema(pydantic_model=pydantic_model)
        else:
            raise ValueError("Either json_schema or pydantic_model must be provided")

    def generate(
        self,
        documents: Sequence[Path | str | bytes | MIMEData | IOBase | PIL.Image.Image],
        instructions: str | None = None,
        model: str = "gpt-4o-2024-11-20",
        temperature: float = 0,
        modality: Modality = "native",
    ) -> Schema:
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
        return Schema.model_validate(response)

    def evaluate(
        self,
        documents: Sequence[Path | str | bytes | MIMEData | IOBase | PIL.Image.Image],
        json_schema: dict[str, Any],
        ground_truths: list[dict[str, Any]] | None = None,
        model: str = "gpt-4o-mini",
        temperature: float = 0.0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        modality: Modality = "native",
        image_resolution_dpi: int = 96,
        browser_canvas: str = "A4",
        n_consensus: int = 1,
    ) -> EvaluateSchemaResponse:
        """
        Evaluate a schema by performing extractions on provided documents.
        If ground truths are provided, compares extractions against them.
        Otherwise, uses consensus likelihoods as metrics.

        Args:
            documents: List of documents to evaluate against
            json_schema: The JSON schema to evaluate
            ground_truths: Optional list of ground truth dictionaries to compare against
            model: The model to use for extraction
            temperature: The temperature to use for extraction
            reasoning_effort: The reasoning effort to use for extraction
            modality: The modality to use for extraction
            image_resolution_dpi: The DPI of the image. Defaults to 96.
            browser_canvas: The canvas size of the browser. Must be one of:
                - "A3" (11.7in x 16.54in)
                - "A4" (8.27in x 11.7in) 
                - "A5" (5.83in x 8.27in)
                Defaults to "A4".
            n_consensus: Number of consensus rounds to perform

        Returns:
            EvaluateSchemaResponse: Dictionary containing evaluation metrics for each document.
            Each metric includes:
            - similarity: Average similarity/likelihood score
            - similarities: Per-field similarity/likelihood scores
            - flat_similarities: Flattened per-field similarity/likelihood scores
            - aligned_* versions of the above metrics

        Raises:
            ValueError: If ground_truths is provided and its length doesn't match documents
            HTTPException: If the request fails or if there are too many documents
        """
        prepared_request = self.prepare_evaluate(
            documents=documents,
            json_schema=json_schema,
            ground_truths=ground_truths,
            model=model,
            temperature=temperature,
            reasoning_effort=reasoning_effort,
            modality=modality,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            n_consensus=n_consensus,
        )
        response = self._client._prepared_request(prepared_request)
        return EvaluateSchemaResponse.model_validate(response)

    def enhance(
        self,
        json_schema: dict[str, Any] | Path | str,
        documents: Sequence[Path | str | bytes | MIMEData | IOBase | PIL.Image.Image],
        ground_truths: list[dict[str, Any]] | None = None,
        instructions: str | None = None,
        model: str = "gpt-4o-2024-11-20",
        temperature: float = 0,
        modality: Modality = "native",
        flat_likelihoods: list[dict[str, float]] | dict[str, float] | None = None,
        tools_config: EnhanceSchemaConfigDict | None = None,
    ) -> Schema:
        prepared_request = self.prepare_enhance(
            json_schema, documents, ground_truths, instructions, model, temperature, modality, flat_likelihoods, EnhanceSchemaConfig.model_validate(tools_config or {})
        )
        response = self._client._prepared_request(prepared_request)
        return Schema(json_schema=response["json_schema"])


class AsyncSchemas(AsyncAPIResource, SchemasMixin):
    async def load(self, json_schema: dict[str, Any] | Path | str | None = None, pydantic_model: type[BaseModel] | None = None) -> Schema:
        """Load a schema from a JSON schema.

        Args:
            json_schema: The JSON schema to load
            pydantic_model: The Pydantic model to load
        """
        if json_schema:
            return Schema(json_schema=load_json_schema(json_schema))
        elif pydantic_model:
            return Schema(pydantic_model=pydantic_model)
        else:
            raise ValueError("Either json_schema or pydantic_model must be provided")

    """Schemas Asyncronous API wrapper"""

    async def get(self, schema_id: str) -> Schema:
        """Retrieve a schema by ID.

        Args:
            schema_id: The ID of the schema to retrieve

        Returns:
            Schema: The retrieved schema object
        """

        prepared_request = self.prepare_get(schema_id)
        response = await self._client._prepared_request(prepared_request)
        return Schema.model_validate(response)

    async def generate(
        self,
        documents: Sequence[Path | str | bytes | MIMEData | IOBase | PIL.Image.Image],
        instructions: str | None = None,
        model: str = "gpt-4o-2024-11-20",
        temperature: float = 0.0,
        modality: Modality = "native",
    ) -> Schema:
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
        prepared_request = self.prepare_generate(documents, instructions, model, temperature, modality)
        response = await self._client._prepared_request(prepared_request)
        return Schema.model_validate(response)

    async def evaluate(
        self,
        documents: Sequence[Path | str | bytes | MIMEData | IOBase | PIL.Image.Image],
        json_schema: dict[str, Any],
        ground_truths: list[dict[str, Any]] | None = None,
        model: str = "gpt-4o-mini",
        temperature: float = 0.0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        modality: Modality = "native",
        image_resolution_dpi: int = 96,
        browser_canvas: str = "A4",
        n_consensus: int = 1,
    ) -> EvaluateSchemaResponse:
        """
        Evaluate a schema by performing extractions on provided documents.
        If ground truths are provided, compares extractions against them.
        Otherwise, uses consensus likelihoods as metrics.

        Args:
            documents: List of documents to evaluate against
            json_schema: The JSON schema to evaluate
            ground_truths: Optional list of ground truth dictionaries to compare against
            model: The model to use for extraction
            temperature: The temperature to use for extraction
            reasoning_effort: The reasoning effort to use for extraction
            modality: The modality to use for extraction
            image_resolution_dpi: The DPI of the image. Defaults to 96.
            browser_canvas: The canvas size of the browser. Must be one of:
                - "A3" (11.7in x 16.54in)
                - "A4" (8.27in x 11.7in) 
                - "A5" (5.83in x 8.27in)
                Defaults to "A4".
            n_consensus: Number of consensus rounds to perform

        Returns:
            EvaluateSchemaResponse: Dictionary containing evaluation metrics for each document.
            Each metric includes:
            - similarity: Average similarity/likelihood score
            - similarities: Per-field similarity/likelihood scores
            - flat_similarities: Flattened per-field similarity/likelihood scores
            - aligned_* versions of the above metrics

        Raises:
            ValueError: If ground_truths is provided and its length doesn't match documents
            HTTPException: If the request fails or if there are too many documents
        """
        prepared_request = self.prepare_evaluate(
            documents=documents,
            json_schema=json_schema,
            ground_truths=ground_truths,
            model=model,
            temperature=temperature,
            reasoning_effort=reasoning_effort,
            modality=modality,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            n_consensus=n_consensus,
        )
        response = await self._client._prepared_request(prepared_request)
        return EvaluateSchemaResponse.model_validate(response)

    async def enhance(
        self,
        json_schema: dict[str, Any] | Path | str,
        documents: Sequence[Path | str | bytes | MIMEData | IOBase | PIL.Image.Image],
        ground_truths: list[dict[str, Any]] | None = None,
        instructions: str | None = None,
        model: str = "gpt-4o-2024-11-20",
        temperature: float = 0,
        modality: Modality = "native",
        flat_likelihoods: list[dict[str, float]] | dict[str, float] | None = None,
        tools_config: EnhanceSchemaConfigDict | None = None,
    ) -> Schema:
        prepared_request = self.prepare_enhance(
            json_schema, documents, ground_truths, instructions, model, temperature, modality, flat_likelihoods, EnhanceSchemaConfig.model_validate(tools_config or {})
        )
        response = await self._client._prepared_request(prepared_request)
        return Schema(json_schema=response["json_schema"])
