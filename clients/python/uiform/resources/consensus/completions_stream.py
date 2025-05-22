import json
from pathlib import Path
from typing import Any, AsyncGenerator, Generator

from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from openai.types.chat.parsed_chat_completion import ParsedChatCompletionMessage
from openai.types.shared_params.response_format_json_schema import ResponseFormatJSONSchema
#from openai.lib._parsing import ResponseFormatT
from pydantic import BaseModel as ResponseFormatT


from ..._resource import AsyncAPIResource, SyncAPIResource
from ..._utils.ai_models import assert_valid_model_extraction
from ..._utils.json_schema import load_json_schema, unflatten_dict
from ..._utils.stream_context_managers import as_async_context_manager, as_context_manager
from ...types.chat import ChatCompletionUiformMessage
from ...types.completions import UiChatCompletionsRequest
from ...types.documents.extractions import UiParsedChatCompletion, UiParsedChatCompletionChunk, UiParsedChoice
from ...types.standards import PreparedRequest
from ...types.schemas.object import Schema



class BaseCompletionsMixin:

    def prepare_parse(
        self,
        response_format: type[ResponseFormatT],
        messages: list[ChatCompletionUiformMessage],
        model: str,
        temperature: float,
        reasoning_effort: ChatCompletionReasoningEffort,
        stream: bool,
        n_consensus: int,
        idempotency_key: str | None = None,
    ) -> PreparedRequest:
        assert_valid_model_extraction(model)

        json_schema = response_format.model_json_schema()

        schema_obj = Schema(json_schema=json_schema)

        data = {
            "messages": messages,
            "response_format": {
                "type": "json_schema",
                "json_schema": {
                    "name": schema_obj.id,
                    "schema": schema_obj.inference_json_schema,
                    "strict": True,
                },
            },
            "model": model,
            "temperature": temperature,
            "stream": stream,
            "reasoning_effort": reasoning_effort,
            "n_consensus": n_consensus,
        }

        # Validate DocumentAPIRequest data (raises exception if invalid)
        ui_chat_completions_request = UiChatCompletionsRequest.model_validate(data)

        return PreparedRequest(method="POST", url="/v1/completions", data=ui_chat_completions_request.model_dump(), idempotency_key=idempotency_key)


    def prepare_create(
        self,
        response_format: ResponseFormatJSONSchema,
        messages: list[ChatCompletionUiformMessage],
        model: str,
        temperature: float,
        reasoning_effort: ChatCompletionReasoningEffort,
        stream: bool,
        n_consensus: int,
        idempotency_key: str | None = None,
    ) -> PreparedRequest:
        
        json_schema = response_format["json_schema"].get("schema")

        assert isinstance(json_schema, dict), f"json_schema must be a dictionary, got {type(json_schema)}"

        schema_obj = Schema(json_schema=json_schema)

        data = {
            "messages": messages,
            "response_format": {
                "type": "json_schema",
                "json_schema": {
                    "name": schema_obj.id,
                    "schema": schema_obj.inference_json_schema,
                    "strict": True,
                },
            },
            "model": model,
            "temperature": temperature,
            "stream": stream,
            "reasoning_effort": reasoning_effort,
            "n_consensus": n_consensus,
        }

        # Validate DocumentAPIRequest data (raises exception if invalid)
        ui_chat_completions_request = UiChatCompletionsRequest.model_validate(data)

        return PreparedRequest(method="POST", url="/v1/completions", data=ui_chat_completions_request.model_dump(), idempotency_key=idempotency_key)


class Completions(SyncAPIResource, BaseCompletionsMixin):
    """Multi-provider Completions API wrapper"""

    @as_context_manager
    def stream(
        self,
        response_format: type[ResponseFormatT],
        messages: list[ChatCompletionUiformMessage],
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        n_consensus: int = 1,
        idempotency_key: str | None = None,
    ) -> Generator[UiParsedChatCompletion, None, None]:
        """
        Process messages using the UiForm API with streaming enabled.

        Args:
            response_format: JSON schema defining the expected data structure
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
        request = self.prepare_parse(
            response_format=response_format,
            messages=messages,
            model=model,
            temperature=temperature,
            reasoning_effort=reasoning_effort,
            stream=True,
            n_consensus=n_consensus,
            idempotency_key=idempotency_key,
        )

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
            ui_parsed_completion.likelihoods = unflatten_dict(ui_parsed_chat_completion_cum_chunk.choices[0].delta.flat_likelihoods)
            parsed = unflatten_dict(ui_parsed_chat_completion_cum_chunk.choices[0].delta.flat_parsed)
            ui_parsed_completion.choices[0].message.content = json.dumps(parsed)
            ui_parsed_completion.choices[0].message.parsed = parsed

            yield ui_parsed_completion

        # change the finish_reason to stop
        ui_parsed_completion.choices[0].finish_reason = "stop"
        yield ui_parsed_completion


class AsyncCompletions(AsyncAPIResource, BaseCompletionsMixin):
    """Multi-provider Completions API wrapper for asynchronous usage."""

   
    @as_async_context_manager
    async def stream(
        self,
        response_format: type[ResponseFormatT],
        messages: list[ChatCompletionUiformMessage],
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        n_consensus: int = 1,
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
            n_consensus: Number of consensus models to use for extraction
            idempotency_key: Idempotency key for request

        Returns:
            AsyncGenerator[UiParsedChatCompletion]: Stream of parsed responses

        Usage:
        ```python
        async with uiform.completions.stream(json_schema, messages, model, temperature, reasoning_effort, n_consensus) as stream:
            async for response in stream:
                print(response)
        ```
        """
        request = self.prepare_parse(
            response_format=response_format,
            messages=messages,
            model=model,
            temperature=temperature,
            reasoning_effort=reasoning_effort,
            stream=True,
            n_consensus=n_consensus,
            idempotency_key=idempotency_key,
        )

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
        async for chunk_json in self._client._prepared_request_stream(request):
            if not chunk_json:
                continue
            ui_parsed_chat_completion_cum_chunk = UiParsedChatCompletionChunk.model_validate(chunk_json).chunk_accumulator(ui_parsed_chat_completion_cum_chunk)
            # Basic stuff
            ui_parsed_completion.id = ui_parsed_chat_completion_cum_chunk.id
            ui_parsed_completion.created = ui_parsed_chat_completion_cum_chunk.created
            ui_parsed_completion.model = ui_parsed_chat_completion_cum_chunk.model

            # Update the ui_parsed_completion object
            ui_parsed_completion.likelihoods = unflatten_dict(ui_parsed_chat_completion_cum_chunk.choices[0].delta.flat_likelihoods)
            parsed = unflatten_dict(ui_parsed_chat_completion_cum_chunk.choices[0].delta.flat_parsed)
            ui_parsed_completion.choices[0].message.content = json.dumps(parsed)
            ui_parsed_completion.choices[0].message.parsed = parsed

            yield ui_parsed_completion

        # change the finish_reason to stop
        ui_parsed_completion.choices[0].finish_reason = "stop"
        yield ui_parsed_completion
