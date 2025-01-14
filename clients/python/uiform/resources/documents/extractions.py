from typing import Any, Optional, Iterator, AsyncIterator
from pathlib import Path
from io import IOBase

from ...types.documents.parse import DocumentExtractRequest, DocumentExtractResponse, Modality
from ...utils import load_json_schema, assert_valid_model_extraction, prepare_mime_document
from ..._resource import SyncAPIResource, AsyncAPIResource
from ...types.documents.create_messages import ChatCompletionUiformMessage
from ...jsonschema.utils import convert_basemodel_to_partial_basemodel 


def maybe_parse_to_pydantic(request: DocumentExtractRequest, response: DocumentExtractResponse) -> DocumentExtractResponse:
    if response.choices[0].message.content:
        try:
            response.choices[0].message.parsed = request.form_schema.pydantic_model.model_validate_json(response.choices[0].message.content)
        except Exception:
            try:
                response.choices[0].message.parsed = request.form_schema._partial_pydantic_model.model_validate_json(response.choices[0].message.content)
            except Exception: pass
    return response


class Extractions(SyncAPIResource):
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

        assert_valid_model_extraction(model)

        json_schema = load_json_schema(json_schema)

        mime_document = prepare_mime_document(document)
        data = {
            "json_schema": json_schema,
            "document": mime_document.model_dump(),
            "model": model,
            "temperature": temperature,
            "stream": False,
            "text_operations": text_operations or {},
            "modality": modality,
            "messages": messages,
        }

        # Validate DocumentAPIRequest data (raises exception if invalid)
        request_object = DocumentExtractRequest.model_validate(data)
        return maybe_parse_to_pydantic(request_object, DocumentExtractResponse.model_validate(self._client._request("POST", "/api/v1/documents/extractions", data=data)))

    def stream(
        self,
        json_schema: dict[str, Any] | Path | str,
        document: Path | str | IOBase,
        text_operations: Optional[dict[str, Any]] = None,
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        modality: Modality = "native",
        messages: list[ChatCompletionUiformMessage] = [],
    ) -> Iterator[DocumentExtractResponse]:
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
            Iterator[DocumentExtractResponse]: Stream of parsed responses
        Raises:
            HTTPException if the request fails
        """
        assert_valid_model_extraction(model)

        json_schema = load_json_schema(json_schema)

        mime_document = prepare_mime_document(document)
        data = {
            "json_schema": json_schema,
            "document": mime_document.model_dump(),
            "model": model,
            "temperature": temperature,
            "stream": True,
            "text_operations": text_operations or {},
            "modality": modality,
            "messages": messages,
        }
        # Validate DocumentAPIRequest data (raises exception if invalid)
        request_object = DocumentExtractRequest.model_validate(data)

        # Request the stream and return a context manager
        for chunk_json in self._client._request_stream("POST", "/api/v1/documents/extractions", data=data):
            if not chunk_json:
                continue
            try:
                yield maybe_parse_to_pydantic(request_object, DocumentExtractResponse.model_validate(chunk_json))
            except Exception:
                pass


class AsyncExtractions(AsyncAPIResource):
    """Extraction API wrapper for asynchronous usage."""

    async def parse(
        self,
        json_schema: dict[str, Any] | Path | str,
        document: Path | str | IOBase,
        text_operations: Optional[dict[str, Any]] = None,
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
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
        assert_valid_model_extraction(model)

        json_schema = load_json_schema(json_schema)

        mime_document = prepare_mime_document(document)
        data = {
            "json_schema": json_schema,
            "document": mime_document.model_dump(),
            "model": model,
            "temperature": temperature,
            "stream": False,
            "text_operations": text_operations or {},
            "modality": modality,
        }

        request_object = DocumentExtractRequest.model_validate(data)
        response = await self._client._request("POST", "/api/v1/documents/extractions", data=data)
        return maybe_parse_to_pydantic(request_object, DocumentExtractResponse.model_validate(response)) 

    async def stream(
        self,
        json_schema: dict[str, Any] | Path | str,
        document: Path | str | IOBase,
        text_operations: Optional[dict[str, Any]] = None,
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        modality: Modality = "native",
    ) -> AsyncIterator[DocumentExtractResponse]:
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
            AsyncIterator[DocumentExtractResponse]: Stream of parsed responses.
        """
        assert_valid_model_extraction(model)

        json_schema = load_json_schema(json_schema)

        mime_document = prepare_mime_document(document)
        data = {
            "json_schema": json_schema,
            "document": mime_document.model_dump(),
            "model": model,
            "temperature": temperature,
            "stream": True,
            "text_operations": text_operations or {},
            "modality": modality,
        }

        request_object = DocumentExtractRequest.model_validate(data)

        async for chunk_json in self._client._request_stream("POST", "/api/v1/documents/extractions", data=data):
            if not chunk_json:
                continue
            
            yield maybe_parse_to_pydantic(request_object, DocumentExtractResponse.model_validate(chunk_json))  
