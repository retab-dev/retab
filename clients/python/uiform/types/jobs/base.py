from pydantic import BaseModel, Field
from typing import Literal, Optional, Any
import datetime


JobType = Literal["prompt-optimization", "annotate-files", "finetune-dataset"]
JobName = Literal["uiform-production-backend-jobs", "uiform-staging-backend-jobs"]


class JobCreateRequest(BaseModel):
    job_type: JobType
    input_data: dict = Field(default_factory=dict)

class JobResponse(BaseModel):
    job_id: str
    job_type: str
    status: str
    result: Optional[dict] = None
    error: Optional[str] = None
    created_at: Optional[datetime.datetime] = None
    updated_at: Optional[datetime.datetime] = None

class JobDocument(BaseModel):
    job_id: str
    job_type: str
    identity: Any
    status: str
    input_data_gcs_path: str
    result: Optional[dict] = None
    error: Optional[str] = None
    created_at: Optional[datetime.datetime] = None
    updated_at: Optional[datetime.datetime] = None
    checkpoint: Any = None  # Useful for jobs that need to be resumed
    checkpoint_data: Optional[dict] = None

