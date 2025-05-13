from typing import Optional

from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel, Field

from ..image_settings import ImageSettings
from ..modalities import Modality


class InferenceSettings(BaseModel):
    model: str = "gpt-4o-mini"
    temperature: float = 0.0
    modality: Modality
    image_settings: ImageSettings = Field(default_factory=ImageSettings, description="Preprocessing operations applied to image before sending them to the llm")
    reasoning_effort: ChatCompletionReasoningEffort = "medium"


class AnnotationInputData(BaseModel):
    dataset_id: str
    files_ids: Optional[list[str]] = None
    upsert: bool = False
    inference_settings: InferenceSettings
