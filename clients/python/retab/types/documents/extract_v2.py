"""
Types for the v2 extract endpoint (``POST /v1/extractions``).

Self-contained (no ``openai`` imports). Mirrors the shape used by
``ClassifyResponse`` in ``classify.py``: ``data + consensus{choices, likelihoods}
+ usage``.

The legacy extract endpoint (``POST /v1/documents/extract``) continues to use
the types in ``extract.py``; this module is only the clean v2 shape.
"""

from __future__ import annotations

import datetime
from typing import Any

from pydantic import BaseModel, ConfigDict, Field

from ..mime import MIMEData
from .usage import RetabUsage


# ---------------------------------------------------------------------------
# Request
# ---------------------------------------------------------------------------


class ExtractRequest(BaseModel):
    model_config = ConfigDict(arbitrary_types_allowed=True, extra="ignore")

    document: MIMEData = Field(..., description="Document to extract from.")
    json_schema: dict[str, Any] = Field(..., description="JSON schema describing the structured output.")
    model: str = Field(..., description="Model used for the extraction.")
    n_consensus: int = Field(default=1, ge=1, description="Number of voters. >=2 enables consensus.")
    image_resolution_dpi: int = Field(default=192, ge=96, le=300)
    stream: bool = Field(default=False)
    chunking_keys: dict[str, str] | None = Field(
        default=None,
        description="Optional keys for parallel OCR chunking of long list fields.",
        examples=[{"properties": "ID", "products": "identity.id"}],
    )
    metadata: dict[str, str] = Field(default_factory=dict)
    extraction_id: str | None = Field(default=None, description="Client-chosen id. Server generates one if omitted.")
    bust_cache: bool = Field(default=False)


# ---------------------------------------------------------------------------
# Response — data / consensus{choices, likelihoods} / usage
# ---------------------------------------------------------------------------


class ExtractChoice(BaseModel):
    """One voter's raw extraction output."""

    data: dict[str, Any] = Field(..., description="Extracted data from this voter.")


class ExtractConsensus(BaseModel):
    choices: list[ExtractChoice] = Field(
        default_factory=list,
        description="Per-voter extraction results used to build the consolidated `data`.",
    )
    likelihoods: dict[str, Any] | None = Field(
        default=None,
        description=(
            "Agreement signal mirroring the shape of `data`. Scalar leaves carry voter-agreement "
            "in [0, 1]; list leaves carry one entry per matched list item."
        ),
    )


class ExtractResponse(BaseModel):
    model_config = ConfigDict(extra="ignore")

    data: dict[str, Any] = Field(..., description="Consolidated extraction result.")
    consensus: ExtractConsensus = Field(
        default_factory=ExtractConsensus,
        description="Consensus metadata for multi-voter extraction runs.",
    )
    usage: RetabUsage = Field(..., description="Usage information for the extraction.")

    extraction_id: str | None = None
    model: str
    request_at: datetime.datetime | None = None
    first_token_at: datetime.datetime | None = None
    last_token_at: datetime.datetime | None = None


# ---------------------------------------------------------------------------
# Streaming chunk
# ---------------------------------------------------------------------------


class ExtractResponseChunk(BaseModel):
    """Incremental chunk for streaming extraction.

    Deltas apply to the accumulated `data` and `likelihoods`. `deleted_keys`
    names keys to remove from the accumulator.
    """

    model_config = ConfigDict(extra="ignore")

    data_delta: dict[str, Any] = Field(default_factory=dict)
    likelihoods_delta: dict[str, Any] | None = None
    deleted_keys: list[str] = Field(default_factory=list)
    is_valid_json: bool = False

    usage: RetabUsage | None = None
    extraction_id: str | None = None
    request_at: datetime.datetime | None = None
    first_token_at: datetime.datetime | None = None
    last_token_at: datetime.datetime | None = None


__all__ = [
    "ExtractRequest",
    "ExtractChoice",
    "ExtractConsensus",
    "ExtractResponse",
    "ExtractResponseChunk",
]
