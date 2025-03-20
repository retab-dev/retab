from io import IOBase
from pathlib import Path
from typing import Any, AsyncGenerator, Generator, Optional
from pydantic import HttpUrl
import base64
from ..._resource import AsyncAPIResource, SyncAPIResource
from ..._utils.ai_models import assert_valid_model_extraction
from ..._utils.json_schema import filter_reasoning_fields_json, load_json_schema
from ..._utils.mime import prepare_mime_document, MIMEData
from ..._utils.stream_context_managers import as_async_context_manager, as_context_manager
from ...types.documents.extractions import DocumentExtractRequest, DocumentExtractResponse, DocumentExtractResponseStream, LogExtractionRequest
from ...types.modalities import Modality
from ...types.standards import PreparedRequest
from ...types.schemas.object import Schema
from ...types.chat import ChatCompletionUiformMessage
from typing import overload

from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from openai.types.chat import ChatCompletionMessageParam
from anthropic.types.message_param import MessageParam

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
        reasoning_effort: ChatCompletionReasoningEffort,
        stream: bool,
        n_consensus: int = 1,
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
            "reasoning_effort": reasoning_effort,
            "n_consensus": n_consensus
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

    def prepare_log_extraction(
        self,
        document: Path | str | IOBase | HttpUrl | None,
        completion: Any,
        json_schema: dict[str, Any],
        model: str,
        temperature: float,
        # The messages can be provided in different formats, we will convert them to the UiForm-compatible format
        messages: list[ChatCompletionUiformMessage] | None = None, 
        openai_messages: list[ChatCompletionMessageParam] | None = None,
        anthropic_messages: list[MessageParam] | None = None,
        anthropic_system_prompt: str | None = None,
    ) -> PreparedRequest:
        if document is None:
            mime_document = MIMEData(
                filename="dummy.txt",
                # url is a base64 encoded string with the mime type and the content. For the dummy one we will send a .txt file with the text "No document provided"
                url="data:text/plain;base64," + base64.b64encode(b"No document provided").decode("utf-8"),
            )

        return PreparedRequest(
            method="POST",
            url="/v1/documents/log_extraction",
            data=LogExtractionRequest(
                document=prepare_mime_document(document) if document else mime_document,
                messages=messages,
                openai_messages=openai_messages,
                anthropic_messages=anthropic_messages,
                anthropic_system_prompt=anthropic_system_prompt,
                completion=completion,
                json_schema=json_schema,
                model=model,
                temperature=temperature,
            ).model_dump(mode="json"),
            raise_for_status=True,
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
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        n_consensus: int = 1,
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
            modality: Modality of the document (e.g., native)
            reasoning_effort: The effort level for the model to reason about the input data.
            n_consensus: Number of consensus extractions to perform (default: 1 which computes a single extraction and the likelihoods comes from the model logprobs)
            idempotency_key: Idempotency key for request
            store: Whether to store the document in the UiForm database
        Returns:
            DocumentAPIResponse
        Raises:
            HTTPException if the request fails
        """

        assert (document is not None), "Either document or messages must be provided"

        # Validate DocumentAPIRequest data (raises exception if invalid)
        request = self.prepare_extraction(json_schema, document, image_settings, model, temperature, modality, reasoning_effort, False, n_consensus=n_consensus, store=store, idempotency_key=idempotency_key)        
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
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        n_consensus: int = 1,
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
            reasoning_effort: The effort level for the model to reason about the input data.
            n_consensus: Number of consensus extractions to perform (default: 1 which computes a single extraction and the likelihoods comes from the model logprobs)
            idempotency_key: Idempotency key for request
            store: Whether to store the document in the UiForm database

        Returns:
            Generator[DocumentExtractResponse]: Stream of parsed responses
        Raises:
            HTTPException if the request fails
        Usage:
        ```python
        with uiform.documents.extractions.stream(json_schema, document, model, temperature, reasoning_effort, modality) as stream:
            for response in stream:
                print(response)
        ```
        """
        request = self.prepare_extraction(json_schema, document, image_settings, model, temperature, modality, reasoning_effort, True, n_consensus=n_consensus, store=store, idempotency_key=idempotency_key)
        schema = Schema(json_schema=load_json_schema(json_schema))

        # Request the stream and return a context manager
        chunk_json: Any = None
        for chunk_json in self._client._prepared_request_stream(request):
            if not chunk_json:
                continue
            yield maybe_parse_to_pydantic(schema, DocumentExtractResponseStream.model_validate(chunk_json), allow_partial=True)
        yield maybe_parse_to_pydantic(schema, DocumentExtractResponseStream.model_validate(chunk_json))

    def log(
        self,
        document: Path | str | IOBase | HttpUrl | None,
        completion: Any,
        json_schema: dict[str, Any],
        model: str,
        temperature: float,
        messages: list[ChatCompletionUiformMessage] | None = None,
        openai_messages: list[ChatCompletionMessageParam] | None = None,
        anthropic_messages: list[MessageParam] | None = None,
        anthropic_system_prompt: str | None = None,
    ) -> None:
        request = self.prepare_log_extraction(
            document,
            completion,
            json_schema,
            model,
            temperature,
            messages=messages,
            openai_messages=openai_messages,
            anthropic_messages=anthropic_messages,
            anthropic_system_prompt=anthropic_system_prompt,
        )
        return self._client._prepared_request(request)

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
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        n_consensus: int = 1,
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
            reasoning_effort: The effort level for the model to reason about the input data.
            n_consensus: Number of consensus extractions to perform (default: 1 which computes a single extraction and the likelihoods comes from the model logprobs)
            idempotency_key: Idempotency key for request
            store: Whether to store the document in the UiForm database
        Returns:
            DocumentExtractResponse: Parsed response from the API.
        """
        request = self.prepare_extraction(json_schema, document, image_settings, model, temperature, modality, reasoning_effort, False, n_consensus=n_consensus, store=store, idempotency_key=idempotency_key)
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
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        n_consensus: int = 1,
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
            reasoning_effort: The effort level for the model to reason about the input data.
            n_consensus: Number of consensus extractions to perform (default: 1 which computes a single extraction and the likelihoods comes from the model logprobs)
            idempotency_key: Idempotency key for request
            store: Whether to store the document in the UiForm database
        Returns:
            AsyncGenerator[DocumentExtractResponse, None]: Stream of parsed responses.

        Usage:
        ```python
        async with uiform.documents.extractions.stream(json_schema, document, model, temperature, reasoning_effort, modality) as stream:
            async for response in stream:
                print(response)
        ```
        """
        request = self.prepare_extraction(json_schema, document, image_settings, model, temperature, modality, reasoning_effort, True, n_consensus=n_consensus, store=store, idempotency_key=idempotency_key)
        schema = Schema(json_schema=load_json_schema(json_schema))
        chunk_json: Any = None
        async for chunk_json in self._client._prepared_request_stream(request):
            if not chunk_json:
                continue
            
            yield maybe_parse_to_pydantic(schema, DocumentExtractResponseStream.model_validate(chunk_json), allow_partial=True)
        # Last chunk with full parsed response
        yield maybe_parse_to_pydantic(schema, DocumentExtractResponseStream.model_validate(chunk_json))

    async def log(
        self,
        document: Path | str | IOBase | HttpUrl | None,
        completion: Any,
        json_schema: dict[str, Any],
        model: str,
        temperature: float,
        messages: list[ChatCompletionUiformMessage] | None = None,
        openai_messages: list[ChatCompletionMessageParam] | None = None,
        anthropic_messages: list[MessageParam] | None = None,
        anthropic_system_prompt: str | None = None,
    ) -> None:
        request = self.prepare_log_extraction(
            document,
            completion,
            json_schema,
            model,
            temperature,
            messages=messages,
            openai_messages=openai_messages,
            anthropic_messages=anthropic_messages,
            anthropic_system_prompt=anthropic_system_prompt,
        )
        return await self._client._prepared_request(request)
