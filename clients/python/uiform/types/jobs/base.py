from pydantic import BaseModel, Field
from typing import Literal, Optional, Any
import datetime


JobType = Literal["prompt-optimization", "annotate-files", "finetune-dataset", "webcrawl"]
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
    identity: Any | None = None
    description: Optional[str] = None
    default_input_data: dict = Field(default_factory=dict)
    # For scheduled jobs, include a valid CRON expression (None for on-demand only jobs)
    cron: Optional[str] = None
    next_run: Optional[datetime.datetime] = None
    created_at: Optional[datetime.datetime] = None
    updated_at: Optional[datetime.datetime] = None
    is_active: bool = True  # Change to status.

class JobTemplateUpdateRequest(BaseModel):
    cron: Optional[str] = None
    default_input_data: Optional[dict] = None
    description: Optional[str] = None
    is_active: Optional[bool] = None    # Change to status.


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

