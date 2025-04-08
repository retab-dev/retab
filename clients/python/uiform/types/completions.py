from typing import Any

from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel, ConfigDict, Field

from .._utils.ai_models import find_provider_from_model
from .ai_models import AIProvider
from .chat import ChatCompletionUiformMessage


class UiChatCompletionsParseRequest(BaseModel):
    model_config = ConfigDict(arbitrary_types_allowed=True)
    model: str = Field(..., description="Model used for chat completion")
    messages: list[ChatCompletionUiformMessage] = Field(..., description="Messages to be parsed")
    json_schema: dict[str, Any] = Field(..., description="JSON schema format used to validate the output data.")
    temperature: float = Field(default=0.0, description="Temperature for sampling. If not provided, the default temperature for the model will be used.", examples=[0.0])
    reasoning_effort: ChatCompletionReasoningEffort = Field(
        default="medium", description="The effort level for the model to reason about the input data. If not provided, the default reasoning effort for the model will be used."
    )
    # Regular fields
    stream: bool = Field(default=False, description="If true, the extraction will be streamed to the user using the active WebSocket connection")
    seed: int | None = Field(default=None, description="Seed for the random number generator. If not provided, a random seed will be generated.", examples=[None])

    @property
    def provider(self) -> AIProvider:
        """
        Determines the AI provider based on the model specified.

        Returns:
            AIProvider: The AI provider corresponding to the given model.
        """
        return find_provider_from_model(self.model)
