from pydantic import BaseModel, Field
from typing import Literal, Optional, Any
import datetime


JobType = Literal["prompt-optimization", "annotate-files", "finetune-dataset"]
JobName = Literal["uiform-production-backend-jobs", "uiform-staging-backend-jobs"]

#### JOBS ####

class JobTemplateCreateRequest(BaseModel):
    job_type: JobType
    default_input_data: dict = Field(default_factory=dict)
    description: Optional[str] = None
    cron: Optional[str] = None


class JobTemplateDocument(BaseModel):
    job_template_id: str
    job_type: JobType
    description: Optional[str] = None
    # For scheduled jobs, include a valid CRON expression (None for on-demand only jobs)
    cron: Optional[str] = None
    default_input_data: dict = Field(default_factory=dict)
    created_at: Optional[datetime.datetime] = None
    updated_at: Optional[datetime.datetime] = None


#### EXECUTIONS ####

class JobExecutionCreateRequest(BaseModel):
    job_type: JobType
    job_template_id: Optional[str] = None
    input_data: dict = Field(default_factory=dict)

class JobExecutionResponse(BaseModel):
    job_execution_id: str
    job_template_id: Optional[str] = None
    job_type: str
    status: str
    result: Optional[dict] = None
    error: Optional[str] = None
    created_at: Optional[datetime.datetime] = None
    updated_at: Optional[datetime.datetime] = None

class JobExecutionDocument(BaseModel):
    job_execution_id: str
    job_template_id: Optional[str] = None
    job_type: str
    identity: Any | None = None
    status: str
    input_data_gcs_path: str
    result: Optional[dict] = None
    error: Optional[str] = None
    created_at: Optional[datetime.datetime] = None
    updated_at: Optional[datetime.datetime] = None
    checkpoint: Any = None  # Useful for jobs that need to be resumed
    checkpoint_data: Optional[dict] = None

