from pydantic import BaseModel
from typing import Any, Literal
from ..modalities import Modality
from ..mime import MIMEData
from ..documents.text_operations import TextOperations

class PromptifyBase(BaseModel):
    text_operations: TextOperations | None = None
    model: str = "gpt-4o-2024-11-20"
    temperature: float = 0.0
    modality: Modality = "native"

class PromptifyRequest(PromptifyBase):
    raw_schema: dict[str, Any]
    documents: list[MIMEData]


