from typing import Any
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel, Field

from ..image_settings import ImageSettings
from ..mime import MIMEData
from ..modalities import Modality


class GenerateBase(BaseModel):
    documents: list[MIMEData]
    model: str = "gpt-4o-mini"
    temperature: float = 0.0
    reasoning_effort: ChatCompletionReasoningEffort = "medium"
    modality: Modality
    """The modality of the document to load."""

    image_settings: ImageSettings = Field(default_factory=ImageSettings, description="Preprocessing operations applied to image before sending them to the llm")
    """The image operations to apply to the document."""

    stream: bool = False
    """Whether to stream the response."""


class GenerateSchemaRequest(GenerateBase):
    """
    The request body for generating a JSON Schema from scratch.
    """

    flat: bool = False
    """Whether to return a flat schema."""


class GenerateSystemPromptRequest(GenerateBase):
    """
    The request body for generating a system prompt for a JSON Schema.
    """

    json_schema: dict[str, Any]
    instructions: str | None = None
