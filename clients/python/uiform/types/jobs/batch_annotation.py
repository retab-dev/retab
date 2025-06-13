from typing import Optional

from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel

from ..modalities import Modality
from ..extractions import BrowserCanvas


class InferenceSettings(BaseModel):
    model: str = "gpt-4o-mini"
    temperature: float = 0.0
    modality: Modality
    image_resolution_dpi: int = 96
    browser_canvas: BrowserCanvas = "A4"
    reasoning_effort: ChatCompletionReasoningEffort = "medium"


class AnnotationInputData(BaseModel):
    dataset_id: str
    files_ids: Optional[list[str]] = None
    upsert: bool = False
    inference_settings: InferenceSettings
