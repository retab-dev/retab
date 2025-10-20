import datetime
from typing import Any, Optional

import nanoid  # type: ignore
from pydantic import BaseModel, Field, ConfigDict

from ..inference_settings import InferenceSettings
from ..mime import MIMEData
from .predictions import PredictionData, PredictionMetadata
from .documents import ProjectDocument

from typing import Self
from pydantic import model_validator

class DatasetDocument(BaseModel):
    model_config = ConfigDict(extra="ignore")
    id: str = Field(default_factory=lambda: "dataset_doc_" + nanoid.generate(), description="The ID of the document. Equal to mime_data.id but robust to the case where mime_data is a BaseMIMEData")
    mime_data: MIMEData = Field(description="The mime data of the document. Can also be a BaseMIMEData, which is why we have this id field (to be able to identify the file, but id is equal to mime_data.id)")
    annotation: dict[str, Any] = Field(default={}, description="The ground truth of the document")
    annotation_metadata: Optional[PredictionMetadata] = Field(default=None, description="The metadata of the annotation when the annotation is a prediction")
    validation_flags: dict[str, Any] | None = None

default_inference_settings = InferenceSettings(
    model="auto-small",
    temperature=0.5,
    reasoning_effort="minimal",
    modality="native",
    image_resolution_dpi=192,
    browser_canvas="A4",
    n_consensus=1,
)

class Dataset(BaseModel):
    id: str = Field(default_factory=lambda: "dataset_" + nanoid.generate())
    name: str = Field(default="", description="The name of the dataset")
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    inference_settings: InferenceSettings = default_inference_settings
    documents: list[DatasetDocument] = Field(default_factory=list)
    iteration_ids: list[str] = Field(default_factory=list)


class SchemaOverrides(BaseModel):
    model_config = ConfigDict(extra="ignore")
    """Schema override for a field path. Only supports non-structural metadata.

    - description: JSON Schema description string
    - reasoning_prompt: value mapped to schema key "X-ReasoningPrompt"
    """

    descriptionsOverride: Optional[dict[str, str]] = None
    reasoningPromptsOverride: Optional[dict[str, str]] = Field(default=None, description="Maps to X-ReasoningPrompt in schema")

class BaseIteration(BaseModel):
    model_config = ConfigDict(extra="ignore")
    id: str = Field(default_factory=lambda: "eval_iter_" + nanoid.generate())
    parent_id: Optional[str] = Field(default=None, description="The ID of the parent iteration")
    inference_settings: InferenceSettings
    # Store only overrides rather than the full schema. Keys are dot-paths like "address.street" or "items.*.price".
    schema_overrides: SchemaOverrides = Field(
        default_factory=SchemaOverrides, description="Map of field path -> non-structural schema overrides (description, reasoning_prompt)"
    )
    updated_at: datetime.datetime = Field(
        default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc),
        description="The last update date of inference settings or schema overrides",
    )
    
class DraftIteration(BaseModel):
    model_config = ConfigDict(extra="ignore")
    # Store draft overrides only.
    schema_overrides: SchemaOverrides = Field(default_factory=SchemaOverrides)
    updated_at: datetime.datetime = Field(
        default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc),
        description="The last update date of draft schema overrides",
    )

class Iteration(BaseIteration):
    model_config = ConfigDict(extra="ignore")
    predictions: dict[str, PredictionData] = Field(default_factory=dict, description="The predictions of the iteration for all the documents")
    draft: Optional[DraftIteration] = Field(default=None, description="The draft iteration of the iteration")

    # if no draft is provided, set it to the current iteration
    @model_validator(mode="after")
    def set_draft_to_current_iteration(self) -> Self:
        if self.draft is None:
            self.draft = DraftIteration(
                schema_overrides=SchemaOverrides(),
                updated_at=datetime.datetime.now(tz=datetime.timezone.utc),
            )
        return self

class Evaluation(BaseModel):
    id: str = Field(default_factory=lambda: "eval_" + nanoid.generate())
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    dataset_id: str
    iteration_ids: list[str] = Field(default_factory=list)

class PublishedConfig(BaseModel):
    inference_settings: InferenceSettings = default_inference_settings
    json_schema: dict[str, Any] = Field(default_factory=dict, description="The json schema of the project")
    human_in_the_loop_criteria: list[str] = Field(default_factory=list)

class DraftConfig(BaseModel):
    inference_settings: InferenceSettings = default_inference_settings
    schema_overrides: SchemaOverrides = Field(default_factory=SchemaOverrides)
    human_in_the_loop_criteria: list[str] = Field(default_factory=list)

class Project(BaseModel):
    model_config = ConfigDict(extra="ignore")
    id: str = Field(default_factory=lambda: "project_" + nanoid.generate())
    name: str = Field(default="", description="The name of the project")
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    dataset_ids: list[str] = Field(default_factory=list)
    is_published: bool = False
    published_config: PublishedConfig
    draft_config: DraftConfig


# Actual Object stored in DB

class CreateProjectRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")
    name: str
    json_schema: dict[str, Any]


# This is basically the same as BaseProject, but everything is optional.
# Could be achieved by convert_basemodel_to_partial_basemodel(BaseProject) but we prefer explicitness
class PatchProjectRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")
    name: Optional[str] = Field(default=None, description="The name of the document")
    json_schema: Optional[dict[str, Any]] = Field(default=None, description="The json schema of the project")
    validation_flags: Optional[dict[str, Any]] = Field(default=None, description="The validation flags of the project")
    inference_settings: Optional[InferenceSettings] = Field(default=None, description="The inference settings of the project")

class AddIterationFromJsonlRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")
    jsonl_gcs_path: str
