from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from openai.types.shared_params.response_format_json_schema import ResponseFormatJSONSchema

# from openai.lib._parsing import ResponseFormatT
from pydantic import BaseModel as ResponseFormatT

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...utils.ai_models import assert_valid_model_extraction
from ...types.chat import ChatCompletionRetabMessage
from ...types.completions import RetabChatCompletionsRequest
from ...types.documents.extractions import RetabParsedChatCompletion
from ...types.schemas.object import Schema
from ...types.standards import PreparedRequest


class BaseCompletionsMixin:
    def prepare_parse(
        self,
        response_format: type[ResponseFormatT],
        messages: list[ChatCompletionRetabMessage],
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

        request = RetabChatCompletionsRequest(
            model=model,
            messages=messages,
            response_format={
                "type": "json_schema",
                "json_schema": {
                    "name": schema_obj.id,
                    "schema": schema_obj.inference_json_schema,
                    "strict": True,
                },
            },
            temperature=temperature,
            stream=stream,
            reasoning_effort=reasoning_effort,
            n_consensus=n_consensus,
        )
        return PreparedRequest(method="POST", url="/v1/completions", data=request.model_dump(), idempotency_key=idempotency_key)

    def prepare_create(
        self,
        response_format: ResponseFormatJSONSchema,
        messages: list[ChatCompletionRetabMessage],
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

        request = RetabChatCompletionsRequest(
            model=model,
            messages=messages,
            response_format={
                "type": "json_schema",
                "json_schema": {
                    "name": schema_obj.id,
                    "schema": schema_obj.inference_json_schema,
                    "strict": True,
                },
            },
            temperature=temperature,
            stream=stream,
            reasoning_effort=reasoning_effort,
            n_consensus=n_consensus,
        )
        return PreparedRequest(method="POST", url="/v1/completions", data=request.model_dump(), idempotency_key=idempotency_key)


class Completions(SyncAPIResource, BaseCompletionsMixin):
    """Multi-provider Completions API wrapper"""

    def create(
        self,
        response_format: ResponseFormatJSONSchema,
        messages: list[ChatCompletionRetabMessage],
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        n_consensus: int = 1,
        idempotency_key: str | None = None,
        stream: bool = False,
    ) -> RetabParsedChatCompletion:
        """
        Create a completion using the Retab API.
        """

        request = self.prepare_create(
            model=model,
            temperature=temperature,
            reasoning_effort=reasoning_effort,
            stream=stream,
            messages=messages,
            response_format=response_format,
            n_consensus=n_consensus,
            idempotency_key=idempotency_key,
        )

        response = self._client._prepared_request(request)

        return RetabParsedChatCompletion.model_validate(response)

    def parse(
        self,
        response_format: type[ResponseFormatT],
        messages: list[ChatCompletionRetabMessage],
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        n_consensus: int = 1,
        idempotency_key: str | None = None,
    ) -> RetabParsedChatCompletion:
        """
        Parse messages using the Retab API to extract structured data according to the provided JSON schema.

        Args:
            response_format: JSON schema defining the expected data structure
            messages: List of chat messages to parse
            model: The AI model to use for processing
            temperature: Model temperature setting (0-1)
            reasoning_effort: The effort level for the model to reason about the input data
            idempotency_key: Idempotency key for request
            store: Whether to store the data in the Retab database

        Returns:
            RetabParsedChatCompletion: Parsed response from the API
        """
        request = self.prepare_parse(
            response_format=response_format,
            messages=messages,
            model=model,
            temperature=temperature,
            reasoning_effort=reasoning_effort,
            stream=False,
            n_consensus=n_consensus,
            idempotency_key=idempotency_key,
        )
        response = self._client._prepared_request(request)

        return RetabParsedChatCompletion.model_validate(response)


class AsyncCompletions(AsyncAPIResource, BaseCompletionsMixin):
    """Multi-provider Completions API wrapper for asynchronous usage."""

    async def create(
        self,
        response_format: ResponseFormatJSONSchema,
        messages: list[ChatCompletionRetabMessage],
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        n_consensus: int = 1,
        idempotency_key: str | None = None,
        stream: bool = False,
    ) -> RetabParsedChatCompletion:
        """
        Create a completion using the Retab API.
        """

        request = self.prepare_create(
            model=model,
            temperature=temperature,
            reasoning_effort=reasoning_effort,
            stream=stream,
            messages=messages,
            response_format=response_format,
            n_consensus=n_consensus,
            idempotency_key=idempotency_key,
        )

        response = await self._client._prepared_request(request)
        return RetabParsedChatCompletion.model_validate(response)

    async def parse(
        self,
        response_format: type[ResponseFormatT],
        messages: list[ChatCompletionRetabMessage],
        model: str = "gpt-4o-2024-08-06",
        temperature: float = 0,
        reasoning_effort: ChatCompletionReasoningEffort = "medium",
        n_consensus: int = 1,
        idempotency_key: str | None = None,
    ) -> RetabParsedChatCompletion:
        """
        Parse messages using the Retab API asynchronously.

        Args:
            json_schema: JSON schema defining the expected data structure
            messages: List of chat messages to parse
            model: The AI model to use
            temperature: Model temperature setting (0-1)
            reasoning_effort: The effort level for the model to reason about the input data
            n_consensus: Number of consensus models to use for extraction
            idempotency_key: Idempotency key for request

        Returns:
            RetabParsedChatCompletion: Parsed response from the API
        """
        request = self.prepare_parse(
            response_format=response_format,
            messages=messages,
            model=model,
            temperature=temperature,
            reasoning_effort=reasoning_effort,
            stream=False,
            n_consensus=n_consensus,
            idempotency_key=idempotency_key,
        )
        response = await self._client._prepared_request(request)
        return RetabParsedChatCompletion.model_validate(response)
