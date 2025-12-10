from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel, ConfigDict, Field

from ..mime import MIMEData


class GenerateSchemaRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")
    documents: list[MIMEData]
    model: str = "gpt-5-mini"
    temperature: float = 0.0
    reasoning_effort: ChatCompletionReasoningEffort = "minimal"
    instructions: str | None = None
    """The modality of the document to load."""
    image_resolution_dpi: int = Field(default=192, description="Resolution of the image sent to the LLM", ge=96, le=300)

    """The image operations to apply to the document."""
    stream: bool = False
    """Whether to stream the response."""
