from typing import Literal
from pydantic import BaseModel, Field
import datetime
import uuid


from typing import Literal, Dict, Any, List
from pydantic import BaseModel, Field
import datetime
import uuid

from ...types.documents.image_operations import ImageOperations
from ...types.documents.text_operations import TextOperations
from ...types.documents.create_messages import ChatCompletionUiformMessage
from ...types.modalities import Modality


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

class GenerateAnnotationRequest(BaseModel):
    dataset_id: str
    file_id: str
    model: str
    modality: Modality = "native"
    text_operations: TextOperations = Field(default=TextOperations(), description="Additional context to be used by the AI model", examples=[{
        "regex_instructions": [{
            "name": "VAT Number",
            "description": "All potential VAT numbers in the documents",
            "pattern": r"[Ff][Rr]\s*(\d\s*){11}"
        }]
    }])
    image_operations: ImageOperations = Field(
        default=ImageOperations(**{
            "correct_image_orientation": True,
            "dpi" : 72,
            "image_to_text": "ocr", 
            "browser_canvas": "A4"
        }),
        description="Preprocessing operations applied to image before sending them to the llm",
        examples=[{
            "correct_image_orientation": True,
            "dpi" : 72,
            "image_to_text": "ocr", 
            "browser_canvas": "A4"
        }]
    )
    temperature: float = 0.0
    messages: List[ChatCompletionUiformMessage] = []
    upsert: bool = Field(default=False, description="If True, the annotation will be upserted if it already exists")
