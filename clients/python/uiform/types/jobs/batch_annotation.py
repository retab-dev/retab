from typing import Literal, Any, Optional

from pydantic import BaseModel, Field

from ..modalities import Modality
from ..image_settings import ImageSettings

CheckPoint = Literal["input_parsed", "file_uploaded", "batch_job_created", "waiting_batch", "results_collected", "completed", "error", None]

class AnnotationProps(BaseModel):
    model: str = "gpt-4o-mini"
    temperature: float = 0.0
    modality: Modality
    image_settings : ImageSettings = Field(default_factory=ImageSettings, description="Preprocessing operations applied to image before sending them to the llm")

class AnnotationInputData(BaseModel):
    dataset_id: str
    files_ids: Optional[list[str]]=None
    upsert: bool = False
    annotation_props: AnnotationProps

class AnnotationJob(BaseModel):
    job_type: Literal["annotate-files"] = "annotate-files"
    input_data: AnnotationInputData
    checkpoint: CheckPoint = None
    checkpoint_data: Optional[dict] = None
