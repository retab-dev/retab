"""Pydantic models for the workflow reviews."""

from __future__ import annotations

import datetime
from typing import Any, Literal, TypeAlias

from pydantic import ConfigDict, Field
from retab.types.base import RetabBaseModel
from retab.types.pagination import ListMetadata

# ---------------------------------------------------------------------------
# Enums / literals
# ---------------------------------------------------------------------------

#: Who authored a version, decision, or audit entry. ``kind`` is descriptive
#: data — code never branches on it; a model/agent/human all use the same
#: review surface.
ActorKind = Literal["model", "agent", "human"]

#: Block type the gated run belongs to.
ReviewBlockType = Literal["extract", "classifier", "split", "for_each"]

#: Verdict carried by a :class:`ReviewDecision`.
ReviewVerdict = Literal["approved", "rejected"]

#: Lifecycle status of a decision submission.
SubmissionStatus = Literal["accepted", "already_applied", "conflict"]

#: Whether the gated workflow run was signalled to resume after a decision.
ResumeStatus = Literal["pending", "resumed", "skipped"]

VersionId: TypeAlias = str


# ---------------------------------------------------------------------------
# Building blocks
# ---------------------------------------------------------------------------


class Actor(RetabBaseModel):
    """Identifies who performed a review action.

    ``kind`` distinguishes a model, an agent, or a human — but it is purely
    descriptive. The review loop treats all three identically.
    """

    model_config = ConfigDict(extra="ignore")

    kind: ActorKind = Field(..., description="model / agent / human — descriptive only.")
    id: str = Field(..., description="Stable identifier of the actor.")
    display_name: str = Field(..., description="Human-readable name for UI surfaces.")


class ReviewVersion(RetabBaseModel):
    """One immutable reviewed output snapshot."""

    model_config = ConfigDict(extra="ignore")

    id: VersionId = Field(..., description="Content-addressed version id.")
    review_id: str = Field(..., description="Review this version belongs to.")
    parent_id: VersionId | None = Field(default=None, description="Parent content-hash version id; None for the model output.")
    author: Actor = Field(..., description="Who created this version.")
    snapshot: dict[str, Any] = Field(..., description="The block output payload at this version.")
    note: str | None = Field(default=None, description="Optional free-text note attached to this version.")
    created_at: datetime.datetime = Field(..., description="When this version was created.")


class ReviewDecision(RetabBaseModel):
    """The terminal verdict recorded against one exact output version.

    Plain shape — ``reason`` is required iff ``verdict == "rejected"``, but
    the invariant lives in the two action endpoints (``approve`` / ``reject``)
    rather than in a discriminated union here.
    """

    model_config = ConfigDict(extra="ignore")

    verdict: ReviewVerdict = Field(..., description="approved / rejected.")
    version_id: VersionId = Field(..., description="Content-hash id of the version being decided.")
    author: Actor = Field(..., description="Who made the decision.")
    decided_at: datetime.datetime = Field(..., description="When the decision was recorded.")
    reason: str | None = Field(default=None, description="Free-text reason; required for rejections.")


# ---------------------------------------------------------------------------
# Top-level review shapes
# ---------------------------------------------------------------------------


class Review(RetabBaseModel):
    """The canonical review sidecar for one gated block run."""

    model_config = ConfigDict(extra="ignore")

    id: str = Field(..., description="Review id.")
    workflow_id: str = Field(..., description="Workflow the run belongs to.")
    workflow_version_id: str = Field(..., description="Pinned workflow version id.")
    workflow_run_id: str = Field(..., description="Workflow run id the gated block belongs to.")
    block_id: str = Field(..., description="Gated block id.")
    step_id: str = Field(..., description="Execution step id for the reviewed block.")
    parent_step_id: str | None = Field(default=None, description="Parent step id for child/iteration review contexts.")
    iteration_key: str | None = Field(default=None, description="for_each iteration key when available.")
    block_type: ReviewBlockType = Field(..., description="Type of the gated block.")
    triggered_by: dict[str, Any] = Field(..., description="Discriminated review predicate that opened the gate.")
    created_at: datetime.datetime = Field(..., description="When the review was created.")
    decision: ReviewDecision | None = Field(default=None, description="Terminal decision, if one has been made.")


class ReviewSummary(RetabBaseModel):
    """A lightweight review queue row."""

    model_config = ConfigDict(extra="ignore")

    id: str = Field(..., description="Review id.")
    workflow_id: str = Field(..., description="Workflow the run belongs to.")
    workflow_run_id: str = Field(..., description="Workflow run id the gated block belongs to.")
    block_id: str = Field(..., description="Gated block id.")
    step_id: str = Field(..., description="Execution step id for the reviewed block.")
    parent_step_id: str | None = Field(default=None, description="Parent step id for child/iteration review contexts.")
    iteration_key: str | None = Field(default=None, description="for_each iteration key when available.")
    block_type: ReviewBlockType = Field(..., description="Type of the gated block.")
    triggered_by: dict[str, Any] = Field(..., description="Discriminated review predicate that opened the gate.")
    created_at: datetime.datetime = Field(..., description="When the review was created.")
    seed_version_id: VersionId = Field(..., description="Version id for the seed output.")
    version_count: int = Field(..., description="Number of versions in the review.")
    decision: ReviewDecision | None = Field(default=None, description="Terminal decision, if one has been made.")


class ReviewQueueResponse(RetabBaseModel):
    """Envelope returned by ``reviews.list(...)``.

    Pages are cursored via :class:`ListMetadata`'s ``before`` / ``after`` ids.
    ``has_more`` is a derived convenience: ``True`` iff ``list_metadata.after``
    is not ``None``.
    """

    model_config = ConfigDict(extra="ignore")

    data: list[ReviewSummary] = Field(default_factory=list, description="Page of queue items.")
    list_metadata: ListMetadata = Field(
        ...,
        description="Boundary resource IDs for page navigation (before/after).",
    )

    @property
    def has_more(self) -> bool:
        """Whether there are more pages available after this page's last review id."""
        return self.list_metadata.after is not None


class ReviewVersionListResponse(RetabBaseModel):
    """Envelope returned by ``reviews.versions.list(...)``."""

    model_config = ConfigDict(extra="ignore")

    data: list[ReviewVersion] = Field(default_factory=list, description="Page of review versions.")
    list_metadata: ListMetadata = Field(
        ...,
        description="Boundary resource IDs for page navigation (before/after).",
    )

    @property
    def has_more(self) -> bool:
        """Whether there are more pages available after this page's last version id."""
        return self.list_metadata.after is not None


class SubmitDecisionResponse(RetabBaseModel):
    """Envelope returned by ``reviews.approve/reject(...)``.

    ``submission_status`` is ``accepted`` when the decision is committed.
    """

    model_config = ConfigDict(extra="ignore")

    submission_status: SubmissionStatus = Field(..., description="Decision submission lifecycle status.")
    review: Review = Field(..., description="The review after the decision was applied.")
    resume_status: ResumeStatus = Field(
        default="resumed",
        description="Whether the gated workflow run was signalled to resume. 'pending' means the decision was recorded and resume delivery is queued for retry.",
    )
    resume_error: str | None = Field(
        default=None,
        description="Reason string when resume_status == 'pending'; None otherwise.",
    )


__all__ = [
    "ActorKind",
    "ReviewBlockType",
    "ReviewVerdict",
    "SubmissionStatus",
    "ResumeStatus",
    "VersionId",
    "Actor",
    "ReviewVersion",
    "ReviewDecision",
    "Review",
    "ReviewSummary",
    "ReviewQueueResponse",
    "ReviewVersionListResponse",
    "SubmitDecisionResponse",
]
