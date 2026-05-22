from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import ConfigDict, Field
from retab.types.base import RetabBaseModel

from ..mime import MIMEData


class GenerateSchemaRequest(RetabBaseModel):
    model_config = ConfigDict(extra="forbid")
    documents: list[MIMEData]
    model: str | None = "retab-small"
    reasoning_effort: ChatCompletionReasoningEffort | None = "minimal"
    instructions: str | None = None
    """The modality of the document to load."""
    image_resolution_dpi: int | None = Field(default=192, description="Resolution of the image sent to the LLM", ge=96, le=300)

    """The image operations to apply to the document."""
    stream: bool | None = False
    """Whether to stream the response."""
