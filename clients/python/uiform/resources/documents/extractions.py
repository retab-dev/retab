from io import IOBase
from pathlib import Path
from typing import Any, AsyncGenerator, Generator, Optional
from pydantic import HttpUrl

from ..._resource import AsyncAPIResource, SyncAPIResource
from ..._utils.ai_model import assert_valid_model_extraction
from ..._utils.json_schema import filter_reasoning_fields_json, load_json_schema
from ..._utils.mime import prepare_mime_document
from ..._utils.stream_context_managers import as_async_context_manager, as_context_manager
from ...types.documents.create_messages import ChatCompletionUiformMessage
from ...types.documents.parse import DocumentExtractRequest, DocumentExtractResponse
from ...types.modalities import Modality


def maybe_parse_to_pydantic(request: DocumentExtractRequest, response: DocumentExtractResponse, allow_partial: bool = False) -> DocumentExtractResponse:
    if response.choices[0].message.content:
        try:
            if allow_partial:
                response.choices[0].message.parsed = request.form_schema._partial_pydantic_model.model_validate(filter_reasoning_fields_json(response.choices[0].message.content))
            else:
                response.choices[0].message.parsed = request.form_schema.pydantic_model.model_validate(filter_reasoning_fields_json(response.choices[0].message.content))
        except Exception as e: 
            pass
    return response

class BaseExtractionsMixin:
    def prepare_extraction(
        self,
        json_schema: dict[str, Any] | Path | str,
        document: Path | str | IOBase | HttpUrl | None,
        image_settings: Optional[dict[str, Any]],
        model: str,
        temperature: float | None,
        modality: Modality,
        stream: bool,
        store: bool = False,
    ) -> DocumentExtractRequest:
        
        assert_valid_model_extraction(model)

        json_schema = load_json_schema(json_schema)

        data = {
            "json_schema": json_schema,
            "document": prepare_mime_document(document).model_dump() if document is not None else None,
            "model": model,
            "temperature": temperature,
            "stream": stream,
            "modality": modality,
            "store": store,
        }
        if image_settings:
            data["image_settings"] = image_settings

        # Validate DocumentAPIRequest data (raises exception if invalid)
        return DocumentExtractRequest.model_validate(data)

    def prepare_parse(
        self,
        json_schema: dict[str, Any] | Path | str,
        document: Path | str | IOBase | HttpUrl | None,
        image_settings: Optional[dict[str, Any]],
        model: str,
        temperature: float,
        modality: Modality,
        store: bool = False,
    ) -> DocumentExtractRequest:
        stream = False
        return self.prepare_extraction(
            json_schema, document, image_settings, model, temperature, modality, stream, store
        )
    
    def prepare_stream(
        self,
        json_schema: dict[str, Any] | Path | str,
        document: Path | str | IOBase | HttpUrl | None,
        image_settings: Optional[dict[str, Any]],
        model: str,
        temperature: float,
        modality: Modality,
        store: bool = False,
    ) -> DocumentExtractRequest:
        stream = True
        return self.prepare_extraction(
            json_schema, document, image_settings, model, temperature, modality, stream, store
        )


class Extractions(SyncAPIResource, BaseExtractionsMixin):
    """Extraction API wrapper"""

    def parse(
        self,
        json_schema: dict[str, Any] | Path | str,
        document: Path | str | IOBase | HttpUrl | None,
        image_settings: Optional[dict[str, Any]] = None,
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        modality: Modality = "native",
        idempotency_key: str | None = None,
        store: bool = False,
    ) -> DocumentExtractResponse:
        """
        Process a document using the UiForm API.

        Args:
            json_schema: JSON schema defining the expected data structure
            document: Single document (as MIMEData) to process
            model: The AI model to use for processing
            temperature: Model temperature setting (0-1)
            idempotency_key: Idempotency key for request
        Returns:
            DocumentAPIResponse
        Raises:
            HTTPException if the request fails
        """

        assert (document is not None), "Either document or messages must be provided"

        # Validate DocumentAPIRequest data (raises exception if invalid)
        request = self.prepare_parse(json_schema, document, image_settings, model, temperature, modality, store)
        response = self._client._request("POST", "/api/v1/documents/extractions", data=request.model_dump(), idempotency_key=idempotency_key)
        return maybe_parse_to_pydantic(request, DocumentExtractResponse.model_validate(response))

    @as_context_manager
    def stream(
        self,
        json_schema: dict[str, Any] | Path | str,
        document: Path | str | IOBase | HttpUrl | None,
        image_settings: Optional[dict[str, Any]] = None,
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        modality: Modality = "native",
        idempotency_key: str | None = None,
        store: bool = False,
    ) -> Generator[DocumentExtractResponse, None, None]:
        """
        Process a document using the UiForm API with streaming enabled.

        Args:
            json_schema: JSON schema defining the expected data structure
            document: Single document (as MIMEData) to process
            image_settings: 
            model: The AI model to use for processing
            temperature: Model temperature setting (0-1)
            modality: Modality of the document (e.g., native)

        Returns:
            Generator[DocumentExtractResponse]: Stream of parsed responses
        Raises:
            HTTPException if the request fails
        Usage:
        ```python
        with uiform.documents.extractions.stream(json_schema, document, model, temperature, messages, modality) as stream:
            for response in stream:
                print(response)
        ```
        """
        request = self.prepare_stream(json_schema, document, image_settings, model, temperature, modality, store)

        # Request the stream and return a context manager
        chunk_json: Any = None
        for chunk_json in self._client._request_stream("POST", "/api/v1/documents/extractions", data=request.model_dump(), idempotency_key=idempotency_key):
            if not chunk_json:
                continue
            yield maybe_parse_to_pydantic(request, DocumentExtractResponse.model_validate(chunk_json), allow_partial=True)
        yield maybe_parse_to_pydantic(request, DocumentExtractResponse.model_validate(chunk_json))


class AsyncExtractions(AsyncAPIResource, BaseExtractionsMixin):
    """Extraction API wrapper for asynchronous usage."""

    async def parse(
        self,
        json_schema: dict[str, Any] | Path | str,
        document: Path | str | IOBase | HttpUrl | None,
        image_settings: Optional[dict[str, Any]] = None,
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        modality: Modality = "native",
        idempotency_key: str | None = None,
        store: bool = False,
    ) -> DocumentExtractResponse:
        """
        Extract structured data from a document asynchronously.

        Args:
            json_schema: JSON schema defining the expected data structure.
            document: Path, string, or file-like object representing the document.
            image_settings : Optional additional context for the model.
            model: The AI model to use.
            temperature: Model temperature setting (0-1).
            modality: Modality of the document (e.g., native).

        Returns:
            DocumentExtractResponse: Parsed response from the API.
        """
        request = self.prepare_parse(json_schema, document, image_settings, model, temperature, modality, store)
        response = await self._client._request("POST", "/api/v1/documents/extractions", data=request.model_dump(), idempotency_key=idempotency_key)
        return maybe_parse_to_pydantic(request, DocumentExtractResponse.model_validate(response))

    @as_async_context_manager
    async def stream(
        self,
        json_schema: dict[str, Any] | Path | str,
        document: Path | str | IOBase | HttpUrl | None,
        image_settings: Optional[dict[str, Any]] = None,
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        modality: Modality = "native",
        idempotency_key: str | None = None,
        store: bool = False,
    ) -> AsyncGenerator[DocumentExtractResponse, None]:
        """
        Extract structured data from a document asynchronously with streaming.

        Args:
            json_schema: JSON schema defining the expected data structure.
            document: Path, string, or file-like object representing the document.
            model: The AI model to use.
            temperature: Model temperature setting (0-1).
            modality: Modality of the document (e.g., native).
            idempotency_key: Idempotency key for request
        Returns:
            AsyncGenerator[DocumentExtractResponse, None]: Stream of parsed responses.

        Usage:
        ```python
        async with uiform.documents.extractions.stream(json_schema, document, model, temperature, messages, modality) as stream:
            async for response in stream:
                print(response)
        ```
        """
        request = self.prepare_stream(json_schema, document, image_settings, model, temperature, modality, store)
        chunk_json: Any = None
        async for chunk_json in self._client._request_stream("POST", "/api/v1/documents/extractions", data=request.model_dump(), idempotency_key=idempotency_key):
            if not chunk_json:
                continue
            
            yield maybe_parse_to_pydantic(request, DocumentExtractResponse.model_validate(chunk_json), allow_partial=True)
        # Last chunk with full parsed response
        yield maybe_parse_to_pydantic(request, DocumentExtractResponse.model_validate(chunk_json))  
