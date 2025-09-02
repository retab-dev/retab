from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel, ConfigDict, Field

from ..mime import MIMEData
from ..browser_canvas import BrowserCanvas


class GenerateSchemaRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")
    documents: list[MIMEData]
    model: str = "gpt-4o-mini"
    temperature: float = 0.0
    reasoning_effort: ChatCompletionReasoningEffort = "minimal"
    instructions: str | None = None
    """The modality of the document to load."""
    image_resolution_dpi: int = Field(default=96, description="Resolution of the image sent to the LLM")
    browser_canvas: BrowserCanvas = Field(
        default="A4", description="Sets the size of the browser canvas for rendering documents in browser-based processing. Choose a size that matches the document type."
    )

    """The image operations to apply to the document."""
    stream: bool = False
    """Whether to stream the response."""
