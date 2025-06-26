from typing import Any, Optional, Union

from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from openai.types.responses.response_input_param import ResponseInputParam
from openai.types.responses.response_text_config_param import ResponseTextConfigParam
from openai.types.shared_params.reasoning import Reasoning
from openai.types.shared_params.response_format_json_schema import ResponseFormatJSONSchema
from pydantic import BaseModel, ConfigDict, Field

from ..utils.ai_models import get_provider_for_model
from .ai_models import AIProvider
from .chat import ChatCompletionRetabMessage


class RetabChatCompletionsRequest(BaseModel):
    model_config = ConfigDict(arbitrary_types_allowed=True)
    model: str = Field(..., description="Model used for chat completion")
    messages: list[ChatCompletionRetabMessage] = Field(..., description="Messages to be parsed")
    response_format: ResponseFormatJSONSchema = Field(..., description="response format used to validate the output data.")
    temperature: float = Field(default=0.0, description="Temperature for sampling. If not provided, the default temperature for the model will be used.", examples=[0.0])
    reasoning_effort: ChatCompletionReasoningEffort = Field(
        default="medium", description="The effort level for the model to reason about the input data. If not provided, the default reasoning effort for the model will be used."
    )
    # Regular fields
    stream: bool = Field(default=False, description="If true, the extraction will be streamed to the user using the active WebSocket connection")
    seed: int | None = Field(default=None, description="Seed for the random number generator. If not provided, a random seed will be generated.", examples=[None])
    n_consensus: int = Field(default=1, description="Number of consensus models to use for extraction. If greater than 1 the temperature cannot be 0.")

    @property
    def provider(self) -> AIProvider:
        """
        Determines the AI provider based on the model specified.

        Returns:
            AIProvider: The AI provider corresponding to the given model.
        """
        return get_provider_for_model(self.model)


class RetabChatCompletionsParseRequest(BaseModel):
    model_config = ConfigDict(arbitrary_types_allowed=True)
    model: str = Field(..., description="Model used for chat completion")
    messages: list[ChatCompletionRetabMessage] = Field(..., description="Messages to be parsed")
    json_schema: dict[str, Any] = Field(..., description="JSON schema format used to validate the output data.")
    temperature: float = Field(default=0.0, description="Temperature for sampling. If not provided, the default temperature for the model will be used.", examples=[0.0])
    reasoning_effort: ChatCompletionReasoningEffort = Field(
        default="medium", description="The effort level for the model to reason about the input data. If not provided, the default reasoning effort for the model will be used."
    )
    # Regular fields
    stream: bool = Field(default=False, description="If true, the extraction will be streamed to the user using the active WebSocket connection")
    seed: int | None = Field(default=None, description="Seed for the random number generator. If not provided, a random seed will be generated.", examples=[None])
    n_consensus: int = Field(default=1, description="Number of consensus models to use for extraction. If greater than 1 the temperature cannot be 0.")

    @property
    def provider(self) -> AIProvider:
        """
        Determines the AI provider based on the model specified.

        Returns:
            AIProvider: The AI provider corresponding to the given model.
        """
        return get_provider_for_model(self.model)


class RetabChatResponseCreateRequest(BaseModel):
    input: Union[str, ResponseInputParam] = Field(..., description="Input to be parsed")
    instructions: Optional[str] = None

    model_config = ConfigDict(arbitrary_types_allowed=True)
    model: str = Field(..., description="Model used for chat completion")
    temperature: Optional[float] = Field(default=0.0, description="Temperature for sampling. If not provided, the default temperature for the model will be used.", examples=[0.0])
    reasoning: Optional[Reasoning] = Field(
        default=None, description="The effort level for the model to reason about the input data. If not provided, the default reasoning effort for the model will be used."
    )

    stream: Optional[bool] = Field(default=False, description="If true, the extraction will be streamed to the user using the active WebSocket connection")
    seed: int | None = Field(default=None, description="Seed for the random number generator. If not provided, a random seed will be generated.", examples=[None])
    text: ResponseTextConfigParam = Field(default={"format": {"type": "text"}}, description="Format of the response")

    n_consensus: int = Field(default=1, description="Number of consensus models to use for extraction. If greater than 1 the temperature cannot be 0.")

    @property
    def provider(self) -> AIProvider:
        """
        Determines the AI provider based on the model specified.

        Returns:
            AIProvider: The AI provider corresponding to the given model.
        """
        return get_provider_for_model(self.model)
