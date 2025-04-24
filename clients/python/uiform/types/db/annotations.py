import datetime
from typing import Any, Dict, Literal

import nanoid  # type: ignore
from pydantic import BaseModel, Field

from ..image_settings import ImageSettings
from ..modalities import Modality


class AnnotationParameters(BaseModel):
    model: str
    modality: Modality | None = "native"
    image_settings: ImageSettings | None = Field(default=ImageSettings(), description="Preprocessing operations applied to image before sending them to the llm")
    temperature: float = 0.0


class Annotation(BaseModel):
    file_id: str = Field(description="ID of the file that the annotation belongs to")
    parameters: AnnotationParameters = Field(description="Parameters used for the annotation")
    data: Dict[str, Any] = Field(default_factory=dict, description="Data of the annotation")
    schema_id: str = Field(description="ID of the schema used for the annotation")
    organization_id: str = Field(description="ID of the organization that owns the annotation")
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc), description="Timestamp for when the annotation was last updated")
