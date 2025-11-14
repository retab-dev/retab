import datetime
from typing import Any, Optional

import nanoid  # type: ignore
from pydantic import BaseModel, Field, ConfigDict

from ..inference_settings import InferenceSettings

default_inference_settings = InferenceSettings(
    model="auto-small",
    temperature=0.5,
    reasoning_effort="minimal",
    modality="native",
    image_resolution_dpi=192,
    browser_canvas="A4",
    n_consensus=1,
)

class HilCriterion(BaseModel):
    path: str
    agentic_fix: bool = Field(default=False, description="Whether to use agentic fix for the criterion")

class HumanInTheLoopParams(BaseModel):
    enabled: bool = Field(default=False)
    url: str = Field(default="", description="The URL of the human in the loop endpoint")
    headers: dict[str, str] = Field(default_factory=dict, description="The headers to send to the human in the loop endpoint")
    criteria: list[HilCriterion] = Field(default_factory=list, description="The criteria to use for the human in the loop")

class PublishedConfig(BaseModel):
    inference_settings: InferenceSettings = default_inference_settings
    json_schema: dict[str, Any] = Field(default_factory=dict, description="The json schema of the project")
    human_in_the_loop_params: HumanInTheLoopParams = Field(default_factory=HumanInTheLoopParams)
    origin: str = Field(default="manual", description="The origin of the published config. Either 'Manual' or the iteration id that was used to generate the config")
class DraftConfig(BaseModel):
    inference_settings: InferenceSettings = default_inference_settings
    json_schema: dict[str, Any] = Field(default_factory=dict, description="The json schema of the builder config")
    human_in_the_loop_criteria: list[HilCriterion] = Field(default_factory=list)
class Project(BaseModel):
    model_config = ConfigDict(extra="ignore")
    id: str = Field(default_factory=lambda: "project_" + nanoid.generate())
    name: str = Field(default="", description="The name of the project")
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    published_config: PublishedConfig
    draft_config: DraftConfig
    is_published: bool = False
    #computation_spec: ComputationSpec = Field(default_factory=ComputationSpec, description="The computation spec of the project")
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
    #computation_spec: Optional[ComputationSpec] = Field(default=None, description="The computation spec of the project")