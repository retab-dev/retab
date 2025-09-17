from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel, Field, ConfigDict
from .modality import Modality
from .browser_canvas import BrowserCanvas

class InferenceSettings(BaseModel):
    model: str = "gpt-5-mini"
    temperature: float = 0.0
    reasoning_effort: ChatCompletionReasoningEffort = "minimal"
    image_resolution_dpi: int = Field(default=192, description="Resolution of the image sent to the LLM", ge=96, le=300)
    browser_canvas: BrowserCanvas = "A4"
    n_consensus: int = Field(default=1, description="Number of consensus rounds to perform")
    modality: Modality = "native"

    model_config = ConfigDict(extra="ignore")

