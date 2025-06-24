import datetime
from typing import Any, Dict

from pydantic import BaseModel, Field

from ..modalities import Modality
from ..browser_canvas import BrowserCanvas


class AnnotationParameters(BaseModel):
    model: str
    modality: Modality | None = "native"
    image_resolution_dpi: int = 96
    browser_canvas: BrowserCanvas = "A4"
    temperature: float = 0.0


class Annotation(BaseModel):
    file_id: str = Field(description="ID of the file that the annotation belongs to")
    parameters: AnnotationParameters = Field(description="Parameters used for the annotation")
    data: Dict[str, Any] = Field(default_factory=dict, description="Data of the annotation")
    schema_id: str = Field(description="ID of the schema used for the annotation")
    organization_id: str = Field(description="ID of the organization that owns the annotation")
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc), description="Timestamp for when the annotation was last updated")
