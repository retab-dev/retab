import json
from pathlib import Path
from typing import Any, AsyncGenerator, Generator

from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from openai.types.chat.parsed_chat_completion import ParsedChatCompletionMessage

from .._resource import AsyncAPIResource, SyncAPIResource
from .._utils.ai_models import assert_valid_model_extraction
from .._utils.json_schema import load_json_schema, unflatten_dict
from .._utils.stream_context_managers import as_async_context_manager, as_context_manager
from ..types.chat import ChatCompletionUiformMessage
from ..types.completions import UiChatCompletionsParseRequest
from ..types.documents.extractions import UiParsedChatCompletion, UiParsedChatCompletionChunk, UiParsedChoice
from ..types.standards import PreparedRequest


class BaseCompletionsMixin:
    def prepare(
        self,
        json_schema: dict[str, Any] | Path | str,
        messages: list[ChatCompletionUiformMessage],
        model: str,
        temperature: float,
        reasoning_effort: ChatCompletionReasoningEffort,
        stream: bool,
        idempotency_key: str | None = None,
    ) -> PreparedRequest:
        assert_valid_model_extraction(model)

        json_schema = load_json_schema(json_schema)

        data = {
            "messages": messages,
            "json_schema": json_schema,
            "model": model,
            "temperature": temperature,
            "stream": stream,
            "reasoning_effort": reasoning_effort,
        }

        # Validate DocumentAPIRequest data (raises exception if invalid)
        ui_chat_completions_parse_request = UiChatCompletionsParseRequest.model_validate(data)

        return PreparedRequest(method="POST", url="/v1/completions/parse", data=ui_chat_completions_parse_request.model_dump(), idempotency_key=idempotency_key)


class Completions(SyncAPIResource, BaseCompletionsMixin):
    """Multi-provider Completions API wrapper"""

    def parse(
        self,
        json_schema: dict[str, Any] | Path | str,
        messages: list[ChatCompletionUiformMessage],
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        idempotency_key: str | None = None,
    ) -> UiParsedChatCompletion:
        """
        Parse messages using the UiForm API to extract structured data according to the provided JSON schema.

        Args:
            json_schema: JSON schema defining the expected data structure
            messages: List of chat messages to parse
            model: The AI model to use for processing
            temperature: Model temperature setting (0-1)
            reasoning_effort: The effort level for the model to reason about the input data
            idempotency_key: Idempotency key for request
            store: Whether to store the data in the UiForm database

        Returns:
            UiParsedChatCompletion: Parsed response from the API
        """
        request = self.prepare(
            json_schema=json_schema, messages=messages, model=model, temperature=temperature, reasoning_effort=reasoning_effort, stream=False, idempotency_key=idempotency_key
        )
        response = self._client._prepared_request(request)

        return UiParsedChatCompletion.model_validate(response)

    @as_context_manager
    def stream(
        self,
        json_schema: dict[str, Any] | Path | str,
        messages: list[ChatCompletionUiformMessage],
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        idempotency_key: str | None = None,
    ) -> Generator[UiParsedChatCompletion, None, None]:
        """
        Process messages using the UiForm API with streaming enabled.

        Args:
            json_schema: JSON schema defining the expected data structure
            messages: List of chat messages to parse
            model: The AI model to use for processing
            temperature: Model temperature setting (0-1)
            reasoning_effort: The effort level for the model to reason about the input data
            idempotency_key: Idempotency key for request

        Returns:
            Generator[UiParsedChatCompletion]: Stream of parsed responses

        Usage:
        ```python
        with uiform.completions.stream(json_schema, messages, model, temperature, reasoning_effort) as stream:
            for response in stream:
                print(response)
        ```
        """
        request = self.prepare(
            json_schema=json_schema, messages=messages, model=model, temperature=temperature, reasoning_effort=reasoning_effort, stream=True, idempotency_key=idempotency_key
        )

        # Request the stream and return a context manager
        flatten_parsed: dict[str, Any] = {}
        flatten_likelihoods: dict[str, float] = {}
        ui_parsed_chat_completion_chunk: UiParsedChatCompletionChunk | None = None
        # Initialize the UiParsedChatCompletion object
        ui_parsed_completion: UiParsedChatCompletion = UiParsedChatCompletion(
            id="",
            created=0,
            model="",
            object="chat.completion",
            likelihoods=unflatten_dict(flatten_likelihoods),
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
            ui_parsed_chat_completion_chunk = UiParsedChatCompletionChunk.model_validate(chunk_json)
            # Basic stuff
            ui_parsed_completion.id = ui_parsed_chat_completion_chunk.id
            ui_parsed_completion.created = ui_parsed_chat_completion_chunk.created
            ui_parsed_completion.model = ui_parsed_chat_completion_chunk.model

            # Accumulate the content and likelihoods
            if ui_parsed_chat_completion_chunk.choices:
                flatten_parsed = {**flatten_parsed, **ui_parsed_chat_completion_chunk.choices[0].delta.flat_parsed}
                flatten_likelihoods = {**flatten_likelihoods, **ui_parsed_chat_completion_chunk.choices[0].delta.flat_likelihoods}

            # Update the ui_parsed_completion object
            parsed = unflatten_dict(flatten_parsed)
            ui_parsed_completion.choices[0].message.content = json.dumps(parsed)
            ui_parsed_completion.choices[0].message.parsed = parsed
            ui_parsed_completion.likelihoods = unflatten_dict(flatten_likelihoods)

            yield ui_parsed_completion

        # change the finish_reason to stop
        ui_parsed_completion.choices[0].finish_reason = "stop"
        yield ui_parsed_completion


class AsyncCompletions(AsyncAPIResource, BaseCompletionsMixin):
    """Multi-provider Completions API wrapper for asynchronous usage."""

    async def parse(
        self,
        json_schema: dict[str, Any] | Path | str,
        messages: list[ChatCompletionUiformMessage],
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        idempotency_key: str | None = None,
    ) -> UiParsedChatCompletion:
        """
        Parse messages using the UiForm API asynchronously.

        Args:
            json_schema: JSON schema defining the expected data structure
            messages: List of chat messages to parse
            model: The AI model to use
            temperature: Model temperature setting (0-1)
            reasoning_effort: The effort level for the model to reason about the input data
            idempotency_key: Idempotency key for request

        Returns:
            UiParsedChatCompletion: Parsed response from the API
        """
        request = self.prepare(
            json_schema=json_schema, messages=messages, model=model, temperature=temperature, reasoning_effort=reasoning_effort, stream=False, idempotency_key=idempotency_key
        )
        response = await self._client._prepared_request(request)
        return UiParsedChatCompletion.model_validate(response)

    @as_async_context_manager
    async def stream(
        self,
        json_schema: dict[str, Any] | Path | str,
        messages: list[ChatCompletionUiformMessage],
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        idempotency_key: str | None = None,
    ) -> AsyncGenerator[UiParsedChatCompletion, None]:
        """
        Parse messages using the UiForm API asynchronously with streaming.

        Args:
            json_schema: JSON schema defining the expected data structure
            messages: List of chat messages to parse
            model: The AI model to use
            temperature: Model temperature setting (0-1)
            reasoning_effort: The effort level for the model to reason about the input data
            idempotency_key: Idempotency key for request

        Returns:
            AsyncGenerator[UiParsedChatCompletion]: Stream of parsed responses

        Usage:
        ```python
        async with uiform.completions.stream(json_schema, messages, model, temperature, reasoning_effort) as stream:
            async for response in stream:
                print(response)
        ```
        """
        request = self.prepare(
            json_schema=json_schema, messages=messages, model=model, temperature=temperature, reasoning_effort=reasoning_effort, stream=True, idempotency_key=idempotency_key
        )

        # Request the stream and return a context manager
        flatten_parsed: dict[str, Any] = {}
        flatten_likelihoods: dict[str, float] = {}
        ui_parsed_chat_completion_chunk: UiParsedChatCompletionChunk | None = None
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
            ui_parsed_chat_completion_chunk = UiParsedChatCompletionChunk.model_validate(chunk_json)
            # Basic stuff
            ui_parsed_completion.id = ui_parsed_chat_completion_chunk.id
            ui_parsed_completion.created = ui_parsed_chat_completion_chunk.created
            ui_parsed_completion.model = ui_parsed_chat_completion_chunk.model

            # Accumulate the content and likelihoods
            if ui_parsed_chat_completion_chunk.choices:
                flatten_parsed = {**flatten_parsed, **ui_parsed_chat_completion_chunk.choices[0].delta.flat_parsed}
                flatten_likelihoods = {**flatten_likelihoods, **ui_parsed_chat_completion_chunk.choices[0].delta.flat_likelihoods}

            # Update the ui_parsed_completion object
            parsed = unflatten_dict(flatten_parsed)
            ui_parsed_completion.choices[0].message.parsed = parsed
            ui_parsed_completion.choices[0].message.content = json.dumps(parsed)
            ui_parsed_completion.likelihoods = unflatten_dict(flatten_likelihoods)

            yield ui_parsed_completion

        # change the finish_reason to stop
        ui_parsed_completion.choices[0].finish_reason = "stop"
        yield ui_parsed_completion
