from pydantic import BaseModel, Field

from ..modalities import Modality
from ..mime import MIMEData
from ..image_settings import ImageSettings

class GenerateSchemaRequest(BaseModel):
    """
    The request body for generating a JSON Schema from scratch.
    """
    documents: list[MIMEData]
    model: str = "gpt-4o-mini"
    temperature: float = 0.0

    modality: Modality
    """The modality of the document to load."""

    image_settings : ImageSettings = Field(default_factory=ImageSettings, description="Preprocessing operations applied to image before sending them to the llm")
    """The image operations to apply to the document."""
