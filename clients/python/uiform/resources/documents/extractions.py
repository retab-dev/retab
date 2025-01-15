from typing import Any, Optional, Generator, AsyncGenerator
from pathlib import Path
from io import IOBase

from ...types.documents.parse import DocumentExtractRequest, DocumentExtractResponse
from ...types.modalities import Modality
from ...types.documents.create_messages import ChatCompletionUiformMessage
from ..._utils.mime import prepare_mime_document
from ..._utils.json_schema import load_json_schema, filter_reasoning_fields_json
from ..._utils.ai_model import assert_valid_model_extraction
from ..._utils.stream_context_managers import as_async_context_manager, as_context_manager
from ..._resource import SyncAPIResource, AsyncAPIResource


def maybe_parse_to_pydantic(request: DocumentExtractRequest, response: DocumentExtractResponse, allow_partial: bool = False) -> DocumentExtractResponse:
    if response.choices[0].message.content:
        try:
            if allow_partial:
                response.choices[0].message.parsed = request.form_schema._partial_pydantic_model.model_validate(filter_reasoning_fields_json(response.choices[0].message.content, request.form_schema._partial_pydantic_model))
            else:
                response.choices[0].message.parsed = request.form_schema.pydantic_model.model_validate(filter_reasoning_fields_json(response.choices[0].message.content, request.form_schema.pydantic_model))
        except Exception as e: 
            pass
    return response

class BaseExtractionsMixin:
    def prepare_extraction(
        self,
        json_schema: dict[str, Any] | Path | str,
        document: Path | str | IOBase,
        text_operations: Optional[dict[str, Any]],
        model: str,
        temperature: float | None,
        messages: list[ChatCompletionUiformMessage],
        modality: Modality,
        stream: bool,
    ) -> DocumentExtractRequest:
        assert_valid_model_extraction(model)

        json_schema = load_json_schema(json_schema)

        mime_document = prepare_mime_document(document)
        data = {
            "json_schema": json_schema,
            "document": mime_document.model_dump(),
            "model": model,
            "temperature": temperature,
            "stream": stream,
            "text_operations": text_operations or {},
            "modality": modality,
            "messages": messages,
        }

        # Validate DocumentAPIRequest data (raises exception if invalid)
        return DocumentExtractRequest.model_validate(data)

    def prepare_parse(
        self,
        json_schema: dict[str, Any] | Path | str,
        document: Path | str | IOBase,
        text_operations: Optional[dict[str, Any]],
        model: str,
        temperature: float,
        messages: list[ChatCompletionUiformMessage],
        modality: Modality,
    ) -> DocumentExtractRequest:
        stream = False
        return self.prepare_extraction(
            json_schema, document, text_operations, model, temperature, messages, modality, stream
        )
    def prepare_stream(
        self,
        json_schema: dict[str, Any] | Path | str,
        document: Path | str | IOBase,
        text_operations: Optional[dict[str, Any]],
        model: str,
        temperature: float,
        messages: list[ChatCompletionUiformMessage],
        modality: Modality,
    ) -> DocumentExtractRequest:
        stream = True
        return self.prepare_extraction(
            json_schema, document, text_operations, model, temperature, messages, modality, stream
        )


class Extractions(SyncAPIResource, BaseExtractionsMixin):
    """Extraction API wrapper"""

    def parse(
        self,
        json_schema: dict[str, Any] | Path | str,
        document: Path | str | IOBase,
        text_operations: Optional[dict[str, Any]] = None,
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        messages: list[ChatCompletionUiformMessage] = [],
        modality: Modality = "native",
    ) -> DocumentExtractResponse:
        """
        Process a document using the UiForm API.

        Args:
            json_schema: JSON schema defining the expected data structure
            document: Single document (as MIMEData) to process
            model: The AI model to use for processing
            temperature: Model temperature setting (0-1)
            text_operations: Optional context with regex instructions or other metadata

        Returns:
            DocumentAPIResponse
        Raises:
            HTTPException if the request fails
        """

        # Validate DocumentAPIRequest data (raises exception if invalid)
        request = self.prepare_parse(json_schema, document, text_operations, model, temperature, messages, modality)
        response = self._client._request("POST", "/api/v1/documents/extractions", data=request.model_dump())
        return maybe_parse_to_pydantic(request, DocumentExtractResponse.model_validate(response))

    @as_context_manager
    def stream(
        self,
        json_schema: dict[str, Any] | Path | str,
        document: Path | str | IOBase,
        text_operations: Optional[dict[str, Any]] = None,
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        messages: list[ChatCompletionUiformMessage] = [],
        modality: Modality = "native",
    ) -> Generator[DocumentExtractResponse, None, None]:
        """
        Process a document using the UiForm API with streaming enabled.

        Args:
            json_schema: JSON schema defining the expected data structure
            document: Single document (as MIMEData) to process
            text_operations: Optional context with regex instructions or other metadata
            model: The AI model to use for processing
            temperature: Model temperature setting (0-1)
            modality: Modality of the document (e.g., native)
            messages: List of chat completion messages for context

        Returns:
            Generator[DocumentExtractResponse]: Stream of parsed responses
        Raises:
            HTTPException if the request fails
        Usage:
        ```python
        with uiform.documents.extractions.stream(json_schema, document, text_operations, model, temperature, messages, modality) as stream:
            for response in stream:
                print(response)
        ```
        """
        request = self.prepare_stream(json_schema, document, text_operations, model, temperature, messages, modality)

        # Request the stream and return a context manager
        chunk_json: Any = None
        for chunk_json in self._client._request_stream("POST", "/api/v1/documents/extractions", data=request.model_dump()):
            if not chunk_json:
                continue
            yield maybe_parse_to_pydantic(request, DocumentExtractResponse.model_validate(chunk_json), allow_partial=True)
        yield maybe_parse_to_pydantic(request, DocumentExtractResponse.model_validate(chunk_json))


class AsyncExtractions(AsyncAPIResource, BaseExtractionsMixin):
    """Extraction API wrapper for asynchronous usage."""

    async def parse(
        self,
        json_schema: dict[str, Any] | Path | str,
        document: Path | str | IOBase,
        text_operations: Optional[dict[str, Any]] = None,
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        messages: list[ChatCompletionUiformMessage] = [],
        modality: Modality = "native",
    ) -> DocumentExtractResponse:
        """
        Extract structured data from a document asynchronously.

        Args:
            json_schema: JSON schema defining the expected data structure.
            document: Path, string, or file-like object representing the document.
            text_operations: Optional additional context for the model.
            model: The AI model to use.
            temperature: Model temperature setting (0-1).
            modality: Modality of the document (e.g., native).

        Returns:
            DocumentExtractResponse: Parsed response from the API.
        """
        request = self.prepare_parse(json_schema, document, text_operations, model, temperature, messages, modality)
        response = await self._client._request("POST", "/api/v1/documents/extractions", data=request.model_dump())
        return maybe_parse_to_pydantic(request, DocumentExtractResponse.model_validate(response))

    @as_async_context_manager
    async def stream(
        self,
        json_schema: dict[str, Any] | Path | str,
        document: Path | str | IOBase,
        text_operations: Optional[dict[str, Any]] = None,
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        messages: list[ChatCompletionUiformMessage] = [],
        modality: Modality = "native",
    ) -> AsyncGenerator[DocumentExtractResponse, None]:
        """
        Extract structured data from a document asynchronously with streaming.

        Args:
            json_schema: JSON schema defining the expected data structure.
            document: Path, string, or file-like object representing the document.
            text_operations: Optional additional context for the model.
            model: The AI model to use.
            temperature: Model temperature setting (0-1).
            modality: Modality of the document (e.g., native).

        Returns:
            AsyncGenerator[DocumentExtractResponse, None]: Stream of parsed responses.

        Usage:
        ```python
        async with uiform.documents.extractions.stream(json_schema, document, text_operations, model, temperature, messages, modality) as stream:
            async for response in stream:
                print(response)
        ```
        """
        request = self.prepare_stream(json_schema, document, text_operations, model, temperature, messages, modality)
        chunk_json: Any = None
        async for chunk_json in self._client._request_stream("POST", "/api/v1/documents/extractions", data=request.model_dump()):
            if not chunk_json:
                continue
            
            yield maybe_parse_to_pydantic(request, DocumentExtractResponse.model_validate(chunk_json), allow_partial=True)
        # Last chunk with full parsed response
        yield maybe_parse_to_pydantic(request, DocumentExtractResponse.model_validate(chunk_json))  
