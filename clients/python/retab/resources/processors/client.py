import base64
from io import IOBase
from pathlib import Path
from typing import Any, List, Literal

import PIL.Image
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel, HttpUrl
from pydantic_core import PydanticUndefined

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...utils.ai_models import assert_valid_model_extraction
from ...utils.json_schema import load_json_schema
from ...utils.mime import MIMEData, prepare_mime_document
from ...types.browser_canvas import BrowserCanvas
from ...types.documents.extractions import RetabParsedChatCompletion
from ...types.logs import ProcessorConfig, UpdateProcessorRequest
from ...types.modalities import Modality
from ...types.pagination import ListMetadata
from ...types.standards import PreparedRequest
from .automations.client import AsyncAutomations, Automations


class ListProcessors(BaseModel):
    """Response model for listing processor configurations."""

    data: list[ProcessorConfig]
    list_metadata: ListMetadata


class ProcessorsMixin:
    def prepare_create(
        self,
        name: str,
        json_schema: dict[str, Any] | Path | str,
        modality: Modality = "native",
        model: str = "gpt-4o-mini",
        temperature: float = PydanticUndefined,  # type: ignore[assignment]
        reasoning_effort: ChatCompletionReasoningEffort = PydanticUndefined,  # type: ignore[assignment]
        image_resolution_dpi: int = PydanticUndefined,  # type: ignore[assignment]
        browser_canvas: BrowserCanvas = PydanticUndefined,  # type: ignore[assignment]
        n_consensus: int = PydanticUndefined,  # type: ignore[assignment]
    ) -> PreparedRequest:
        assert_valid_model_extraction(model)

        # Load the JSON schema from file path, string, or dict
        loaded_schema = load_json_schema(json_schema)

        processor_config = ProcessorConfig(
            name=name,
            json_schema=loaded_schema,
            modality=modality,
            model=model,
            temperature=temperature,
            reasoning_effort=reasoning_effort,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            n_consensus=n_consensus,
        )

        return PreparedRequest(method="POST", url="/v1/processors", data=processor_config.model_dump(mode="json"))

    def prepare_list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int | None = 10,
        order: Literal["asc", "desc"] | None = "desc",
        # Filtering parameters
        name: str | None = None,
        modality: str | None = None,
        model: str | None = None,
        schema_id: str | None = None,
        schema_data_id: str | None = None,
    ) -> PreparedRequest:
        params = {
            "before": before,
            "after": after,
            "limit": limit,
            "order": order,
            "name": name,
            "modality": modality,
            "model": model,
            "schema_id": schema_id,
            "schema_data_id": schema_data_id,
        }
        # Remove None values
        params = {k: v for k, v in params.items() if v is not None}

        return PreparedRequest(method="GET", url="/v1/processors", params=params)

    def prepare_get(self, processor_id: str) -> PreparedRequest:
        """Get a specific processor configuration.

        Args:
            processor_id: ID of the processor

        Returns:
            ProcessorConfig: The processor configuration
        """
        return PreparedRequest(method="GET", url=f"/v1/processors/{processor_id}")

    def prepare_update(
        self,
        processor_id: str,
        name: str | None = None,
        modality: Modality | None = None,
        image_resolution_dpi: int | None = None,
        browser_canvas: BrowserCanvas | None = None,
        model: str | None = None,
        json_schema: dict[str, Any] | Path | str | None = None,
        temperature: float | None = None,
        reasoning_effort: ChatCompletionReasoningEffort | None = None,
        n_consensus: int | None = None,
    ) -> PreparedRequest:
        if model is not None:
            assert_valid_model_extraction(model)

        # Load the JSON schema from file path, string, or dict if provided
        loaded_schema = None
        if json_schema is not None:
            loaded_schema = load_json_schema(json_schema)

        update_request = UpdateProcessorRequest(
            name=name,
            modality=modality,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            model=model,
            json_schema=loaded_schema,
            temperature=temperature,
            reasoning_effort=reasoning_effort,
            n_consensus=n_consensus,
        )

        return PreparedRequest(method="PUT", url=f"/v1/processors/{processor_id}", data=update_request.model_dump(mode="json", exclude_none=True))

    def prepare_delete(self, processor_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/v1/processors/{processor_id}")

    def prepare_submit(
        self,
        processor_id: str,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        documents: list[Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl] | None = None,
        temperature: float | None = None,
        seed: int | None = None,
        store: bool = True,
    ) -> PreparedRequest:
        """Prepare a request to submit documents to a processor.

        Args:
            processor_id: ID of the processor
            document: Single document to process (mutually exclusive with documents)
            documents: List of documents to process (mutually exclusive with document)
            temperature: Optional temperature override
            seed: Optional seed for reproducibility
            store: Whether to store the results

        Returns:
            PreparedRequest: The prepared request
        """
        # Validate that either document or documents is provided, but not both
        if not document and not documents:
            raise ValueError("Either 'document' or 'documents' must be provided")

        if document and documents:
            raise ValueError("Provide either 'document' (single) or 'documents' (multiple), not both")

        # Prepare form data parameters
        form_data = {
            "temperature": temperature,
            "seed": seed,
            "store": store,
        }
        # Remove None values
        form_data = {k: v for k, v in form_data.items() if v is not None}

        # Prepare files for upload
        files = {}
        if document:
            # Convert document to MIMEData if needed
            mime_document = prepare_mime_document(document)
            # Single document upload
            files["document"] = (mime_document.filename, base64.b64decode(mime_document.content), mime_document.mime_type)
        elif documents:
            # Multiple documents upload - httpx supports multiple files with same field name using a list
            files_list = []
            for doc in documents:
                # Convert each document to MIMEData if needed
                mime_doc = prepare_mime_document(doc)
                files_list.append(
                    (
                        "documents",  # field name
                        (mime_doc.filename, base64.b64decode(mime_doc.content), mime_doc.mime_type),
                    )
                )
            files = files_list

        url = f"/v1/processors/{processor_id}/submit"

        return PreparedRequest(method="POST", url=url, form_data=form_data, files=files)


class Processors(SyncAPIResource, ProcessorsMixin):
    """Processors API wrapper for managing processor configurations"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.automations = Automations(client=client)

    def create(
        self,
        name: str,
        json_schema: dict[str, Any] | Path | str,
        modality: Modality = "native",
        model: str = "gpt-4o-mini",
        temperature: float = PydanticUndefined,  # type: ignore[assignment]
        reasoning_effort: ChatCompletionReasoningEffort = PydanticUndefined,  # type: ignore[assignment]
        image_resolution_dpi: int = PydanticUndefined,  # type: ignore[assignment]
        browser_canvas: BrowserCanvas = PydanticUndefined,  # type: ignore[assignment]
        n_consensus: int = PydanticUndefined,  # type: ignore[assignment]
    ) -> ProcessorConfig:
        """Create a new processor configuration.

        Args:
            name: Name of the processor
            json_schema: JSON schema for the processor. Can be a dictionary, file path (Path or str), or JSON string.
            image_resolution_dpi: Optional image resolution DPI
            browser_canvas: Optional browser canvas size
            modality: Processing modality (currently only "native" supported)
            model: AI model to use for processing
            temperature: Model temperature setting
            reasoning_effort: The effort level for the model to reason about the input data.
            n_consensus: Number of consensus required to validate the data
        Returns:
            ProcessorConfig: The created processor configuration
        """
        request = self.prepare_create(
            name=name,
            json_schema=json_schema,
            modality=modality,
            model=model,
            temperature=temperature,
            reasoning_effort=reasoning_effort,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            n_consensus=n_consensus,
        )
        response = self._client._prepared_request(request)
        print(f"Processor ID: {response['id']}. Processor available at https://www.retab.com/dashboard/processors/{response['id']}")
        return ProcessorConfig.model_validate(response)

    def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int | None = 10,
        order: Literal["asc", "desc"] | None = "desc",
        name: str | None = None,
        modality: str | None = None,
        model: str | None = None,
        schema_id: str | None = None,
        schema_data_id: str | None = None,
    ) -> ListProcessors:
        """List processor configurations with pagination support.

        Args:
            before: Optional cursor for pagination before a specific processor ID
            after: Optional cursor for pagination after a specific processor ID
            limit: Optional limit on number of results (max 100)
            order: Optional sort order ("asc" or "desc")
            name: Optional filter by processor name
            modality: Optional filter by modality
            model: Optional filter by model
            schema_id: Optional filter by schema ID
            schema_data_id: Optional filter by schema data ID

        Returns:
            ListProcessors: Paginated list of processor configurations with metadata
        """
        request = self.prepare_list(before, after, limit, order, name, modality, model, schema_id, schema_data_id)
        response = self._client._prepared_request(request)
        return ListProcessors.model_validate(response)

    def get(self, processor_id: str) -> ProcessorConfig:
        """Get a specific processor configuration.

        Args:
            processor_id: ID of the processor

        Returns:
            ProcessorConfig: The processor configuration
        """
        request = self.prepare_get(processor_id)
        response = self._client._prepared_request(request)
        return ProcessorConfig.model_validate(response)

    def update(
        self,
        processor_id: str,
        name: str | None = None,
        modality: Modality | None = None,
        image_resolution_dpi: int | None = None,
        browser_canvas: BrowserCanvas | None = None,
        model: str | None = None,
        json_schema: dict[str, Any] | Path | str | None = None,
        temperature: float | None = None,
        reasoning_effort: ChatCompletionReasoningEffort | None = None,
        n_consensus: int | None = None,
    ) -> ProcessorConfig:
        """Update a processor configuration.

        Args:
            processor_id: ID of the processor to update
            name: New name for the processor
            modality: New processing modality
            image_resolution_dpi: New image resolution DPI
            browser_canvas: New browser canvas size
            model: New AI model
            json_schema: New JSON schema for the processor. Can be a dictionary, file path (Path or str), or JSON string.
            temperature: New temperature setting
            reasoning_effort: The effort level for the model to reason about the input data.
            n_consensus: New number of consensus required
        Returns:
            ProcessorConfig: The updated processor configuration
        """
        request = self.prepare_update(
            processor_id=processor_id,
            name=name,
            modality=modality,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            model=model,
            json_schema=json_schema,
            temperature=temperature,
            reasoning_effort=reasoning_effort,
            n_consensus=n_consensus,
        )
        response = self._client._prepared_request(request)
        return ProcessorConfig.model_validate(response)

    def delete(self, processor_id: str) -> None:
        """Delete a processor configuration.

        Args:
            processor_id: ID of the processor to delete
        """
        request = self.prepare_delete(processor_id)
        self._client._prepared_request(request)
        print(f"Processor Deleted. ID: {processor_id}")

    def submit(
        self,
        processor_id: str,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        documents: List[Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl] | None = None,
        temperature: float | None = None,
        seed: int | None = None,
        store: bool = True,
    ) -> RetabParsedChatCompletion:
        """Submit documents to a processor for processing.

        Args:
            processor_id: ID of the processor
            document: Single document to process (mutually exclusive with documents)
            documents: List of documents to process (mutually exclusive with document)
            temperature: Optional temperature override
            seed: Optional seed for reproducibility
            store: Whether to store the results

        Returns:
            RetabParsedChatCompletion: The processing result
        """
        request = self.prepare_submit(processor_id=processor_id, document=document, documents=documents, temperature=temperature, seed=seed, store=store)
        response = self._client._prepared_request(request)
        return RetabParsedChatCompletion.model_validate(response)


class AsyncProcessors(AsyncAPIResource, ProcessorsMixin):
    """Async Processors API wrapper for managing processor configurations"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.automations = AsyncAutomations(client=client)

    async def create(
        self,
        name: str,
        json_schema: dict[str, Any] | Path | str,
        modality: Modality = "native",
        model: str = "gpt-4o-mini",
        temperature: float = PydanticUndefined,  # type: ignore[assignment]
        reasoning_effort: ChatCompletionReasoningEffort = PydanticUndefined,  # type: ignore[assignment]
        image_resolution_dpi: int = PydanticUndefined,  # type: ignore[assignment]
        browser_canvas: BrowserCanvas = PydanticUndefined,  # type: ignore[assignment]
        n_consensus: int = PydanticUndefined,  # type: ignore[assignment]
    ) -> ProcessorConfig:
        request = self.prepare_create(
            name=name,
            json_schema=json_schema,
            modality=modality,
            model=model,
            temperature=temperature,
            reasoning_effort=reasoning_effort,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            n_consensus=n_consensus,
        )
        response = await self._client._prepared_request(request)
        print(f"Processor ID: {response['id']}. Processor available at https://www.retab.com/dashboard/processors/{response['id']}")

        return ProcessorConfig.model_validate(response)

    async def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int | None = 10,
        order: Literal["asc", "desc"] | None = "desc",
        name: str | None = None,
        modality: str | None = None,
        model: str | None = None,
        schema_id: str | None = None,
        schema_data_id: str | None = None,
    ) -> ListProcessors:
        request = self.prepare_list(before, after, limit, order, name, modality, model, schema_id, schema_data_id)
        response = await self._client._prepared_request(request)
        return ListProcessors.model_validate(response)

    async def get(self, processor_id: str) -> ProcessorConfig:
        request = self.prepare_get(processor_id)
        response = await self._client._prepared_request(request)
        return ProcessorConfig.model_validate(response)

    async def update(
        self,
        processor_id: str,
        name: str | None = None,
        modality: Modality | None = None,
        image_resolution_dpi: int | None = None,
        browser_canvas: BrowserCanvas | None = None,
        model: str | None = None,
        json_schema: dict[str, Any] | Path | str | None = None,
        temperature: float | None = None,
        reasoning_effort: ChatCompletionReasoningEffort | None = None,
        n_consensus: int | None = None,
    ) -> ProcessorConfig:
        """Update a processor configuration.

        Args:
            processor_id: ID of the processor to update
            name: New name for the processor
            modality: New processing modality
            image_resolution_dpi: New image resolution DPI
            browser_canvas: New browser canvas size
            model: New AI model
            json_schema: New JSON schema for the processor. Can be a dictionary, file path (Path or str), or JSON string.
            temperature: New temperature setting
            reasoning_effort: The effort level for the model to reason about the input data.
            n_consensus: New number of consensus required
        Returns:
            ProcessorConfig: The updated processor configuration
        """
        request = self.prepare_update(processor_id, name, modality, image_resolution_dpi, browser_canvas, model, json_schema, temperature, reasoning_effort, n_consensus)
        response = await self._client._prepared_request(request)
        return ProcessorConfig.model_validate(response)

    async def delete(self, processor_id: str) -> None:
        request = self.prepare_delete(processor_id)
        await self._client._prepared_request(request)
        print(f"Processor Deleted. ID: {processor_id}")

    async def submit(
        self,
        processor_id: str,
        document: Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl | None = None,
        documents: List[Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl] | None = None,
        temperature: float | None = None,
        seed: int | None = None,
        store: bool = True,
    ) -> RetabParsedChatCompletion:
        """Submit documents to a processor for processing.

        Args:
            processor_id: ID of the processor
            document: Single document to process (mutually exclusive with documents)
            documents: List of documents to process (mutually exclusive with document)
            temperature: Optional temperature override
            seed: Optional seed for reproducibility
            store: Whether to store the results

        Returns:
            RetabParsedChatCompletion: The processing result
        """
        request = self.prepare_submit(processor_id=processor_id, document=document, documents=documents, temperature=temperature, seed=seed, store=store)
        response = await self._client._prepared_request(request)
        return RetabParsedChatCompletion.model_validate(response)
