from pydantic import BaseModel
from ..modalities import Modality
from ..mime import MIMEData
from ..documents.text_operations import TextOperations

# Schemas API
class GenerateSchemaBase(BaseModel):
    text_operations: TextOperations | None = None
    model: str = "gpt-4o-2024-11-20"
    temperature: float = 0.0
    modality: Modality = "native"

class GenerateSchemaRequest(GenerateSchemaBase):
    """
    The request body for generating a JSON Schema from scratch.
    """
    documents: list[MIMEData]
