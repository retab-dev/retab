from io import IOBase
from pathlib import Path
from typing import Any, AsyncGenerator, Generator, Optional
from pydantic import HttpUrl

from ..._resource import AsyncAPIResource, SyncAPIResource
from ..._utils.ai_model import assert_valid_model_extraction
from ..._utils.json_schema import filter_reasoning_fields_json, load_json_schema
from ..._utils.mime import prepare_mime_document
from ..._utils.stream_context_managers import as_async_context_manager, as_context_manager
from ...types.documents.extractions import DocumentExtractRequest, DocumentExtractResponse, DocumentExtractResponseStream, LogExtractionRequest
from ...types.modalities import Modality
from ...types.standards import PreparedRequest
from ...types.schemas.object import Schema
from ...types.chat import ChatCompletionUiformMessage
from typing import overload

@overload
def maybe_parse_to_pydantic(schema: Schema, response: DocumentExtractResponseStream, allow_partial: bool = False) -> DocumentExtractResponseStream: ...

@overload
def maybe_parse_to_pydantic(schema: Schema, response: DocumentExtractResponse, allow_partial: bool = False) -> DocumentExtractResponse: ...


def maybe_parse_to_pydantic(schema: Schema, response: DocumentExtractResponse | DocumentExtractResponseStream, allow_partial: bool = False) -> DocumentExtractResponse | DocumentExtractResponseStream:
    
    if response.choices[0].message.content:
        try:
            if allow_partial:
                response.choices[0].message.parsed = schema._partial_pydantic_model.model_validate(filter_reasoning_fields_json(response.choices[0].message.content))
            else:
                response.choices[0].message.parsed = schema.pydantic_model.model_validate(filter_reasoning_fields_json(response.choices[0].message.content))
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
        temperature: float,
        modality: Modality,
        stream: bool,
        store: bool = False,
        idempotency_key: str | None = None,
    ) -> PreparedRequest:
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
        document_extract_request = DocumentExtractRequest.model_validate(data)

        return PreparedRequest(
            method="POST",
            url="/v1/documents/extractions",
            data=document_extract_request.model_dump(),
            idempotency_key=idempotency_key
        )

    def prepare_log_extraction(self, messages: list[ChatCompletionUiformMessage], completion: Any, json_schema: dict[str, Any], model: str, temperature: float) -> PreparedRequest:
        return PreparedRequest(
            method="POST",
            url="/v1/documents/log_extraction",
            data=LogExtractionRequest(messages=messages, completion=completion, json_schema=json_schema, model=model, temperature=temperature).model_dump(),
            raise_for_status=True
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
        request = self.prepare_extraction(json_schema, document, image_settings, model, temperature, modality, False, store, idempotency_key=idempotency_key)        
        response = self._client._prepared_request(request)

        schema = Schema(json_schema=load_json_schema(json_schema))
        return maybe_parse_to_pydantic(schema, DocumentExtractResponse.model_validate(response))

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
    ) -> Generator[DocumentExtractResponseStream, None, None]:
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
        request = self.prepare_extraction(json_schema, document, image_settings, model, temperature, modality, True, store, idempotency_key=idempotency_key)
        schema = Schema(json_schema=load_json_schema(json_schema))

        # Request the stream and return a context manager
        chunk_json: Any = None
        for chunk_json in self._client._prepared_request_stream(request):
            if not chunk_json:
                continue
            yield maybe_parse_to_pydantic(schema, DocumentExtractResponseStream.model_validate(chunk_json), allow_partial=True)
        yield maybe_parse_to_pydantic(schema, DocumentExtractResponseStream.model_validate(chunk_json))

    def log_extraction(self, messages: list[ChatCompletionUiformMessage], completion: Any, json_schema: dict[str, Any], model: str, temperature: float) -> None:
        request = self.prepare_log_extraction(messages, completion, json_schema, model, temperature)
        self._client._prepared_request(request)
        return None

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
        request = self.prepare_extraction(json_schema, document, image_settings, model, temperature, modality, False, store, idempotency_key=idempotency_key)
        response = await self._client._prepared_request(request)
        schema = Schema(json_schema=load_json_schema(json_schema))
        return maybe_parse_to_pydantic(schema, DocumentExtractResponse.model_validate(response))

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
    ) -> AsyncGenerator[DocumentExtractResponseStream, None]:
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
        request = self.prepare_extraction(json_schema, document, image_settings, model, temperature, modality, True, store, idempotency_key=idempotency_key)
        schema = Schema(json_schema=load_json_schema(json_schema))
        chunk_json: Any = None
        async for chunk_json in self._client._prepared_request_stream(request):
            if not chunk_json:
                continue
            
            yield maybe_parse_to_pydantic(schema, DocumentExtractResponseStream.model_validate(chunk_json), allow_partial=True)
        # Last chunk with full parsed response
        yield maybe_parse_to_pydantic(schema, DocumentExtractResponseStream.model_validate(chunk_json))  
    
    async def log_extraction(self, messages: list[ChatCompletionUiformMessage], completion: Any, json_schema: dict[str, Any], model: str, temperature: float) -> None:
        request = self.prepare_log_extraction(messages, completion, json_schema, model, temperature)
        await self._client._prepared_request(request)
        return None
