from pydantic import BaseModel, Field
from typing import Optional, Literal
from .batch_annotation import AnnotationProps

CheckPoint = Literal[None, "file_uploaded", "openai_job_created", "openai_job_running", "openai_job_succeeded", "openai_job_failed", "succeeded", "failed"]

class FineTuningInputData(BaseModel):
    dataset_id: str
    schema_id: str
    finetuning_props : AnnotationProps

class FineTuningJob(BaseModel):
    job_type: Literal["finetune-dataset"] = "finetune-dataset"
    input_data: FineTuningInputData
    checkpoint: CheckPoint = None
    checkpoint_data: Optional[dict] = None

