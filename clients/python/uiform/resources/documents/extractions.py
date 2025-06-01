import base64
import json
from io import IOBase
from pathlib import Path
from typing import Any, AsyncGenerator, Generator, Literal, Optional

from anthropic.types.message_param import MessageParam
from openai.types.chat import ChatCompletionMessageParam
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from openai.types.chat.parsed_chat_completion import ParsedChatCompletionMessage
from openai.types.responses.response import Response
from openai.types.responses.response_input_param import ResponseInputItemParam
from pydantic import HttpUrl

from ..._resource import AsyncAPIResource, SyncAPIResource
from ..._utils.ai_models import assert_valid_model_extraction
from ..._utils.json_schema import filter_auxiliary_fields_json, load_json_schema, unflatten_dict
from ..._utils.mime import MIMEData, prepare_mime_document
from ..._utils.stream_context_managers import as_async_context_manager, as_context_manager
from ...types.chat import ChatCompletionUiformMessage
from ...types.documents.extractions import DocumentExtractRequest, LogExtractionRequest, UiParsedChatCompletion, UiParsedChatCompletionChunk, UiParsedChoice
from ...types.modalities import Modality
from ...types.schemas.object import Schema
from ...types.standards import PreparedRequest


def maybe_parse_to_pydantic(schema: Schema, response: UiParsedChatCompletion, allow_partial: bool = False) -> UiParsedChatCompletion:
    if response.choices[0].message.content:
        try:
            if allow_partial:
                response.choices[0].message.parsed = schema._partial_pydantic_model.model_validate(filter_auxiliary_fields_json(response.choices[0].message.content))
            else:
                response.choices[0].message.parsed = schema.pydantic_model.model_validate(filter_auxiliary_fields_json(response.choices[0].message.content))
        except Exception as e:
            pass
    return response


class BaseExtractionsMixin:
    def prepare_extraction(
        self,
        json_schema: dict[str, Any] | Path | str,
        document: Path | str | IOBase | HttpUrl | None,
        image_resolution_dpi: int | None,
        browser_canvas: Literal['A3', 'A4', 'A5'] | None,
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
            "n_consensus": n_consensus,
        }
        if image_resolution_dpi:
            data["image_resolution_dpi"] = image_resolution_dpi
        if browser_canvas:
            data["browser_canvas"] = browser_canvas

        # Validate DocumentAPIRequest data (raises exception if invalid)
        document_extract_request = DocumentExtractRequest.model_validate(data)

        return PreparedRequest(method="POST", url="/v1/documents/extractions", data=document_extract_request.model_dump(), idempotency_key=idempotency_key)

    def prepare_log_extraction(
        self,
        document: Path | str | IOBase | HttpUrl | None,
        json_schema: dict[str, Any],
        model: str,
        temperature: float,
        completion: Any | None = None,
        # The messages can be provided in different formats, we will convert them to the UiForm-compatible format
        messages: list[ChatCompletionUiformMessage] | None = None,
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
                openai_responses_input=openai_responses_input,
                openai_responses_output=openai_responses_output,
                json_schema=json_schema,
                model=model,
                temperature=temperature,
            ).model_dump(mode="json", by_alias=True),  # by_alias is necessary to enable serialization/deserialization ('schema' was being converted to 'schema_')
            raise_for_status=True,
        )


class Extractions(SyncAPIResource, BaseExtractionsMixin):
    """Extraction API wrapper"""

    def parse(
        self,
        json_schema: dict[str, Any] | Path | str,
        model: str,
        document: Path | str | IOBase | HttpUrl | None,
        image_resolution_dpi: int | None = None,
        browser_canvas: Literal['A3', 'A4', 'A5'] | None = None,
        temperature: float = 0,
        modality: Modality = "native",
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        n_consensus: int = 1,
        idempotency_key: str | None = None,
        store: bool = False,
    ) -> UiParsedChatCompletion:
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

        assert document is not None, "Either document or messages must be provided"

        # Validate DocumentAPIRequest data (raises exception if invalid)
        request = self.prepare_extraction(
            json_schema, document, image_resolution_dpi, browser_canvas, model, temperature, modality, reasoning_effort, False, n_consensus=n_consensus, store=store, idempotency_key=idempotency_key
        )
        response = self._client._prepared_request(request)

        schema = Schema(json_schema=load_json_schema(json_schema))
        return maybe_parse_to_pydantic(schema, UiParsedChatCompletion.model_validate(response))

    @as_context_manager
    def stream(
        self,
        json_schema: dict[str, Any] | Path | str,
        model: str,
        document: Path | str | IOBase | HttpUrl | None,
        image_resolution_dpi: int | None = None,
        browser_canvas: Literal['A3', 'A4', 'A5'] | None = None,
        temperature: float = 0,
        modality: Modality = "native",
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        n_consensus: int = 1,
        idempotency_key: str | None = None,
        store: bool = False,
    ) -> Generator[UiParsedChatCompletion, None, None]:
        """
        Process a document using the UiForm API with streaming enabled.

        Args:
            json_schema: JSON schema defining the expected data structure
            document: Single document (as MIMEData) to process
            image_resolution_dpi: Optional image resolution DPI.
            browser_canvas: Optional browser canvas size.
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
        request = self.prepare_extraction(
            json_schema, document, image_resolution_dpi, browser_canvas, model, temperature, modality, reasoning_effort, True, n_consensus=n_consensus, store=store, idempotency_key=idempotency_key
        )
        schema = Schema(json_schema=load_json_schema(json_schema))

        # Request the stream and return a context manager
        ui_parsed_chat_completion_cum_chunk: UiParsedChatCompletionChunk | None = None
        # Initialize the UiParsedChatCompletion object
        ui_parsed_completion: UiParsedChatCompletion = UiParsedChatCompletion(
            id="",
            created=0,
            model="",
            object="chat.completion",
            likelihoods={},
            choices=[
                UiParsedChoice(
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
            ui_parsed_chat_completion_cum_chunk = UiParsedChatCompletionChunk.model_validate(chunk_json).chunk_accumulator(ui_parsed_chat_completion_cum_chunk)
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
        # The messages can be provided in different formats, we will convert them to the UiForm-compatible format
        messages: list[ChatCompletionUiformMessage] | None = None,
        openai_messages: list[ChatCompletionMessageParam] | None = None,
        anthropic_messages: list[MessageParam] | None = None,
        anthropic_system_prompt: str | None = None,
        # New fields for the Responses API
        openai_responses_input: list[ResponseInputItemParam] | None = None,
        openai_responses_output: Response | None = None,
    ) -> None:
        request = self.prepare_log_extraction(
            document,
            json_schema,
            model,
            temperature,
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
        document: Path | str | IOBase | HttpUrl | None,
        image_resolution_dpi: int | None = None,
        browser_canvas: Literal['A3', 'A4', 'A5'] | None = None,
        temperature: float = 0,
        modality: Modality = "native",
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        n_consensus: int = 1,
        idempotency_key: str | None = None,
        store: bool = False,
    ) -> UiParsedChatCompletion:
        """
        Extract structured data from a document asynchronously.

        Args:
            json_schema: JSON schema defining the expected data structure.
            document: Path, string, or file-like object representing the document.
            image_resolution_dpi: Optional image resolution DPI.
            browser_canvas: Optional browser canvas size.
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
        request = self.prepare_extraction(
            json_schema, document, image_resolution_dpi, browser_canvas, model, temperature, modality, reasoning_effort, False, n_consensus=n_consensus, store=store, idempotency_key=idempotency_key
        )
        response = await self._client._prepared_request(request)
        schema = Schema(json_schema=load_json_schema(json_schema))
        return maybe_parse_to_pydantic(schema, UiParsedChatCompletion.model_validate(response))

    @as_async_context_manager
    async def stream(
        self,
        json_schema: dict[str, Any] | Path | str,
        model: str,
        document: Path | str | IOBase | HttpUrl | None,
        image_resolution_dpi: int | None = None,
        browser_canvas: Literal['A3', 'A4', 'A5'] | None = None,
        temperature: float = 0,
        modality: Modality = "native",
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        n_consensus: int = 1,
        idempotency_key: str | None = None,
        store: bool = False,
    ) -> AsyncGenerator[UiParsedChatCompletion, None]:
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
        request = self.prepare_extraction(
            json_schema, document, image_resolution_dpi, browser_canvas, model, temperature, modality, reasoning_effort, True, n_consensus=n_consensus, store=store, idempotency_key=idempotency_key
        )
        schema = Schema(json_schema=load_json_schema(json_schema))
        ui_parsed_chat_completion_cum_chunk: UiParsedChatCompletionChunk | None = None
        # Initialize the UiParsedChatCompletion object
        ui_parsed_completion: UiParsedChatCompletion = UiParsedChatCompletion(
            id="",
            created=0,
            model="",
            object="chat.completion",
            likelihoods={},
            choices=[
                UiParsedChoice(
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
            ui_parsed_chat_completion_cum_chunk = UiParsedChatCompletionChunk.model_validate(chunk_json).chunk_accumulator(ui_parsed_chat_completion_cum_chunk)
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
        # The messages can be provided in different formats, we will convert them to the UiForm-compatible format
        messages: list[ChatCompletionUiformMessage] | None = None,
        openai_messages: list[ChatCompletionMessageParam] | None = None,
        anthropic_messages: list[MessageParam] | None = None,
        anthropic_system_prompt: str | None = None,
        # New fields for the Responses API
        openai_responses_input: list[ResponseInputItemParam] | None = None,
        openai_responses_output: Response | None = None,
    ) -> None:
        request = self.prepare_log_extraction(
            document,
            json_schema,
            model,
            temperature,
            completion=completion,
            messages=messages,
            openai_messages=openai_messages,
            anthropic_messages=anthropic_messages,
            anthropic_system_prompt=anthropic_system_prompt,
            openai_responses_input=openai_responses_input,
            openai_responses_output=openai_responses_output,
        )
        return await self._client._prepared_request(request)
