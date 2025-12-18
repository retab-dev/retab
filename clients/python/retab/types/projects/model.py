import datetime
from typing import Any, Optional

import nanoid  # type: ignore
from pydantic import BaseModel, Field, ConfigDict

from ..mime import BaseMIMEData, MIMEData
from ..inference_settings import InferenceSettings
from .predictions import PredictionData

default_inference_settings = InferenceSettings(
    model="retab-small",
    temperature=0.5,
    reasoning_effort="minimal",
    image_resolution_dpi=192,
    n_consensus=1,
)

class Computation(BaseModel):
    expression: str = Field(description="The expression to use for the computation")

class ComputationSpec(BaseModel):
    computations: dict[str, Computation] = Field(default_factory=dict, description="The computations to use for the project")

class HilCriterion(BaseModel):
    path: str
    agentic_fix: bool = Field(default=False, description="Whether to use agentic fix for the criterion")

class DraftConfig(BaseModel):
    inference_settings: InferenceSettings = Field(default=default_inference_settings, description="The inference settings of the project")
    json_schema: dict[str, Any] = Field(default_factory=dict, description="The json schema of the builder config")
    human_in_the_loop_criteria: list[HilCriterion] = Field(default_factory=list)
    computation_spec: ComputationSpec = Field(default_factory=ComputationSpec, description="The computation spec of the project")

class PublishedConfig(DraftConfig):
    origin: str = Field(default="manual", description="The origin of the published config. Either 'Manual' or the iteration id that was used to generate the config")


class Project(BaseModel):
    model_config = ConfigDict(extra="ignore")
    id: str = Field(default_factory=lambda: "project_" + nanoid.generate())
    name: str = Field(default="", description="The name of the project")
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    published_config: PublishedConfig
    draft_config: DraftConfig
    is_published: bool = False
    is_schema_generated: bool = Field(default=True, description="Whether the schema has been generated for the project")

class StoredProject(Project):
    """Project model with organization_id for database storage"""
    organization_id: str

class CreateProjectRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")
    name: str
    json_schema: dict[str, Any] = Field(default_factory=dict, description="The json schema of the project")

# This is basically the same as Project, but everything is optional.
class PatchProjectRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")
    name: Optional[str] = Field(default=None, description="The name of the document")
    published_config: Optional[PublishedConfig] = Field(default=None, description="The published config of the project")
    draft_config: Optional[DraftConfig] = Field(default=None, description="The draft config of the project")
    is_published: Optional[bool] = Field(default=None, description="The published status of the project")
    computation_spec: Optional[ComputationSpec] = Field(default=None, description="The computation spec of the project")
# ----------------------------
# ----------------------------
# ----------------------------

class BuilderDocument(BaseModel):
    model_config = ConfigDict(extra="ignore")
    id: str = Field(default_factory=lambda: "builder_doc_" + nanoid.generate())
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    project_id: str
    mime_data: BaseMIMEData = Field(description="The mime data of the document")
    
    prediction_data: PredictionData = Field(default=PredictionData(), description="The prediction data of the document")
    extraction_id: str | None = Field(default=None, description="The extraction id of the document")

class StoredBuilderDocument(BuilderDocument):
    """Builder document model with organization_id and project_id for database storage"""
    organization_id: str

class PatchBuilderDocumentRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")
    extraction_id: Optional[str] = Field(default=None, description="The extraction id of the builder document")
    prediction_data: Optional[PredictionData] = Field(default=None, description="The prediction data of the document")


class AddBuilderDocumentRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")
    mime_data: MIMEData
    prediction_data: PredictionData = Field(default=PredictionData(), description="The prediction data of the document")
    project_id: str