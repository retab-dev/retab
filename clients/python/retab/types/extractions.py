import datetime
from typing import Any, Literal, Optional

from pydantic import BaseModel, ConfigDict, Field


class ExtractionFile(BaseModel):
    """File metadata associated with an extraction."""

    id: str
    filename: str
    mime_type: str


class ProcessingRequestOrigin(BaseModel):
    """Origin of the extraction request."""

    type: str
    id: Optional[str] = None


class ExtractionInferenceSettings(BaseModel):
    """Inference settings used for the extraction."""

    model: str = "retab-small"
    image_resolution_dpi: int = 192
    n_consensus: int = 1
    chunking_keys: Optional[dict[str, str]] = None


class Extraction(BaseModel):
    """A stored extraction record from the Retab API."""

    model_config = ConfigDict(extra="allow")

    id: str
    file: ExtractionFile
    predictions: Optional[dict[str, Any]] = None
    likelihoods: Optional[dict[str, Any]] = None
    consensus_details: Optional[list[dict[str, Any]]] = None
    origin: ProcessingRequestOrigin
    inference_settings: ExtractionInferenceSettings
    original_model: Optional[str] = None
    json_schema: dict[str, Any] = Field(default_factory=dict)
    metadata: dict[str, str] = Field(default_factory=dict)
    created_at: Optional[datetime.datetime] = None
    updated_at: Optional[datetime.datetime] = None
    organization_id: Optional[str] = None


class SourcesResponse(BaseModel):
    """Response from the extraction sources endpoint."""

    model_config = ConfigDict(extra="allow")

    object: Literal["extraction.sources"] = "extraction.sources"
    extraction_id: str
    document_type: Optional[str] = None
    file: Optional[ExtractionFile] = None
    extraction: dict[str, Any] = Field(default_factory=dict)
    sources: dict[str, Any] = Field(default_factory=dict)
