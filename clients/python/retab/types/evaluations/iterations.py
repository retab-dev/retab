import datetime
from typing import Any, Optional, Self

import nanoid  # type: ignore
from pydantic import BaseModel, Field, model_validator

from ..inference_settings import InferenceSettings
from ..predictions import PredictionData


class BaseIteration(BaseModel):
    id: str = Field(default_factory=lambda: "eval_iter_" + nanoid.generate())
    inference_settings: InferenceSettings
    json_schema: dict[str, Any]
    updated_at: datetime.datetime = Field(
        default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc),
        description="The last update date of inference settings or json schema",
    )

class Iteration(BaseIteration):
    predictions: dict[str, PredictionData] = Field(default_factory=dict, description="The predictions of the iteration for all the documents")

class CreateIterationRequest(BaseModel):
    """
    Request model for performing a new iteration with custom inference settings and optional JSON schema.
    """

    inference_settings: InferenceSettings
    json_schema: Optional[dict[str, Any]] = None
    from_iteration_id: Optional[str] = Field(
        default=None,
        description="The ID of the iteration to copy the JSON Schema from.",
    )

    # validate that exactly one of from_iteration_id or json_schema is provided
    @model_validator(mode="after")
    def validate_one_of_from_iteration_id_or_json_schema(self) -> Self:
        if (self.from_iteration_id is None) ^ (self.json_schema is None):
            return self
        raise ValueError("Exactly one of from_iteration_id or json_schema must be provided")


class PatchIterationRequest(BaseModel):
    inference_settings: Optional[InferenceSettings] = Field(default=None, description="The new inference settings of the iteration")
    json_schema: Optional[dict[str, Any]] = Field(default=None, description="The new json schema of the iteration")
    version: Optional[int] = Field(default=None, description="Current version for optimistic locking")


class ProcessIterationRequest(BaseModel):
    """Request model for processing an iteration - running extractions on documents."""

    document_ids: Optional[list[str]] = Field(default=None, description="Specific document IDs to process. If None, all documents will be processed.")
    only_outdated: bool = Field(default=True, description="Only process documents that need updates (prediction.updated_at is None or older than iteration.updated_at)")


class DocumentStatus(BaseModel):
    """Status of a document within an iteration."""

    document_id: str
    filename: str
    needs_update: bool = Field(description="True if prediction is missing or outdated")
    has_prediction: bool = Field(description="True if any prediction exists")
    prediction_updated_at: Optional[datetime.datetime] = Field(description="When the prediction was last updated")
    iteration_updated_at: datetime.datetime = Field(description="When the iteration settings were last updated")


class IterationDocumentStatusResponse(BaseModel):
    """Response showing the status of all documents in an iteration."""

    iteration_id: str
    documents: list[DocumentStatus]
    total_documents: int
    documents_needing_update: int
    documents_up_to_date: int


class AddIterationFromJsonlRequest(BaseModel):
    jsonl_gcs_path: str
