from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel, Field, ConfigDict

from .browser_canvas import BrowserCanvas

class InferenceSettings(BaseModel):
    model_config = ConfigDict(extra="ignore")
    model: str = "gpt-5-mini"
    temperature: float = 0.0
    reasoning_effort: ChatCompletionReasoningEffort = "minimal"
    image_resolution_dpi: int = 96
    browser_canvas: BrowserCanvas = "A4"
    n_consensus: int = Field(default=1, description="Number of consensus rounds to perform")
