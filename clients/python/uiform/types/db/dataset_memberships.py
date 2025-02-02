from typing import Literal, Dict, Any
from pydantic import BaseModel, Field
import datetime
import uuid

from ..image_settings import ImageSettings
from ..modalities import Modality

AnnotationStatus = Literal["empty", "incomplete", "completed"]

class Annotation(BaseModel):
    status: AnnotationStatus = Field(default="empty", description="Status of the annotation (empty, incomplete, completed)")
    data: Dict[str, Any] = Field(default_factory=dict, description="Data of the annotation")
    updated_at: datetime.datetime = Field(default_factory=datetime.datetime.now, description="Timestamp for when the annotation was last updated")

class DatasetMembership(BaseModel):
    object: Literal["dataset_membership"] = "dataset_membership"
    id: str = Field(default_factory=lambda: f"dm_{uuid.uuid4()}", description="Unique identifier for the dataset membership")
    file_id: str = Field(description="ID of the file that belongs to the dataset")
    dataset_id: str = Field(description="ID of the dataset that contains the file") 
    created_at: datetime.datetime = Field(default_factory=datetime.datetime.now, description="Timestamp for when the membership was created")
    annotation: Annotation = Field(default=Annotation(), description="Annotation of the file")


class BaseGenerateAnnotationRequest(BaseModel):
    dataset_id: str
    model: str
    modality: Modality = "native"
    image_settings: ImageSettings = Field(default=ImageSettings(), description="Preprocessing operations applied to image before sending them to the llm")
    temperature: float = 0.0
    upsert: bool = Field(default=False, description="If True, the annotation will be upserted if it already exists")


class GenerateAnnotationRequest(BaseGenerateAnnotationRequest):
    file_id: str

class GenerateAnnotationsRequest(BaseGenerateAnnotationRequest):
    file_ids: list[str]
