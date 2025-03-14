from pydantic import BaseModel, Field
from typing import Any, Literal
import datetime

from ..modalities import Modality
from ..mime import MIMEData
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort

class PromptifyBase(BaseModel):
    model: str = "gpt-4o-2024-11-20"
    temperature: float = 0.0
    modality: Modality = "native"
    stream: bool = False
    reasoning_effort: ChatCompletionReasoningEffort = "medium"
    
class PromptifyRequest(PromptifyBase):
    raw_schema: dict[str, Any]
    documents: list[MIMEData]

