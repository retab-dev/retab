"""
Jobs API Types

Pydantic models for the asynchronous Jobs API.
"""

import datetime
from typing import Any, Literal

from pydantic import ConfigDict, Field
from retab.types.base import RetabBaseModel
from retab.types.pagination import ListMetadata


JobStatus = Literal[
    "validating",
    "queued",
    "in_progress",
    "completed",
    "failed",
    "cancelled",
    "expired",
]

JobListSource = Literal[
    "api",
    "project",
    "workflow",
]

JobListOrder = Literal[
    "asc",
    "desc",
]

SupportedEndpoint = str


class JobResponse(RetabBaseModel):
    """Response stored when job completes successfully."""

    status_code: int
    body: dict[str, Any]


class JobError(RetabBaseModel):
    """Error details when job fails."""

    code: str
    message: str
    details: dict[str, Any] | None = None


class JobWarning(RetabBaseModel):
    """Non-fatal warning attached to a job.

    Mirrors the backend ``JobWarning`` shape: a structured (code, message,
    details) tuple. Surfaced on the job so SDK consumers can render
    advisory issues that didn't fail the job — e.g. a slow upstream, a
    deprecated field in the request, a partial degradation.
    """

    code: str
    message: str
    details: dict[str, Any] | None = None


class Job(RetabBaseModel):
    """
    Job object representing an asynchronous operation.

    Use this to track the status of long-running operations like extract, parse,
    split, classify, schema generation, and template operations.
    """

    id: str | None = None
    object: Literal["job"] = "job"
    status: JobStatus | None = "validating"
    endpoint: SupportedEndpoint
    request: dict[str, Any] | None = None
    response: JobResponse | None = None
    error: JobError | None = None
    warnings: list[JobWarning] | None = None

    # Timestamps (ISO 8601 datetimes)
    created_at: datetime.datetime | None = None
    started_at: datetime.datetime | None = None
    completed_at: datetime.datetime | None = None
    expires_at: datetime.datetime | None = None

    metadata: dict[str, str] | None = None

    # Retry / lifecycle tracking. Mirrors the backend's ``JobBase`` fields
    # so the SDK can surface execution observability:
    #   - ``cancelled``: explicit cancellation flag, distinct from
    #     ``status == "cancelled"`` (which is also set for expiry / cleanup).
    #   - ``attempt_count``: number of execution attempts so far. Stays at 0
    #     until the runner picks the job up.
    #   - ``last_attempt_at``: epoch seconds of the most recent attempt.
    #   - ``last_failure_code``: machine-readable code from the last failed
    #     attempt (matches ``error.code`` if the job's terminal state was
    #     error; remains set across retries to aid debugging).
    cancelled: bool | None = False
    attempt_count: int | None = 0
    last_attempt_at: datetime.datetime | None = None
    last_failure_code: str | None = None


class CreateJobRequest(RetabBaseModel):
    """Request body for creating a new job."""

    model_config = ConfigDict(extra="forbid")

    endpoint: SupportedEndpoint
    request: dict[str, Any]
    metadata: dict[str, str] | None = Field(default=None, description="Max 16 pairs; keys ≤64 chars, values ≤512 chars")


class JobListResponse(RetabBaseModel):
    """Response for listing jobs.

    Uses the canonical cursor pagination envelope ``{data, list_metadata}``
    shared by every other paginated list endpoint in the Retab API.
    To check whether more pages are available, read
    ``list_metadata.after is not None`` directly.
    """

    data: list[Job]
    list_metadata: ListMetadata
