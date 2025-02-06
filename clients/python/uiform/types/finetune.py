from pydantic import BaseModel, Field
from typing import Optional, Literal
from .batch_annotation import AnnotationProps

class FineTuningInputData(BaseModel):
    dataset_id: str
    finetuning_props : AnnotationProps

class FineTuningJob(BaseModel):
    job_type: Literal["finetune-dataset"]
    input_data: FineTuningInputData
