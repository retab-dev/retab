from __future__ import annotations

import datetime
from typing import Any, Literal, Optional

from pydantic import BaseModel, ConfigDict, Field

from .documents.usage import RetabUsage
from .mime import FileRef, MIMEData


class ExtractionRequest(BaseModel):
    document: MIMEData = Field(..., description="The document to extract from")
    json_schema: dict[str, Any] = Field(..., description="JSON schema describing the structured output")
    model: str = Field(default="retab-small", description="The model to use for the extraction")
    image_resolution_dpi: int = Field(
        default=192,
        ge=96,
        le=300,
        description="Resolution of the image sent to the LLM",
    )
    instructions: Optional[str] = Field(
        default=None,
        description="Free-form instructions appended to the system prompt to steer the extraction.",
    )
    n_consensus: int = Field(
        default=1,
        ge=1,
        le=16,
        description="Number of consensus extraction runs to perform. Uses deterministic single-pass when set to 1.",
    )
    metadata: dict[str, str] = Field(
        default_factory=dict,
        description="User-defined metadata to associate with this extraction",
    )
    additional_messages: list[dict[str, Any]] | None = Field(
        default=None,
        description="Additional chat messages forwarded to the extraction model.",
    )
    bust_cache: bool = Field(default=False, description="If true, skip the LLM cache and force a fresh completion")


class ProcessingRequestOrigin(BaseModel):
    """Origin of the extraction request (extract-specific)."""

    type: str
    id: Optional[str] = None


class ExtractionConsensus(BaseModel):
    choices: list[dict[str, Any]] = Field(
        default_factory=list,
        description="Alternative extraction vote outputs used to build the consolidated result.",
    )
    likelihoods: Optional[dict[str, Any]] = Field(
        default=None,
        description=(
            "Consensus likelihood tree mirroring the extraction output. "
            "Scalar leaves carry per-value voter-agreement in [0, 1]; list leaves "
            "carry one entry per matched list item."
        ),
    )


class Extraction(BaseModel):
    """A stored extraction record from the Retab API."""

    model_config = ConfigDict(extra="allow")

    id: str = Field(..., description="Unique identifier of the extraction")
    file: FileRef = Field(..., description="Information about the extracted file")

    model: str = Field(..., description="Model used for the extraction")
    json_schema: dict[str, Any] = Field(..., description="JSON schema used for the extraction")
    n_consensus: int = Field(default=1, description="Number of consensus votes used")
    image_resolution_dpi: int = Field(default=192, description="DPI used to render document images")
    instructions: Optional[str] = Field(
        default=None,
        description="Free-form instructions supplied with the extraction request.",
    )

    output: dict[str, Any] = Field(..., description="The extracted structured data")

    consensus: ExtractionConsensus = Field(
        default_factory=ExtractionConsensus,
        description="Consensus metadata for multi-vote extraction runs",
    )

    origin: Optional[ProcessingRequestOrigin] = Field(
        default=None,
        description="Origin of the extraction request",
    )
    metadata: dict[str, str] = Field(default_factory=dict)

    usage: Optional[RetabUsage] = Field(default=None, description="Usage information for the extraction")
    created_at: Optional[datetime.datetime] = None
    updated_at: Optional[datetime.datetime] = None
    organization_id: Optional[str] = None


class SourcesResponse(BaseModel):
    """Response from the extraction sources endpoint."""

    model_config = ConfigDict(extra="allow")

    object: Literal["extraction.sources"] = "extraction.sources"
    extraction_id: str
    document_type: Optional[str] = None
    file: Optional[FileRef] = None
    extraction: dict[str, Any] = Field(default_factory=dict)
    sources: dict[str, Any] = Field(default_factory=dict)
