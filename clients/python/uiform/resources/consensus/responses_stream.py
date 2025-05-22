import json
from pathlib import Path
from typing import Any, AsyncGenerator, Generator, TypeVar, Generic, Optional, Union, List, Sequence, cast

from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from openai.types.chat.parsed_chat_completion import ParsedChatCompletionMessage
from openai.types.responses.response import Response
from openai.types.responses.response_input_param import ResponseInputParam, ResponseInputItemParam
from openai.types.responses.response_output_item import ResponseOutputItem
from openai.types.shared_params.response_format_json_schema import ResponseFormatJSONSchema
from pydantic import BaseModel

from ..._resource import AsyncAPIResource, SyncAPIResource
from ..._utils.ai_models import assert_valid_model_extraction
from ..._utils.json_schema import load_json_schema, unflatten_dict
from ..._utils.responses import convert_to_openai_format, convert_from_openai_format, parse_openai_responses_response
from ..._utils.stream_context_managers import as_async_context_manager, as_context_manager
from ...types.chat import ChatCompletionUiformMessage
from ...types.completions import UiChatResponseCreateRequest, UiChatCompletionsRequest
from ...types.documents.extractions import UiParsedChatCompletion, UiParsedChatCompletionChunk, UiParsedChoice, UiResponse
from ...types.standards import PreparedRequest
from ...types.schemas.object import Schema

from typing import Optional, Union
from openai.types.shared_params.reasoning import Reasoning
from openai.types.responses.response_input_param import ResponseInputParam
from openai.types.responses.response_text_config_param import ResponseTextConfigParam
from openai.types.shared_params.response_format_json_schema import ResponseFormatJSONSchema

T = TypeVar('T', bound=BaseModel)

class BaseResponsesMixin:
    def prepare_create(
        self,
        model: str,
        input: Union[str, ResponseInputParam],
        text: ResponseTextConfigParam,
        temperature: float = 0,
        reasoning: Optional[Reasoning] = None,
        stream: bool = False,
        n_consensus: int = 1,
        instructions: Optional[str] = None,
        idempotency_key: Optional[str] = None,
    ) -> PreparedRequest:
        """
        Prepare a request for the Responses API create method.
        """
        assert_valid_model_extraction(model)

        text_format = text.get("format", None)
        assert text_format is not None, "text.format is required"
        json_schema = text_format.get("schema", None)
        assert json_schema is not None, "text.format.schema is required"

        schema_obj = Schema(json_schema=json_schema)

        if instructions is None:
            instructions = schema_obj.developer_system_prompt
        
        # Create the request object based on the UiChatResponseCreateRequest model
        data = UiChatResponseCreateRequest(
            model=model,
            input=input,
            temperature=temperature,
            stream=stream,
            reasoning=reasoning, 
            n_consensus=n_consensus,
            text={
                "format": {
                    "type": "json_schema", 
                    "name": schema_obj.id, 
                    "schema": schema_obj.inference_json_schema, 
                    "strict": True
                }
            },
            instructions=instructions,
        )

        # Validate the request data
        ui_chat_response_create_request = UiChatResponseCreateRequest.model_validate(data)

        return PreparedRequest(
            method="POST", 
            url="/v1/responses", 
            data=ui_chat_response_create_request.model_dump(), 
            idempotency_key=idempotency_key
        )

    def prepare_parse(
        self,
        model: str,
        input: Union[str, ResponseInputParam],
        text_format: type[BaseModel],
        temperature: float = 0,
        reasoning: Optional[Reasoning] = None,
        stream: bool = False,
        n_consensus: int = 1,
        instructions: Optional[str] = None,
        idempotency_key: Optional[str] = None,
    ) -> PreparedRequest:
        """
        Prepare a request for the Responses API parse method.
        """

        assert_valid_model_extraction(model)

        schema_obj = Schema(pydantic_model=text_format)

        if instructions is None:
            instructions = schema_obj.developer_system_prompt
        
        # Create the request object based on the UiChatResponseCreateRequest model
        data = UiChatResponseCreateRequest(
            model=model,
            input=input,
            temperature=temperature,
            stream=stream,
            reasoning=reasoning, 
            n_consensus=n_consensus,
            text={
                "format": {
                    "type": "json_schema", 
                    "name": schema_obj.id, 
                    "schema": schema_obj.inference_json_schema, 
                    "strict": True
                }
            },
            instructions=instructions,
        )

        # Validate the request data
        ui_chat_response_create_request = UiChatResponseCreateRequest.model_validate(data)

        return PreparedRequest(
            method="POST", 
            url="/v1/responses", 
            data=ui_chat_response_create_request.model_dump(), 
            idempotency_key=idempotency_key
        )
        

        return PreparedRequest(
            method="POST", 
            url="/v1/completions", 
            data=ui_chat_completions_request.model_dump(), 
            idempotency_key=idempotency_key
        )


class Responses(SyncAPIResource, BaseResponsesMixin):
    """UiForm Responses API compatible with OpenAI Responses API"""

    @as_context_manager
    def stream(
        self,
        model: str,
        input: Union[str, ResponseInputParam],
        text: ResponseTextConfigParam,
        temperature: float = 0,
        reasoning: Optional[Reasoning] = None,
        n_consensus: int = 1,
        instructions: Optional[str] = None,
        idempotency_key: Optional[str] = None,
    ) -> Generator[UiResponse, None, None]:
        """
        Create a completion using the UiForm API with streaming enabled.
        
        Args:
            model: The model to use
            input: The input text or message array
            text: The response format configuration
            temperature: Model temperature setting (0-1)
            reasoning: The effort level for the model to reason about the input data
            n_consensus: Number of consensus models to use
            instructions: Optional system instructions
            idempotency_key: Idempotency key for request
            
        Returns:
            Generator[UiResponse]: Stream of responses

        Usage:
        ```python
        with uiform.responses.stream(model, input, text, temperature, reasoning) as stream:
            for response in stream:
                print(response)
        ```
        """
        request = self.prepare_create(
            model=model,
            input=input,
            temperature=temperature,
            reasoning=reasoning,
            stream=True,
            text=text,
            instructions=instructions,
            n_consensus=n_consensus,
            idempotency_key=idempotency_key,
        )

        # Request the stream and return a context manager
        for chunk_json in self._client._prepared_request_stream(request):
            if not chunk_json:
                continue
            response = UiResponse.model_validate(chunk_json)
            yield response

    @as_context_manager
    def stream_parse(
        self,
        model: str,
        input: Union[str, ResponseInputParam],
        text_format: type[T],
        temperature: float = 0,
        reasoning: Optional[Reasoning] = None,
        n_consensus: int = 1,
        instructions: Optional[str] = None,
        idempotency_key: Optional[str] = None,
    ) -> Generator[UiResponse, None, None]:
        """
        Parse content using the UiForm API with streaming enabled.
        
        Args:
            model: The model to use
            input: The input text or message array
            text_format: The Pydantic model defining the expected output format
            temperature: Model temperature setting (0-1)
            reasoning: The effort level for the model to reason about the input data
            n_consensus: Number of consensus models to use
            instructions: Optional system instructions
            idempotency_key: Idempotency key for request
            
        Returns:
            Generator[UiResponse]: Stream of parsed responses

        Usage:
        ```python
        with uiform.responses.stream_parse(model, input, MyModel, temperature, reasoning) as stream:
            for response in stream:
                print(response)
        ```
        """
        request = self.prepare_parse(
            model=model,
            input=input,
            temperature=temperature,
            reasoning=reasoning,
            stream=True,
            text_format=text_format,
            instructions=instructions,
            n_consensus=n_consensus,
            idempotency_key=idempotency_key,
        )

        # Request the stream and return a context manager
        for chunk_json in self._client._prepared_request_stream(request):
            if not chunk_json:
                continue
            response = UiResponse.model_validate(chunk_json)
            yield response



class AsyncResponses(AsyncAPIResource, BaseResponsesMixin):
    """UiForm Responses API compatible with OpenAI Responses API for async usage"""

    @as_async_context_manager
    async def stream(
        self,
        model: str,
        input: Union[str, ResponseInputParam],
        text: ResponseTextConfigParam,
        temperature: float = 0,
        reasoning: Optional[Reasoning] = None,
        n_consensus: int = 1,
        instructions: Optional[str] = None,
        idempotency_key: Optional[str] = None,
    ) -> AsyncGenerator[UiResponse, None]:
        """
        Create a completion using the UiForm API asynchronously with streaming enabled.
        
        Args:
            model: The model to use
            input: The input text or message array
            text: The response format configuration
            temperature: Model temperature setting (0-1)
            reasoning: The effort level for the model to reason about the input data
            n_consensus: Number of consensus models to use
            instructions: Optional system instructions
            idempotency_key: Idempotency key for request
            
        Returns:
            AsyncGenerator[UiResponse]: Async stream of responses

        Usage:
        ```python
        async with uiform.responses.async_stream(model, input, text, temperature, reasoning) as stream:
            async for response in stream:
                print(response)
        ```
        """
        request = self.prepare_create(
            model=model,
            input=input,
            temperature=temperature,
            reasoning=reasoning,
            stream=True,
            text=text,
            instructions=instructions,
            n_consensus=n_consensus,
            idempotency_key=idempotency_key,
        )

        # Request the stream and return a context manager
        async for chunk_json in self._client._prepared_request_stream(request):
            if not chunk_json:
                continue
            response = UiResponse.model_validate(chunk_json)
            yield response
        
    @as_async_context_manager
    async def stream_parse(
        self,
        model: str,
        input: Union[str, ResponseInputParam],
        text_format: type[T],
        temperature: float = 0,
        reasoning: Optional[Reasoning] = None,
        n_consensus: int = 1,
        instructions: Optional[str] = None,
        idempotency_key: Optional[str] = None,
    ) -> AsyncGenerator[UiResponse, None]:
        """
        Parse content using the UiForm API asynchronously with streaming enabled.
        
        Args:
            model: The model to use
            input: The input text or message array
            text_format: The Pydantic model defining the expected output format
            temperature: Model temperature setting (0-1)
            reasoning: The effort level for the model to reason about the input data
            n_consensus: Number of consensus models to use
            instructions: Optional system instructions
            idempotency_key: Idempotency key for request
            
        Returns:
            AsyncGenerator[UiResponse]: Async stream of parsed responses

        Usage:
        ```python
        async with uiform.responses.async_stream_parse(model, input, MyModel, temperature, reasoning) as stream:
            async for response in stream:
                print(response)
        ```
        """
        request = self.prepare_parse(
            model=model,
            input=input,
            temperature=temperature,
            reasoning=reasoning,
            stream=True,
            text_format=text_format,
            instructions=instructions,
            n_consensus=n_consensus,
            idempotency_key=idempotency_key,
        )

        # Request the stream and return a context manager
        async for chunk_json in self._client._prepared_request_stream(request):
            if not chunk_json:
                continue
            response = UiResponse.model_validate(chunk_json)
            yield response
        