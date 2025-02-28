from pydantic import BaseModel, Field
from typing import Optional, Literal
from .batch_annotation import AnnotationProps

CheckPoint = Literal[None, "file_uploaded", "openai_job_created", "openai_job_running", "openai_job_succeeded", "openai_job_failed", "succeeded", "failed"]

class FineTuningInputData(BaseModel):
    dataset_id: str
    finetuning_props : AnnotationProps

class FineTuningJob(BaseModel):
    job_type: Literal["finetune-dataset"] = "finetune-dataset"
    input_data: FineTuningInputData
    checkpoint: CheckPoint = None
    checkpoint_data: Optional[dict] = None

from typing import Any
class Evaluation(BaseModel): 
    dataset_id: str # or job_id: str -> You can create and evaluation job that will create annotations for the two models, and then you can evaluate the two models on the same annotations.
    reference_model: str
    other_model: str
    #reference_likelihood: dict[str,Any]
    #other_likelihood: dict[str,Any]
    levensthein_distance: dict[str,Any]
    #... there is another per field distance that is good : 
    # - levensthein_distance
    # - jaccard_distance ?
    # - hamming_distance ?