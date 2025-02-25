from pydantic import BaseModel, Field, model_validator
from typing import Literal, Optional, Any
import datetime


JobType = Literal["prompt-optimization", "annotate-files", "finetune-dataset", "webcrawl"]
JobName = Literal["uiform-production-backend-jobs", "uiform-staging-backend-jobs"]
JobStatus = Literal["pending", "running", "completed", "failed"]
#### JOBS ####

class JobTemplateCreateRequest(BaseModel):
    job_type: JobType
    default_input_data: dict = Field(default_factory=dict)
    description: Optional[str] = None
    cron: Optional[str] = None


class JobTemplateDocument(BaseModel):
    object: Literal["job_template"] = "job_template"
    id: str
    type: JobType
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
    type: JobType
    template_id: Optional[str] = None
    input_data: dict = Field(default_factory=dict)

    @model_validator(mode='before')
    @classmethod
    def validate_job_identifiers(cls, data: Any) -> Any:
        if isinstance(data, dict):
            if bool(data.get('job_type')) == bool(data.get('job_template_id')):
                raise ValueError("Either job_type or job_template_id must be provided")
        return data

class JobExecutionResponse(BaseModel):
    id: str
    template_id: Optional[str] = None
    type: JobType
    status: JobStatus
    result: Optional[dict] = None
    error: Optional[str] = None
    created_at: Optional[datetime.datetime] = None
    updated_at: Optional[datetime.datetime] = None

class JobExecutionDocument(BaseModel):
    object: Literal["job_execution"] = "job_execution"
    id: str
    template_id: Optional[str] = None
    type: JobType
    identity: Any | None = None
    status: JobStatus
    input_data_gcs_path: str
    result: Optional[dict] = None
    error: Optional[str] = None
    created_at: Optional[datetime.datetime] = None
    updated_at: Optional[datetime.datetime] = None
    checkpoint: Any = None  # Useful for jobs that need to be resumed
    checkpoint_data: Optional[dict] = None
    runs_on: list[str] = Field(default_factory=list, description="list of jobs execution id that must be completed before this job can run")


class Workflow(BaseModel):
    name: str
    jobs: list[JobExecutionDocument]

