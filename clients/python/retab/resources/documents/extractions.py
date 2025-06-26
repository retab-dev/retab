import base64
import json
from io import IOBase
from pathlib import Path
from typing import Any, AsyncGenerator, Generator

from anthropic.types.message_param import MessageParam
from openai.types.chat import ChatCompletionMessageParam
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from openai.types.chat.parsed_chat_completion import ParsedChatCompletionMessage
from openai.types.responses.response import Response
from openai.types.responses.response_input_param import ResponseInputItemParam
from pydantic_core import PydanticUndefined
from pydantic import HttpUrl

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...utils.ai_models import assert_valid_model_extraction
from ...utils.json_schema import filter_auxiliary_fields_json, load_json_schema, unflatten_dict
from ...utils.mime import MIMEData, prepare_mime_document
from ...utils.stream_context_managers import as_async_context_manager, as_context_manager
from ...types.chat import ChatCompletionRetabMessage
from ...types.documents.extractions import DocumentExtractRequest, LogExtractionRequest, RetabParsedChatCompletion, RetabParsedChatCompletionChunk, RetabParsedChoice
from ...types.browser_canvas import BrowserCanvas
from ...types.modalities import Modality
from ...types.schemas.object import Schema
from ...types.standards import PreparedRequest


def maybe_parse_to_pydantic(schema: Schema, response: RetabParsedChatCompletion, allow_partial: bool = False) -> RetabParsedChatCompletion:
    if response.choices[0].message.content:
        try:
            if allow_partial:
                response.choices[0].message.parsed = schema._partial_pydantic_model.model_validate(filter_auxiliary_fields_json(response.choices[0].message.content))
            else:
                response.choices[0].message.parsed = schema.pydantic_model.model_validate(filter_auxiliary_fields_json(response.choices[0].message.content))
        except Exception:
            pass
    return response


class BaseExtractionsMixin:
    def prepare_extraction(
        self,
        json_schema: dict[str, Any] | Path | str,
        document: Path | str | IOBase | HttpUrl | None = None,
        documents: list[Path | str | IOBase | HttpUrl] | None = None,
        image_resolution_dpi: int = PydanticUndefined,  # type: ignore[assignment]
        browser_canvas: BrowserCanvas = PydanticUndefined,  # type: ignore[assignment]
        model: str = PydanticUndefined,  # type: ignore[assignment]
        temperature: float = PydanticUndefined,  # type: ignore[assignment]
        modality: Modality = PydanticUndefined,  # type: ignore[assignment]
        reasoning_effort: ChatCompletionReasoningEffort = PydanticUndefined,  # type: ignore[assignment]
        stream: bool = False,
        n_consensus: int = PydanticUndefined,  # type: ignore[assignment]
        store: bool = False,
        idempotency_key: str | None = None,
    ) -> PreparedRequest:
        assert_valid_model_extraction(model)

        json_schema = load_json_schema(json_schema)

        # Handle both single document and multiple documents
        if document is not None and documents is not None:
            raise ValueError("Cannot provide both 'document' and 'documents' parameters. Use either one.")

        # Convert single document to documents list for consistency
        if document is not None:
            processed_documents = [prepare_mime_document(document)]
        elif documents is not None:
            processed_documents = [prepare_mime_document(doc) for doc in documents]
        else:
            raise ValueError("Must provide either 'document' or 'documents' parameter.")

        # Validate DocumentAPIRequest data (raises exception if invalid)
        request = DocumentExtractRequest(
            json_schema=json_schema,
            documents=processed_documents,
            model=model,
            temperature=temperature,
            stream=stream,
            modality=modality,
            store=store,
            reasoning_effort=reasoning_effort,
            n_consensus=n_consensus,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
        )

        return PreparedRequest(
            method="POST", url="/v1/documents/extractions", data=request.model_dump(mode="json", exclude_unset=True, exclude_defaults=True), idempotency_key=idempotency_key
        )

    def prepare_log_extraction(
        self,
        document: Path | str | IOBase | HttpUrl | None,
        json_schema: dict[str, Any],
        model: str,
        temperature: float,
        completion: Any | None = None,
        # The messages can be provided in different formats, we will convert them to the Retab-compatible format
        messages: list[ChatCompletionRetabMessage] | None = None,
        openai_messages: list[ChatCompletionMessageParam] | None = None,
        anthropic_messages: list[MessageParam] | None = None,
        anthropic_system_prompt: str | None = None,
        # New fields for the Responses API
        openai_responses_input: list[ResponseInputItemParam] | None = None,
        openai_responses_output: Response | None = None,
    ) -> PreparedRequest:
        if document is None:
            mime_document = MIMEData(
                filename="dummy.txt",
                # url is a base64 encoded string with the mime type and the content. For the dummy one we will send a .txt file with the text "No document provided"
                url="data:text/plain;base64," + base64.b64encode(b"No document provided").decode("utf-8"),
            )
        else:
            mime_document = prepare_mime_document(document)

        return PreparedRequest(
            method="POST",
            url="/v1/documents/log_extraction",
            data=LogExtractionRequest(
                document=mime_document,
                messages=messages,
                openai_messages=openai_messages,
                anthropic_messages=anthropic_messages,
                anthropic_system_prompt=anthropic_system_prompt,
                completion=completion,
                openai_responses_input=openai_responses_input,
                openai_responses_output=openai_responses_output,
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
        model: str,
        document: Path | str | IOBase | HttpUrl | None = None,
        documents: list[Path | str | IOBase | HttpUrl] | None = None,
        image_resolution_dpi: int = PydanticUndefined,  # type: ignore[assignment]
        browser_canvas: BrowserCanvas = PydanticUndefined,  # type: ignore[assignment]
        temperature: float = PydanticUndefined,  # type: ignore[assignment]
        modality: Modality = PydanticUndefined,  # type: ignore[assignment]
        reasoning_effort: ChatCompletionReasoningEffort = PydanticUndefined,  # type: ignore[assignment]
        n_consensus: int = PydanticUndefined,  # type: ignore[assignment]
        idempotency_key: str | None = None,
        store: bool = False,
    ) -> RetabParsedChatCompletion:
        """
        Process one or more documents using the Retab API.

        Args:
            json_schema: JSON schema defining the expected data structure
            model: The AI model to use for processing
            document: Single document to process (use either this or documents, not both)
            documents: List of documents to process (use either this or document, not both)
            image_resolution_dpi: Optional image resolution DPI
            browser_canvas: Optional browser canvas size
            temperature: Model temperature setting (0-1)
            modality: Modality of the document (e.g., native)
            reasoning_effort: The effort level for the model to reason about the input data.
            n_consensus: Number of consensus extractions to perform (default: 1 which computes a single extraction and the likelihoods comes from the model logprobs)
            idempotency_key: Idempotency key for request
            store: Whether to store the document in the Retab database
        Returns:
            RetabParsedChatCompletion: Parsed response from the API
        Raises:
            ValueError: If neither document nor documents is provided, or if both are provided
            HTTPException: If the request fails
        """

        # Validate DocumentAPIRequest data (raises exception if invalid)
        request = self.prepare_extraction(
            json_schema=json_schema,
            document=document,
            documents=documents,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            model=model,
            temperature=temperature,
            modality=modality,
            reasoning_effort=reasoning_effort,
            stream=False,
            n_consensus=n_consensus,
            store=store,
            idempotency_key=idempotency_key,
        )
        response = self._client._prepared_request(request)

        schema = Schema(json_schema=load_json_schema(json_schema))
        return maybe_parse_to_pydantic(schema, RetabParsedChatCompletion.model_validate(response))

    @as_context_manager
    def stream(
        self,
        json_schema: dict[str, Any] | Path | str,
        model: str,
        document: Path | str | IOBase | HttpUrl | None = None,
        documents: list[Path | str | IOBase | HttpUrl] | None = None,
        image_resolution_dpi: int = PydanticUndefined,  # type: ignore[assignment]
        browser_canvas: BrowserCanvas = PydanticUndefined,  # type: ignore[assignment]
        temperature: float = PydanticUndefined,  # type: ignore[assignment]
        modality: Modality = PydanticUndefined,  # type: ignore[assignment]
        reasoning_effort: ChatCompletionReasoningEffort = PydanticUndefined,  # type: ignore[assignment]
        n_consensus: int = PydanticUndefined,  # type: ignore[assignment]
        idempotency_key: str | None = None,
        store: bool = False,
    ) -> Generator[RetabParsedChatCompletion, None, None]:
        """
        Process one or more documents using the Retab API with streaming enabled.

        Args:
            json_schema: JSON schema defining the expected data structure
            model: The AI model to use for processing
            document: Single document to process (use either this or documents, not both)
            documents: List of documents to process (use either this or document, not both)
            image_resolution_dpi: Optional image resolution DPI.
            browser_canvas: Optional browser canvas size.
            temperature: Model temperature setting (0-1)
            modality: Modality of the document (e.g., native)
            reasoning_effort: The effort level for the model to reason about the input data.
            n_consensus: Number of consensus extractions to perform (default: 1 which computes a single extraction and the likelihoods comes from the model logprobs)
            idempotency_key: Idempotency key for request
            store: Whether to store the document in the Retab database

        Returns:
            Generator[RetabParsedChatCompletion]: Stream of parsed responses
        Raises:
            ValueError: If neither document nor documents is provided, or if both are provided
            HTTPException: If the request fails
        Usage:
        ```python
        # Single document
        with retab.documents.extractions.stream(json_schema, model, document=document) as stream:
            for response in stream:
                print(response)

        # Multiple documents
        with retab.documents.extractions.stream(json_schema, model, documents=[doc1, doc2]) as stream:
            for response in stream:
                print(response)
        ```
        """
        request = self.prepare_extraction(
            json_schema=json_schema,
            document=document,
            documents=documents,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            model=model,
            temperature=temperature,
            modality=modality,
            reasoning_effort=reasoning_effort,
            stream=True,
            n_consensus=n_consensus,
            store=store,
            idempotency_key=idempotency_key,
        )
        schema = Schema(json_schema=load_json_schema(json_schema))

        # Request the stream and return a context manager
        ui_parsed_chat_completion_cum_chunk: RetabParsedChatCompletionChunk | None = None
        # Initialize the RetabParsedChatCompletion object
        ui_parsed_completion: RetabParsedChatCompletion = RetabParsedChatCompletion(
            id="",
            created=0,
            model="",
            object="chat.completion",
            likelihoods={},
            choices=[
                RetabParsedChoice(
                    index=0,
                    message=ParsedChatCompletionMessage(content="", role="assistant"),
                    finish_reason=None,
                    logprobs=None,
                )
            ],
        )
        for chunk_json in self._client._prepared_request_stream(request):
            if not chunk_json:
                continue
            ui_parsed_chat_completion_cum_chunk = RetabParsedChatCompletionChunk.model_validate(chunk_json).chunk_accumulator(ui_parsed_chat_completion_cum_chunk)
            # Basic stuff
            ui_parsed_completion.id = ui_parsed_chat_completion_cum_chunk.id
            ui_parsed_completion.created = ui_parsed_chat_completion_cum_chunk.created
            ui_parsed_completion.model = ui_parsed_chat_completion_cum_chunk.model
            # Update the ui_parsed_completion object
            parsed = unflatten_dict(ui_parsed_chat_completion_cum_chunk.choices[0].delta.flat_parsed)
            likelihoods = unflatten_dict(ui_parsed_chat_completion_cum_chunk.choices[0].delta.flat_likelihoods)
            ui_parsed_completion.choices[0].message.content = json.dumps(parsed)
            ui_parsed_completion.choices[0].message.parsed = parsed
            ui_parsed_completion.likelihoods = likelihoods

            yield maybe_parse_to_pydantic(schema, ui_parsed_completion, allow_partial=True)

        # change the finish_reason to stop
        ui_parsed_completion.choices[0].finish_reason = "stop"
        yield maybe_parse_to_pydantic(schema, ui_parsed_completion)

    def log(
        self,
        document: Path | str | IOBase | HttpUrl | None,
        json_schema: dict[str, Any],
        model: str,
        temperature: float,
        completion: Any | None = None,
        # The messages can be provided in different formats, we will convert them to the Retab-compatible format
        messages: list[ChatCompletionRetabMessage] | None = None,
        openai_messages: list[ChatCompletionMessageParam] | None = None,
        anthropic_messages: list[MessageParam] | None = None,
        anthropic_system_prompt: str | None = None,
        # New fields for the Responses API
        openai_responses_input: list[ResponseInputItemParam] | None = None,
        openai_responses_output: Response | None = None,
    ) -> None:
        request = self.prepare_log_extraction(
            document=document,
            json_schema=json_schema,
            model=model,
            temperature=temperature,
            completion=completion,
            messages=messages,
            openai_messages=openai_messages,
            anthropic_messages=anthropic_messages,
            anthropic_system_prompt=anthropic_system_prompt,
            openai_responses_input=openai_responses_input,
            openai_responses_output=openai_responses_output,
        )
        return self._client._prepared_request(request)


class AsyncExtractions(AsyncAPIResource, BaseExtractionsMixin):
    """Extraction API wrapper for asynchronous usage."""

    async def parse(
        self,
        json_schema: dict[str, Any] | Path | str,
        model: str,
        document: Path | str | IOBase | HttpUrl | None = None,
        documents: list[Path | str | IOBase | HttpUrl] | None = None,
        image_resolution_dpi: int = PydanticUndefined,  # type: ignore[assignment]
        browser_canvas: BrowserCanvas = PydanticUndefined,  # type: ignore[assignment]
        temperature: float = PydanticUndefined,  # type: ignore[assignment]
        modality: Modality = PydanticUndefined,  # type: ignore[assignment]
        reasoning_effort: ChatCompletionReasoningEffort = PydanticUndefined,  # type: ignore[assignment]
        n_consensus: int = PydanticUndefined,  # type: ignore[assignment]
        idempotency_key: str | None = None,
        store: bool = False,
    ) -> RetabParsedChatCompletion:
        """
        Extract structured data from one or more documents asynchronously.

        Args:
            json_schema: JSON schema defining the expected data structure.
            model: The AI model to use.
            document: Single document to process (use either this or documents, not both)
            documents: List of documents to process (use either this or document, not both)
            image_resolution_dpi: Optional image resolution DPI.
            browser_canvas: Optional browser canvas size.
            temperature: Model temperature setting (0-1).
            modality: Modality of the document (e.g., native).
            reasoning_effort: The effort level for the model to reason about the input data.
            n_consensus: Number of consensus extractions to perform (default: 1 which computes a single extraction and the likelihoods comes from the model logprobs)
            idempotency_key: Idempotency key for request
            store: Whether to store the document in the Retab database
        Returns:
            RetabParsedChatCompletion: Parsed response from the API.
        Raises:
            ValueError: If neither document nor documents is provided, or if both are provided
        """
        request = self.prepare_extraction(
            json_schema=json_schema,
            document=document,
            documents=documents,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            model=model,
            temperature=temperature,
            modality=modality,
            reasoning_effort=reasoning_effort,
            stream=False,
            n_consensus=n_consensus,
            store=store,
            idempotency_key=idempotency_key,
        )
        response = await self._client._prepared_request(request)
        schema = Schema(json_schema=load_json_schema(json_schema))
        return maybe_parse_to_pydantic(schema, RetabParsedChatCompletion.model_validate(response))

    @as_async_context_manager
    async def stream(
        self,
        json_schema: dict[str, Any] | Path | str,
        model: str,
        document: Path | str | IOBase | HttpUrl | None = None,
        documents: list[Path | str | IOBase | HttpUrl] | None = None,
        image_resolution_dpi: int = PydanticUndefined,  # type: ignore[assignment]
        browser_canvas: BrowserCanvas = PydanticUndefined,  # type: ignore[assignment]
        temperature: float = PydanticUndefined,  # type: ignore[assignment]
        modality: Modality = PydanticUndefined,  # type: ignore[assignment]
        reasoning_effort: ChatCompletionReasoningEffort = PydanticUndefined,  # type: ignore[assignment]
        n_consensus: int = PydanticUndefined,  # type: ignore[assignment]
        idempotency_key: str | None = None,
        store: bool = False,
    ) -> AsyncGenerator[RetabParsedChatCompletion, None]:
        """
        Extract structured data from one or more documents asynchronously with streaming.

        Args:
            json_schema: JSON schema defining the expected data structure.
            model: The AI model to use.
            document: Single document to process (use either this or documents, not both)
            documents: List of documents to process (use either this or document, not both)
            image_resolution_dpi: Optional image resolution DPI.
            browser_canvas: Optional browser canvas size.
            temperature: Model temperature setting (0-1).
            modality: Modality of the document (e.g., native).
            reasoning_effort: The effort level for the model to reason about the input data.
            n_consensus: Number of consensus extractions to perform (default: 1 which computes a single extraction and the likelihoods comes from the model logprobs)
            idempotency_key: Idempotency key for request
            store: Whether to store the document in the Retab database
        Returns:
            AsyncGenerator[RetabParsedChatCompletion, None]: Stream of parsed responses.
        Raises:
            ValueError: If neither document nor documents is provided, or if both are provided

        Usage:
        ```python
        # Single document
        async with retab.documents.extractions.stream(json_schema, model, document=document) as stream:
            async for response in stream:
                print(response)

        # Multiple documents
        async with retab.documents.extractions.stream(json_schema, model, documents=[doc1, doc2]) as stream:
            async for response in stream:
                print(response)
        ```
        """
        request = self.prepare_extraction(
            json_schema=json_schema,
            document=document,
            documents=documents,
            image_resolution_dpi=image_resolution_dpi,
            browser_canvas=browser_canvas,
            model=model,
            temperature=temperature,
            modality=modality,
            reasoning_effort=reasoning_effort,
            stream=True,
            n_consensus=n_consensus,
            store=store,
            idempotency_key=idempotency_key,
        )
        schema = Schema(json_schema=load_json_schema(json_schema))
        ui_parsed_chat_completion_cum_chunk: RetabParsedChatCompletionChunk | None = None
        # Initialize the RetabParsedChatCompletion object
        ui_parsed_completion: RetabParsedChatCompletion = RetabParsedChatCompletion(
            id="",
            created=0,
            model="",
            object="chat.completion",
            likelihoods={},
            choices=[
                RetabParsedChoice(
                    index=0,
                    message=ParsedChatCompletionMessage(content="", role="assistant"),
                    finish_reason=None,
                    logprobs=None,
                )
            ],
        )

        async for chunk_json in self._client._prepared_request_stream(request):
            if not chunk_json:
                continue
            ui_parsed_chat_completion_cum_chunk = RetabParsedChatCompletionChunk.model_validate(chunk_json).chunk_accumulator(ui_parsed_chat_completion_cum_chunk)
            # Basic stuff
            ui_parsed_completion.id = ui_parsed_chat_completion_cum_chunk.id
            ui_parsed_completion.created = ui_parsed_chat_completion_cum_chunk.created
            ui_parsed_completion.model = ui_parsed_chat_completion_cum_chunk.model

            # Update the ui_parsed_completion object
            parsed = unflatten_dict(ui_parsed_chat_completion_cum_chunk.choices[0].delta.flat_parsed)
            likelihoods = unflatten_dict(ui_parsed_chat_completion_cum_chunk.choices[0].delta.flat_likelihoods)
            ui_parsed_completion.choices[0].message.content = json.dumps(parsed)
            ui_parsed_completion.choices[0].message.parsed = parsed
            ui_parsed_completion.likelihoods = likelihoods

            yield maybe_parse_to_pydantic(schema, ui_parsed_completion, allow_partial=True)

        # change the finish_reason to stop
        ui_parsed_completion.choices[0].finish_reason = "stop"
        yield maybe_parse_to_pydantic(schema, ui_parsed_completion)

    async def log(
        self,
        document: Path | str | IOBase | HttpUrl | None,
        json_schema: dict[str, Any],
        model: str,
        temperature: float,
        completion: Any | None = None,
        # The messages can be provided in different formats, we will convert them to the Retab-compatible format
        messages: list[ChatCompletionRetabMessage] | None = None,
        openai_messages: list[ChatCompletionMessageParam] | None = None,
        anthropic_messages: list[MessageParam] | None = None,
        anthropic_system_prompt: str | None = None,
        # New fields for the Responses API
        openai_responses_input: list[ResponseInputItemParam] | None = None,
        openai_responses_output: Response | None = None,
    ) -> None:
        request = self.prepare_log_extraction(
            document=document,
            json_schema=json_schema,
            model=model,
            temperature=temperature,
            completion=completion,
            messages=messages,
            openai_messages=openai_messages,
            anthropic_messages=anthropic_messages,
            anthropic_system_prompt=anthropic_system_prompt,
            openai_responses_input=openai_responses_input,
            openai_responses_output=openai_responses_output,
        )
        return await self._client._prepared_request(request)
