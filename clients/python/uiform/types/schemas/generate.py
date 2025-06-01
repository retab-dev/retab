from typing import Any, Literal
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel, Field

from ..mime import MIMEData
from ..modalities import Modality


class GenerateSchemaRequest(BaseModel):
    documents: list[MIMEData]
    model: str = "gpt-4o-mini"
    temperature: float = 0.0
    reasoning_effort: ChatCompletionReasoningEffort = "medium"
    modality: Modality
    instructions: str | None = None
    """The modality of the document to load."""
    image_resolution_dpi: int = Field(default=96, description="Resolution of the image sent to the LLM")
    browser_canvas: Literal['A3', 'A4', 'A5'] = Field(default='A4', description="Sets the size of the browser canvas for rendering documents in browser-based processing. Choose a size that matches the document type.")

    """The image operations to apply to the document."""

    stream: bool = False
    """Whether to stream the response."""


class GenerateSystemPromptRequest(GenerateSchemaRequest):
    """
    The request body for generating a system prompt for a JSON Schema.
    """

    json_schema: dict[str, Any]
    
