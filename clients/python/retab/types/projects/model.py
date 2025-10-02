import datetime
from typing import Any, Optional

import nanoid  # type: ignore
from pydantic import BaseModel, Field, ConfigDict

from .documents import ProjectDocument
from .iterations import Iteration
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

class SheetsIntegration(BaseModel):
    sheet_id: str
    spreadsheet_id: str

class BaseProject(BaseModel):
    model_config = ConfigDict(extra="ignore")
    id: str = Field(default_factory=lambda: "proj_" + nanoid.generate())
    name: str = Field(default="", description="The name of the project")
    json_schema: dict[str, Any] = Field(default_factory=dict, description="The json schema of the project")
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    sheets_integration: SheetsIntegration | None = None
    validation_flags: dict[str, Any] | None = None
    hardcoded_keys: Optional[dict[str, str]] = Field(default=None, description="hardcoded keys to be used for the extraction of long lists of data", examples=[{"properties": "ID", "products": "identity.id"}])
    inference_settings: InferenceSettings = default_inference_settings

# Actual Object stored in DB
class Project(BaseProject):
    documents: list[ProjectDocument] = Field(default_factory=list)
    iterations: list[Iteration] = Field(default_factory=list)

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
    sheets_integration: SheetsIntegration | None = None
    validation_flags: Optional[dict[str, Any]] = Field(default=None, description="The validation flags of the project")
    hardcoded_keys: Optional[dict[str, str]] = Field(default=None, description="hardcoded keys to be used for the extraction of long lists of data", examples=[{"properties": "ID", "products": "identity.id"}])
    inference_settings: Optional[InferenceSettings] = Field(default=None, description="The inference settings of the project")

class AddIterationFromJsonlRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")
    jsonl_gcs_path: str
