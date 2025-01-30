from pydantic import BaseModel
from ..mime import MIMEData
from ..documents.create_messages import DocumentProcessingConfig

class GenerateSchemaRequest(DocumentProcessingConfig):
    """
    The request body for generating a JSON Schema from scratch.
    """
    documents: list[MIMEData]
    model: str = "gpt-4o-mini"
    temperature: float = 0.0
