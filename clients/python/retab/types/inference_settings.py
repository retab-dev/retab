from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel, Field, ConfigDict
from .modality import Modality
from .browser_canvas import BrowserCanvas
from typing import Any

class InferenceSettings(BaseModel):
    model: str = "gpt-5-mini"
    temperature: float = 0.0
    reasoning_effort: ChatCompletionReasoningEffort = "minimal"
    image_resolution_dpi: int = Field(default=192, description="Resolution of the image sent to the LLM", ge=96, le=300)
    browser_canvas: BrowserCanvas = "A4"
    n_consensus: int = Field(default=1, description="Number of consensus rounds to perform")
    modality: Modality = "native"
    parallel_ocr_keys: dict[str, str] | None = Field(default=None, description="If set, keys to be used for the extraction of long lists of data using Parallel OCR", examples=[{"properties": "ID", "products": "identity.id"}])

    model_config = ConfigDict(extra="ignore")


class ExtractionSettings(InferenceSettings):
    json_schema: dict[str, Any] = Field(..., description="JSON schema format used to validate the output data.")
    