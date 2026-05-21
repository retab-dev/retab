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

#: Provenance of an :class:`OutputVersion` snapshot.
OutputVersionOrigin = Literal["model_output", "agent_created", "human_created"]

#: Verdict carried by a :class:`ReviewDecision`.
ReviewVerdict = Literal["approved", "rejected"]

#: Origin accepted when posting a new version through ``reviews.create_version(...)``.
VersionOrigin = Literal["human_created", "agent_created"]

#: Lifecycle status of a decision submission.
#: ``accepted`` — fresh decision recorded.
#: ``already_applied`` — server detected an idempotent replay.
#: ``accepted_pending_resume`` — decision committed but the temporal resume
#: signal failed; the reconcile loop will retry the resume.
SubmissionStatus = Literal["accepted", "already_applied", "accepted_pending_resume"]

#: Whether the gated workflow run was signalled to resume after a decision.
ResumeStatus = Literal["resumed", "skipped", "failed"]

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


class OutputVersion(RetabBaseModel):
    """One immutable reviewed output snapshot."""

    model_config = ConfigDict(extra="ignore")

    parent_id: VersionId | None = Field(default=None, description="Parent content-hash version id; None for the model output.")
    author: Actor = Field(..., description="Who created this version.")
    origin: OutputVersionOrigin = Field(..., description="How this snapshot was produced.")
    snapshot: dict[str, Any] = Field(..., description="The block output payload at this version.")
    note: str | None = Field(default=None, description="Optional free-text note attached to this version.")
    created_at: datetime.datetime = Field(..., description="When this version was created.")


class ReviewDecision(RetabBaseModel):
    """The terminal verdict recorded against one exact output version."""

    model_config = ConfigDict(extra="ignore")

    verdict: ReviewVerdict = Field(..., description="approved / rejected.")
    version_id: VersionId = Field(..., description="Content-hash id of the version being decided.")
    decided_by: Actor = Field(..., description="Who made the decision.")
    decided_at: datetime.datetime = Field(..., description="When the decision was recorded.")
    reason: str | None = Field(default=None, description="Free-text reason; required for rejections.")


# ---------------------------------------------------------------------------
# Top-level review shapes
# ---------------------------------------------------------------------------


class ReviewOverlay(RetabBaseModel):
    """The canonical review sidecar for one gated block run."""

    model_config = ConfigDict(extra="ignore", populate_by_name=True)

    id: str = Field(..., alias="_id", description="Overlay document id.")
    organization_id: str | None = Field(default=None, description="Organization that owns the run. Not present in public API responses; kept for internal echoes only.")
    workflow_id: str = Field(..., description="Workflow the run belongs to.")
    workflow_version_id: str = Field(..., description="Pinned workflow version id.")
    workflow_run_id: str = Field(..., description="Workflow run id the gated block belongs to.")
    block_id: str = Field(..., description="Gated block id.")
    block_run_id: str = Field(..., description="Block run id.")
    runtime_block_id: str | None = Field(
        default=None, description="Runtime block id, when different from block_id. Not present in public API responses; kept for internal echoes only."
    )
    block_type: ReviewBlockType = Field(..., description="Type of the gated block.")
    triggered_by: dict[str, Any] = Field(..., description="Discriminated review predicate that opened the gate.")
    awaiting_since: datetime.datetime = Field(..., description="When the review started awaiting a decision.")
    priority: int = Field(..., description="Queue priority; higher is more urgent.")
    versions_by_id: dict[VersionId, OutputVersion] = Field(..., description="Output versions keyed by content-hash id.")
    decision: ReviewDecision | None = Field(default=None, description="Terminal decision, if one has been made.")


class ReviewQueueItem(RetabBaseModel):
    """A lightweight review queue row."""

    model_config = ConfigDict(extra="ignore", populate_by_name=True)

    id: str = Field(..., alias="_id", description="Overlay document id.")
    organization_id: str | None = Field(default=None, description="Organization that owns the run. Not present in public API responses; kept for internal echoes only.")
    workflow_id: str = Field(..., description="Workflow the run belongs to.")
    workflow_version_id: str = Field(..., description="Pinned workflow version id.")
    workflow_run_id: str = Field(..., description="Workflow run id the gated block belongs to.")
    block_id: str = Field(..., description="Gated block id.")
    block_run_id: str = Field(..., description="Block run id.")
    block_type: ReviewBlockType = Field(..., description="Type of the gated block.")
    triggered_by: dict[str, Any] = Field(..., description="Discriminated review predicate that opened the gate.")
    awaiting_since: datetime.datetime = Field(..., description="When the review started awaiting a decision.")
    priority: int = Field(..., description="Queue priority; higher is more urgent.")


class ReviewQueueResponse(RetabBaseModel):
    """Envelope returned by ``reviews.list(...)``.

    Pages are cursored via :class:`ListMetadata`'s ``before`` / ``after`` ids.
    ``has_more`` is a derived convenience: ``True`` iff ``list_metadata.after``
    is not ``None``.
    """

    model_config = ConfigDict(extra="ignore")

    data: list[ReviewQueueItem] = Field(default_factory=list, description="Page of queue items.")
    list_metadata: ListMetadata = Field(
        ...,
        description="Boundary resource IDs for page navigation (before/after).",
    )

    @property
    def has_more(self) -> bool:
        """Whether there are more pages available after this page's last review id."""
        return self.list_metadata.after is not None


class SubmitDecisionResponse(RetabBaseModel):
    """Envelope returned by ``reviews.approve/reject(...)``.

    ``submission_status`` is ``accepted`` for a fresh decision, or one of the
    idempotent ``already_*`` values when the server detects a replay.
    """

    model_config = ConfigDict(extra="ignore")

    submission_status: SubmissionStatus = Field(..., description="Decision submission lifecycle status.")
    review: ReviewOverlay = Field(..., description="The review after the decision was applied.")
    resume_status: ResumeStatus = Field(
        default="resumed",
        description="Whether the gated workflow run was signalled to resume. 'failed' means the decision was recorded but the temporal signal could not be sent; the reconcile loop will retry.",
    )
    resume_error: str | None = Field(
        default=None,
        description="Reason string when resume_status == 'failed'; None otherwise.",
    )


__all__ = [
    "ActorKind",
    "ReviewBlockType",
    "OutputVersionOrigin",
    "ReviewVerdict",
    "VersionOrigin",
    "SubmissionStatus",
    "ResumeStatus",
    "VersionId",
    "Actor",
    "OutputVersion",
    "ReviewDecision",
    "ReviewOverlay",
    "ReviewQueueItem",
    "ReviewQueueResponse",
    "SubmitDecisionResponse",
]
