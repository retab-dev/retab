from typing import Literal, Optional, Self

from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel, Field, model_validator

from ..modalities import Modality

SelectionMode = Literal["all", "manual"]


# This is the input data for the prepare_dataset job
class PrepareDatasetInputData(BaseModel):
    dataset_id: Optional[str] = None
    schema_id: Optional[str] = None
    schema_data_id: Optional[str] = None

    selection_model: SelectionMode = "all"

    @model_validator(mode="after")
    def validate_input(self) -> Self:
        # The preference is:
        # 1. dataset_id
        # 2. schema_id
        # 3. schema_data_id
        if self.dataset_id is None and self.schema_id is None and self.schema_data_id is None:
            raise ValueError("At least one of dataset_id, schema_id, or schema_data_id must be provided")

        return self


# This is the input data for the split_dataset job
class DatasetSplitInputData(BaseModel):
    dataset_id: str
    train_size: Optional[int | float] = None
    eval_size: Optional[int | float] = None

    @model_validator(mode="after")
    def validate_input(self) -> Self:
        if self.train_size is not None and self.eval_size is not None:
            raise ValueError("train_size and eval_size cannot both be provided")
        return self


# This is the input data for the batch annotation job
class InferenceSettings(BaseModel):
    model: str = "gpt-4o-mini"
    temperature: float = 0.0
    modality: Modality = "native"
    reasoning_effort: ChatCompletionReasoningEffort = "medium"
    image_resolution_dpi: int = 96
    browser_canvas: Literal['A3', 'A4', 'A5'] = 'A4'
    n_consensus: int = Field(default=1, description="Number of consensus rounds to perform")


class AnnotationInputData(BaseModel):
    data_file: str
    schema_id: str
    inference_settings: InferenceSettings


# This is the input data for the evaluation job
class EvaluationInputData(BaseModel):
    eval_data_file: str
    schema_id: str
    inference_settings_1: InferenceSettings | None = None
    inference_settings_2: InferenceSettings


# from pydantic import BaseModel, Field, model_validator
# from typing import Literal, Optional, Any
# import datetime


# JobType = Literal["prompt-optimization", "annotate-files", "finetune-dataset", "webcrawl"]
# JobStatus = Literal["pending", "running", "completed", "failed"]
#### JOBS ####

# class JobTemplateCreateRequest(BaseModel):
#     job_type: JobType
#     default_input_data: dict = Field(default_factory=dict)
#     description: Optional[str] = None
#     cron: Optional[str] = None


# class JobTemplateDocument(BaseModel):
#     object: Literal["job_template"] = "job_template"
#     id: str
#     type: JobType
#     identity: Any | None = None
#     description: Optional[str] = None
#     default_input_data: dict = Field(default_factory=dict)
#     # For scheduled jobs, include a valid CRON expression (None for on-demand only jobs)
#     cron: Optional[str] = None
#     next_run: Optional[datetime.datetime] = None
#     created_at: Optional[datetime.datetime] = None
#     updated_at: Optional[datetime.datetime] = None
#     is_active: bool = True  # Change to status.

# class JobTemplateUpdateRequest(BaseModel):
#     cron: Optional[str] = None
#     default_input_data: Optional[dict] = None
#     description: Optional[str] = None
#     is_active: Optional[bool] = None    # Change to status.


#### EXECUTIONS ####

# class JobExecutionCreateRequest(BaseModel):
#     type: JobType
#     template_id: Optional[str] = None
#     input_data: dict = Field(default_factory=dict)

#     @model_validator(mode='before')
#     @classmethod
#     def validate_job_identifiers(cls, data: Any) -> Any:
#         if isinstance(data, dict):
#             if bool(data.get('job_type')) == bool(data.get('job_template_id')):
#                 raise ValueError("Either job_type or job_template_id must be provided")
#         return data

# class JobExecutionResponse(BaseModel):
#     id: str
#     template_id: Optional[str] = None
#     type: JobType
#     status: JobStatus
#     result: Optional[dict] = None
#     error: Optional[str] = None
#     created_at: Optional[datetime.datetime] = None
#     updated_at: Optional[datetime.datetime] = None

# class JobExecutionDocument(BaseModel):
#     object: Literal["job_execution"] = "job_execution"
#     id: str
#     template_id: Optional[str] = None
#     type: JobType
#     identity: Any | None = None
#     status: JobStatus
#     input_data_gcs_path: str
#     result: Optional[dict] = None
#     error: Optional[str] = None
#     created_at: Optional[datetime.datetime] = None
#     updated_at: Optional[datetime.datetime] = None
#     checkpoint: Any = None  # Useful for jobs that need to be resumed
#     checkpoint_data: Optional[dict] = None
#     needs: list[str] = Field(default_factory=list, description="list of jobs execution id that must be completed before this job can run")


# class Workflow(BaseModel):
#     name: str
#     jobs: list[JobExecutionDocument]
