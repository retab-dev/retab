"""Pydantic models for the workflow HIL review overlay.

Mirrors the backend "HIL review overlay" — a versioned sidecar attached to a
gated workflow block run, served under ``/workflows/reviews``. The SDK drives
the review loop through ``client.workflows.runs.reviews.*``.

Actor-neutral by construction: a proposal authored by a model, an agent, or a
human all flow through the SAME types. ``Actor.kind`` is data, never branched
on — there is no ``AgentProposedDecision`` / ``agent_proposal`` here, only the
neutral :class:`OutputVersion` / :class:`ReviewDecision` shapes.
"""

from __future__ import annotations

import datetime
from typing import Any, Literal

from pydantic import ConfigDict, Field
from retab.types.base import RetabBaseModel

# ---------------------------------------------------------------------------
# Enums / literals
# ---------------------------------------------------------------------------

#: Who authored a version, decision, or audit entry. ``kind`` is descriptive
#: data — code never branches on it; a model/agent/human all use the same
#: review surface.
ActorKind = Literal["model", "agent", "human"]

#: Block type the gated run belongs to.
ReviewBlockType = Literal["extract", "classifier", "split", "conditional"]

#: Overlay lifecycle status.
ReviewStatus = Literal["awaiting_review", "approved", "rejected"]

#: Provenance of an :class:`OutputVersion` snapshot.
OutputVersionOrigin = Literal["model_output", "agent_edit", "human_edit", "revert"]

#: Verdict carried by a :class:`ReviewDecision`.
ReviewVerdict = Literal["approved", "rejected", "escalated"]

#: Origin accepted when posting a new version through ``reviews.edit(...)``.
EditOrigin = Literal["human_edit", "agent_edit"]

#: Result of submitting a decision; ``already_*`` values are idempotent replays.
SubmissionStatus = Literal["accepted", "already_received", "already_applied"]


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
    """One immutable snapshot of the block output in the version history.

    Versions form a tree: ``parent_seq`` points at the version this one was
    derived from (``None`` for the root model output). ``origin`` records how
    the snapshot came to be, and ``content_sha256`` is the integrity hash.
    """

    model_config = ConfigDict(extra="ignore")

    seq: int = Field(..., description="Monotonic sequence number of this version.")
    parent_seq: int | None = Field(default=None, description="Sequence of the parent version; None for the root.")
    author: Actor = Field(..., description="Who created this version.")
    origin: OutputVersionOrigin = Field(..., description="How this snapshot was produced.")
    snapshot: dict[str, Any] = Field(..., description="The block output payload at this version.")
    content_sha256: str = Field(..., description="Integrity hash of the snapshot.")
    note: str | None = Field(default=None, description="Optional free-text note attached to this version.")
    created_at: datetime.datetime = Field(..., description="When this version was created.")


class ReviewDecision(RetabBaseModel):
    """A verdict recorded against a specific output version.

    ``on_seq`` is the version the decision was made against; ``effective_seq``
    is the version that becomes effective when the verdict is applied (it may
    differ when an edited output is supplied). ``supersedes_decision_id`` links
    a corrected decision back to the one it replaces.
    """

    model_config = ConfigDict(extra="ignore")

    decision_id: str = Field(..., description="Stable identifier of this decision.")
    verdict: ReviewVerdict = Field(..., description="approved / rejected / escalated.")
    decided_by: Actor = Field(..., description="Who made the decision.")
    decided_at: datetime.datetime = Field(..., description="When the decision was recorded.")
    on_seq: int = Field(..., description="Version sequence the decision was made against.")
    effective_seq: int | None = Field(default=None, description="Version that becomes effective on apply.")
    reason: str | None = Field(default=None, description="Free-text reason; required for rejections.")
    escalate_to: str | None = Field(default=None, description="Escalation target when verdict is 'escalated'.")
    supersedes_decision_id: str | None = Field(default=None, description="Decision id this one corrects, if any.")


class AuditEntry(RetabBaseModel):
    """One row of the overlay's append-only audit log.

    Each entry records the CAS token observed (``rev_observed``) and the
    resulting token (``rev_result``) so the full history of mutations is
    reconstructable.
    """

    model_config = ConfigDict(extra="ignore")

    entry_id: str = Field(..., description="Stable identifier of this audit entry.")
    action: str = Field(..., description="What action was performed.")
    actor: Actor = Field(..., description="Who performed the action.")
    at: datetime.datetime = Field(..., description="When the action occurred.")
    rev_observed: int = Field(..., description="CAS token observed before the action.")
    rev_result: int | None = Field(default=None, description="CAS token after the action; None if it did not advance.")
    detail: dict[str, Any] = Field(default_factory=dict, description="Action-specific detail payload.")


class ReviewClaim(RetabBaseModel):
    """A soft lock held by an actor while it works on the review.

    A claim is advisory — it prevents accidental concurrent edits, and it
    expires at ``expires_at`` so a crashed reviewer never blocks the queue.
    """

    model_config = ConfigDict(extra="ignore")

    holder: Actor = Field(..., description="Actor currently holding the claim.")
    claimed_at: datetime.datetime = Field(..., description="When the claim was taken.")
    expires_at: datetime.datetime = Field(..., description="When the claim lapses.")


# ---------------------------------------------------------------------------
# Top-level overlay shapes
# ---------------------------------------------------------------------------


class ReviewOverlay(RetabBaseModel):
    """The full versioned review sidecar for one gated block run.

    ``rev`` is the compare-and-swap token: every mutating call carries the
    last-observed ``rev`` as ``version_stamp``, and the server returns HTTP 409
    if it has since advanced. :attr:`version_stamp` exposes ``rev`` under the
    name the review loop speaks.
    """

    model_config = ConfigDict(extra="ignore", populate_by_name=True)

    id: str = Field(..., alias="_id", description="Overlay document id.")
    organization_id: str = Field(..., description="Organization that owns the run.")
    workflow_id: str = Field(..., description="Workflow the run belongs to.")
    workflow_version_id: str = Field(..., description="Pinned workflow version id.")
    workflow_run_id: str = Field(..., description="Workflow run id the gated block belongs to.")
    block_id: str = Field(..., description="Gated block id.")
    block_run_id: str = Field(..., description="Block run id.")
    block_type: ReviewBlockType = Field(..., description="Type of the gated block.")
    triggered_by: dict[str, Any] = Field(..., description="Discriminated HIL predicate that opened the gate.")
    status: ReviewStatus = Field(..., description="Overlay lifecycle status.")
    awaiting_since: datetime.datetime = Field(..., description="When the overlay started awaiting review.")
    decided_at: datetime.datetime | None = Field(default=None, description="When a terminal decision landed.")
    priority: int = Field(..., description="Queue priority; higher is more urgent.")
    rev: int = Field(..., description="Compare-and-swap token == version_stamp.")
    claim: ReviewClaim | None = Field(default=None, description="Active soft lock, if any.")
    versions: list[OutputVersion] = Field(default_factory=list, description="Full version history.")
    decisions: list[ReviewDecision] = Field(default_factory=list, description="All decisions recorded.")
    audit: list[AuditEntry] = Field(default_factory=list, description="Append-only audit log.")
    head_seq: int = Field(..., description="Sequence of the latest version.")
    effective_seq: int | None = Field(default=None, description="Sequence of the effective version, if decided.")

    @property
    def version_stamp(self) -> int:
        """The CAS token to pass back on the next mutating call (alias of ``rev``)."""
        return self.rev


class ReviewQueueItem(RetabBaseModel):
    """A lightweight overlay summary returned by ``reviews.list(...)``.

    Identical to :class:`ReviewOverlay` minus the heavy ``versions`` /
    ``decisions`` / ``audit`` arrays — enough to render a queue row and then
    fetch the full overlay with ``reviews.get(...)``.
    """

    model_config = ConfigDict(extra="ignore", populate_by_name=True)

    id: str = Field(..., alias="_id", description="Overlay document id.")
    organization_id: str = Field(..., description="Organization that owns the run.")
    workflow_id: str = Field(..., description="Workflow the run belongs to.")
    workflow_version_id: str = Field(..., description="Pinned workflow version id.")
    workflow_run_id: str = Field(..., description="Workflow run id the gated block belongs to.")
    block_id: str = Field(..., description="Gated block id.")
    block_run_id: str = Field(..., description="Block run id.")
    block_type: ReviewBlockType = Field(..., description="Type of the gated block.")
    triggered_by: dict[str, Any] = Field(..., description="Discriminated HIL predicate that opened the gate.")
    status: ReviewStatus = Field(..., description="Overlay lifecycle status.")
    awaiting_since: datetime.datetime = Field(..., description="When the overlay started awaiting review.")
    decided_at: datetime.datetime | None = Field(default=None, description="When a terminal decision landed.")
    priority: int = Field(..., description="Queue priority; higher is more urgent.")
    rev: int = Field(..., description="Compare-and-swap token == version_stamp.")
    claim: ReviewClaim | None = Field(default=None, description="Active soft lock, if any.")
    head_seq: int = Field(..., description="Sequence of the latest version.")
    effective_seq: int | None = Field(default=None, description="Sequence of the effective version, if decided.")

    @property
    def version_stamp(self) -> int:
        """The CAS token to pass back on the next mutating call (alias of ``rev``)."""
        return self.rev


class ReviewQueueResponse(RetabBaseModel):
    """Envelope returned by ``reviews.list(...)``."""

    model_config = ConfigDict(extra="ignore")

    data: list[ReviewQueueItem] = Field(default_factory=list, description="Page of queue items.")
    has_more: bool = Field(..., description="Whether more items exist past this page.")


class SubmitDecisionResponse(RetabBaseModel):
    """Envelope returned by ``reviews.approve/reject/escalate(...)``.

    ``submission_status`` is ``accepted`` for a fresh decision, or one of the
    idempotent ``already_*`` values when a ``command_id`` replay is detected.
    """

    model_config = ConfigDict(extra="ignore")

    submission_status: SubmissionStatus = Field(..., description="Decision submission lifecycle status.")
    overlay: ReviewOverlay = Field(..., description="The overlay after the decision was applied.")


__all__ = [
    "ActorKind",
    "ReviewBlockType",
    "ReviewStatus",
    "OutputVersionOrigin",
    "ReviewVerdict",
    "EditOrigin",
    "SubmissionStatus",
    "Actor",
    "OutputVersion",
    "ReviewDecision",
    "AuditEntry",
    "ReviewClaim",
    "ReviewOverlay",
    "ReviewQueueItem",
    "ReviewQueueResponse",
    "SubmitDecisionResponse",
]
