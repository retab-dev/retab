import datetime
from typing import Any, Literal

from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel, Field

from ..mime import MIMEData
from ..modalities import Modality
from ..._utils.json_schema import generate_schema_data_id, generate_schema_id

class PromptifyBase(BaseModel):
    model: str = "gpt-4o-2024-11-20"
    temperature: float = 0.0
    modality: Modality = "native"
    stream: bool = False
    reasoning_effort: ChatCompletionReasoningEffort = "medium"


class PromptifyRequest(PromptifyBase):
    instructions: str | None = None
    json_schema: dict[str, Any]
    documents: list[MIMEData]


