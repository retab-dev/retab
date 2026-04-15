import datetime
from typing import Any, Optional

from pydantic import BaseModel, ConfigDict, Field

from ..mime import FileRef, MIMEData
from .predictions import PredictionData


class Dataset(BaseModel):
    model_config = ConfigDict(extra="ignore")

    id: str
    name: str = Field(default="", description="The name of the dataset")
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    base_json_schema: dict[str, Any] = Field(default_factory=dict, description="The base json schema of the dataset")
    project_id: str


class CreateDatasetRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")

    name: str
    base_json_schema: dict[str, Any] = Field(default_factory=dict, description="The base json schema of the dataset")


class PatchDatasetRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")

    name: Optional[str] = Field(default=None, description="The name of the dataset")


class DatasetDocument(BaseModel):
    model_config = ConfigDict(extra="ignore")

    id: str
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    project_id: str
    dataset_id: str
    mime_data: FileRef = Field(description="The mime data of the document")
    prediction_data: PredictionData = Field(default_factory=PredictionData, description="The prediction data of the document")
    extraction_id: str | None = Field(default=None, description="The extraction id of the document")
    validation_flags: dict[str, Any] = Field(default_factory=dict, description="The validation flags of the dataset document")


class PatchDatasetDocumentRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")

    validation_flags: Optional[dict[str, Any]] = Field(default=None, description="The validation flags of the dataset document")
    prediction_data: Optional[PredictionData] = Field(default=None, description="The prediction data of the document")
    extraction_id: Optional[str] = Field(default=None, description="The extraction id of the document")
