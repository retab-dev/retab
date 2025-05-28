from pydantic import BaseModel, Field
from typing import Any
from ..mime import MIMEData
from ..modalities import Modality
from ..image_settings import ImageSettings
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort


class EvaluateSchemaRequest(BaseModel):
    """
    The request body for evaluating a JSON Schema.
    """

    documents: list[MIMEData]
    ground_truths: list[dict[str, Any]] | None = None
    model: str = "gpt-4o-mini"
    temperature: float = 0.0
    reasoning_effort: ChatCompletionReasoningEffort = "medium"
    modality: Modality
    image_settings: ImageSettings = Field(default_factory=ImageSettings, description="Preprocessing operations applied to image before sending them to the llm")
    n_consensus: int = 1
    json_schema: dict[str, Any]
