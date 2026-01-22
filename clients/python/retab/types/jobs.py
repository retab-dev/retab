"""
Jobs API Types

Pydantic models for the asynchronous Jobs API.
"""

from typing import Any, Literal

from pydantic import BaseModel, Field


JobStatus = Literal[
    "validating",
    "queued",
    "in_progress",
    "completed",
    "failed",
    "cancelled",
    "expired",
]

SupportedEndpoint = Literal[
    "/v1/documents/extract",
    "/v1/documents/parse",
    "/v1/documents/split",
    "/v1/documents/classify",
    "/v1/schemas/generate",
    "/v1/edit/agent/fill",
    "/v1/edit/templates/fill",
    "/v1/edit/templates/generate",
    "/v1/projects/extract",  # Requires "project_id" in request body
]


class JobResponse(BaseModel):
    """Response stored when job completes successfully."""
    status_code: int
    body: dict[str, Any]


class JobError(BaseModel):
    """Error details when job fails."""
    code: str
    message: str
    details: dict[str, Any] | None = None


class Job(BaseModel):
    """
    Job object representing an asynchronous operation.

    Use this to track the status of long-running operations like extract, parse,
    split, classify, schema generation, and template operations.
    """
    id: str
    object: Literal["job"] = "job"
    status: JobStatus
    endpoint: SupportedEndpoint
    request: dict[str, Any]
    response: JobResponse | None = None
    error: JobError | None = None

    # Timestamps (Unix timestamps)
    created_at: int
    started_at: int | None = None
    completed_at: int | None = None
    expires_at: int

    # User context
    organization_id: str
    metadata: dict[str, str] | None = None


class CreateJobRequest(BaseModel):
    """Request body for creating a new job."""
    endpoint: SupportedEndpoint
    request: dict[str, Any]
    metadata: dict[str, str] | None = Field(
        default=None,
        description="Max 16 pairs; keys ≤64 chars, values ≤512 chars"
    )


class JobListResponse(BaseModel):
    """Response for listing jobs."""
    object: Literal["list"] = "list"
    data: list[Job]
    first_id: str | None = None
    last_id: str | None = None
    has_more: bool = False
