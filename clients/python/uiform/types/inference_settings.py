from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel, Field

from .browser_canvas import BrowserCanvas
from .modalities import Modality


class InferenceSettings(BaseModel):
    model: str = "gpt-4.1-mini"
    temperature: float = 0.0
    modality: Modality = "native"
    reasoning_effort: ChatCompletionReasoningEffort = "medium"
    image_resolution_dpi: int = 96
    browser_canvas: BrowserCanvas = "A4"
    n_consensus: int = Field(default=1, description="Number of consensus rounds to perform")
