import datetime
from typing import Any, Optional, Self

import nanoid  # type: ignore
from pydantic import BaseModel, Field, model_validator, ConfigDict

from ..inference_settings import InferenceSettings
from .predictions import PredictionData

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

class CreateIterationRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")
    """
    Request model for performing a new iteration with custom inference settings and optional schema overrides.
    """

    inference_settings: InferenceSettings
    # Backward-compat: allow full json_schema, from which overrides will be computed and stored
    json_schema: Optional[dict[str, Any]] = None
    # Preferred: provide only non-structural overrides to apply on top of the parent/base schema
    schema_overrides: Optional[SchemaOverrides] = None
    parent_id: Optional[str] = Field(
        default=None,
        description="The ID of the parent iteration to copy the JSON Schema from.",
    )

    # validate that exactly one of parent_id or json_schema is provided
    #@model_validator(mode="after")
    #def validate_one_of_parent_id_or_json_schema(self) -> Self:
    #    if (self.parent_id is None) ^ (self.json_schema is None):
    #        return self
    #     raise ValueError("Exactly one of parent_id or json_schema must be provided")


class PatchIterationRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")
    inference_settings: Optional[InferenceSettings] = Field(default=None, description="The new inference settings of the iteration")
    # Replace full-schema editing with overrides editing. If provided, replaces the whole overrides map.
    schema_overrides: Optional[SchemaOverrides] = Field(default=None, description="Override map for non-structural schema changes")
    version: Optional[int] = Field(default=None, description="Current version for optimistic locking")

class ProcessIterationRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")
    """Request model for processing an iteration - running extractions on documents."""

    document_ids: Optional[list[str]] = Field(default=None, description="Specific document IDs to process. If None, all documents will be processed.")
    only_outdated: bool = Field(default=True, description="Only process documents that need updates (prediction.updated_at is None or older than iteration.updated_at)")


class DocumentStatus(BaseModel):
    model_config = ConfigDict(extra="ignore")
    """Status of a document within an iteration."""

    document_id: str
    filename: str
    needs_update: bool = Field(description="True if prediction is missing or outdated")
    has_prediction: bool = Field(description="True if any prediction exists")
    prediction_updated_at: Optional[datetime.datetime] = Field(description="When the prediction was last updated")
    iteration_updated_at: datetime.datetime = Field(description="When the iteration settings were last updated")


class IterationDocumentStatusResponse(BaseModel):
    model_config = ConfigDict(extra="ignore")
    """Response showing the status of all documents in an iteration."""

    iteration_id: str
    documents: list[DocumentStatus]
    total_documents: int
    documents_needing_update: int
    documents_up_to_date: int


class AddIterationFromJsonlRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")
    jsonl_gcs_path: str
