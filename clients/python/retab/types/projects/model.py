import datetime
from typing import Any, Optional

import nanoid  # type: ignore
from pydantic import BaseModel, Field

from .documents import ProjectDocument
from .iterations import Iteration


class BaseProject(BaseModel):
    id: str = Field(default_factory=lambda: "proj_" + nanoid.generate())
    name: str = Field(default="", description="The name of the project")
    json_schema: dict[str, Any] = Field(default_factory=dict, description="The json schema of the project")
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))


# Actual Object stored in DB
class Project(BaseProject):
    documents: list[ProjectDocument] = Field(default_factory=list)
    iterations: list[Iteration] = Field(default_factory=list)

class CreateProjectRequest(BaseModel):
    name: str
    json_schema: dict[str, Any]


# This is basically the same as BaseProject, but everything is optional.
# Could be achieved by convert_basemodel_to_partial_basemodel(BaseProject) but we prefer explicitness
class PatchProjectRequest(BaseModel):
    name: Optional[str] = Field(default=None, description="The name of the document")
    json_schema: Optional[dict[str, Any]] = Field(default=None, description="The json schema of the project")


class AddIterationFromJsonlRequest(BaseModel):
    jsonl_gcs_path: str
