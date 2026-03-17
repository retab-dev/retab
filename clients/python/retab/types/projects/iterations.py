import datetime
from typing import Literal, Optional

from pydantic import BaseModel, ConfigDict, Field

from ..inference_settings import InferenceSettings
from ..mime import BaseMIMEData
from .predictions import PredictionData


class SchemaOverrides(BaseModel):
    model_config = ConfigDict(extra="ignore")

    descriptionsOverride: Optional[dict[str, str]] = Field(default=None, description="Maps field path to description override")
    reasoningPromptsOverride: Optional[dict[str, str]] = Field(default=None, description="Maps field path to X-ReasoningPrompt override")


class DraftIteration(BaseModel):
    model_config = ConfigDict(extra="ignore")

    schema_overrides: SchemaOverrides = Field(default_factory=SchemaOverrides)
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    inference_settings: InferenceSettings = Field(default_factory=InferenceSettings)


class Iteration(BaseModel):
    model_config = ConfigDict(extra="ignore")

    id: str
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    inference_settings: InferenceSettings = Field(default_factory=InferenceSettings)
    schema_overrides: SchemaOverrides = Field(default_factory=SchemaOverrides)
    parent_id: Optional[str] = Field(default=None, description="The ID of the parent iteration")
    project_id: str
    dataset_id: str
    draft: DraftIteration = Field(default_factory=DraftIteration, description="The drafted changes of the iteration")
    status: Literal["draft", "completed"] = Field(default="draft", description="Iteration status: 'draft' for editable, 'completed' for finalized")


class IterationDocument(BaseModel):
    model_config = ConfigDict(extra="ignore")

    id: str
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    project_id: str
    iteration_id: str
    dataset_id: str
    dataset_document_id: str
    mime_data: BaseMIMEData = Field(description="The mime data of the document")
    prediction_data: PredictionData = Field(default_factory=PredictionData, description="The prediction data of the document")
    extraction_id: str | None = Field(default=None, description="The extraction id of the document")


