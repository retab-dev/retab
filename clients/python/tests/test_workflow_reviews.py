"""Smoke tests for `client.workflows.runs.reviews.*` — types + request builders.

Mirrors `test_workflow_steps.py`: mock `client._prepared_request` so we can
assert on the constructed `PreparedRequest` without hitting the API, and
round-trip the documented JSON shapes through the new Pydantic types.

These tests do NOT need a live server.
"""

from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock

import pytest

from retab.exceptions import NotFoundError
from retab.resources.workflows.runs.reviews.client import (
    AsyncWorkflowReviews,
    WorkflowReviews,
)
from retab.types.workflows import (
    Actor,
    AuditEntry,
    OutputVersion,
    ReviewDecision,
    ReviewOverlay,
    ReviewQueueItem,
    ReviewQueueResponse,
    SubmitDecisionResponse,
)

_NOW = "2026-05-01T14:30:00Z"

_ACTOR_HUMAN = {"kind": "human", "id": "user_1", "display_name": "Ada"}
_ACTOR_MODEL = {"kind": "model", "id": "gpt", "display_name": "Extractor"}
_ACTOR_AGENT = {"kind": "agent", "id": "agent_1", "display_name": "Reviewer Agent"}

_OUTPUT_VERSION = {
    "seq": 1,
    "parent_seq": None,
    "author": _ACTOR_MODEL,
    "origin": "model_output",
    "snapshot": {"total": 100},
    "content_sha256": "abc123",
    "note": None,
    "created_at": _NOW,
}

_DECISION = {
    "decision_id": "dec_1",
    "verdict": "approved",
    "decided_by": _ACTOR_HUMAN,
    "decided_at": _NOW,
    "on_seq": 1,
    "effective_seq": 1,
    "reason": None,
    "escalate_to": None,
    "supersedes_decision_id": None,
}

_AUDIT_ENTRY = {
    "entry_id": "audit_1",
    "action": "decision_submitted",
    "actor": _ACTOR_HUMAN,
    "at": _NOW,
    "rev_observed": 3,
    "rev_result": 4,
    "detail": {"verdict": "approved"},
}

_OVERLAY = {
    "_id": "ovl_abc",
    "organization_id": "org_x",
    "workflow_id": "wf_1",
    "workflow_version_id": "wv_1",
    "workflow_run_id": "run_1",
    "block_id": "extract-1",
    "block_run_id": "brun_1",
    "block_type": "extract",
    "triggered_by": {"kind": "low_confidence", "threshold": 0.8},
    "status": "awaiting_review",
    "awaiting_since": _NOW,
    "decided_at": None,
    "priority": 5,
    "rev": 3,
    "claim": {"holder": _ACTOR_HUMAN, "claimed_at": _NOW, "expires_at": _NOW},
    "versions": [_OUTPUT_VERSION],
    "decisions": [_DECISION],
    "audit": [_AUDIT_ENTRY],
    "head_seq": 1,
    "effective_seq": None,
}

_QUEUE_ITEM = {
    "_id": "ovl_abc",
    "organization_id": "org_x",
    "workflow_id": "wf_1",
    "workflow_version_id": "wv_1",
    "workflow_run_id": "run_1",
    "block_id": "extract-1",
    "block_run_id": "brun_1",
    "block_type": "extract",
    "triggered_by": {"kind": "low_confidence"},
    "status": "awaiting_review",
    "awaiting_since": _NOW,
    "decided_at": None,
    "priority": 5,
    "rev": 3,
    "claim": None,
    "head_seq": 1,
    "effective_seq": None,
}


# ---------------------------------------------------------------------------
# Type round-tripping
# ---------------------------------------------------------------------------


def test_review_overlay_round_trips() -> None:
    overlay = ReviewOverlay.model_validate(_OVERLAY)
    assert overlay.id == "ovl_abc"
    assert overlay.rev == 3
    # version_stamp is a read-only convenience alias of rev
    assert overlay.version_stamp == 3
    assert overlay.status == "awaiting_review"
    assert len(overlay.versions) == 1
    assert len(overlay.decisions) == 1
    assert len(overlay.audit) == 1
    assert overlay.claim is not None
    assert overlay.claim.holder.kind == "human"
    # dumping by alias yields _id again
    dumped = overlay.model_dump(by_alias=True)
    assert dumped["_id"] == "ovl_abc"
    assert ReviewOverlay.model_validate(dumped).rev == 3


def test_review_queue_item_round_trips_without_heavy_arrays() -> None:
    item = ReviewQueueItem.model_validate(_QUEUE_ITEM)
    assert item.id == "ovl_abc"
    assert item.version_stamp == 3
    # queue items intentionally carry no version/decision/audit history
    assert "versions" not in item.model_dump()
    assert "decisions" not in item.model_dump()
    assert "audit" not in item.model_dump()


def test_review_queue_response_round_trips() -> None:
    resp = ReviewQueueResponse.model_validate({"data": [_QUEUE_ITEM], "has_more": True})
    assert resp.has_more is True
    assert len(resp.data) == 1
    assert resp.data[0].block_id == "extract-1"


def test_submit_decision_response_round_trips() -> None:
    resp = SubmitDecisionResponse.model_validate(
        {"submission_status": "accepted", "overlay": _OVERLAY}
    )
    assert resp.submission_status == "accepted"
    assert resp.overlay.rev == 3


def test_actor_kind_is_neutral_data_not_a_branch() -> None:
    # model / agent / human all parse into the SAME Actor type.
    for raw in (_ACTOR_MODEL, _ACTOR_AGENT, _ACTOR_HUMAN):
        actor = Actor.model_validate(raw)
        assert actor.kind in {"model", "agent", "human"}
        assert actor.id and actor.display_name


def test_output_version_and_decision_and_audit_round_trip() -> None:
    version = OutputVersion.model_validate(_OUTPUT_VERSION)
    assert version.seq == 1
    assert version.parent_seq is None
    assert version.snapshot == {"total": 100}

    decision = ReviewDecision.model_validate(_DECISION)
    assert decision.verdict == "approved"
    assert decision.on_seq == 1

    entry = AuditEntry.model_validate(_AUDIT_ENTRY)
    assert entry.rev_observed == 3
    assert entry.rev_result == 4


# ---------------------------------------------------------------------------
# prepare_* request builders
# ---------------------------------------------------------------------------


def test_prepare_list_builds_get_with_params() -> None:
    reviews = WorkflowReviews(client=MagicMock())
    request = reviews.prepare_list(workflow_id="wf_1", status="approved", mine=True, limit=10)
    assert request.method == "GET"
    assert request.url == "/workflows/reviews"
    assert request.params == {"status": "approved", "mine": True, "limit": 10, "workflow_id": "wf_1"}


def test_prepare_list_omits_workflow_id_when_absent() -> None:
    reviews = WorkflowReviews(client=MagicMock())
    request = reviews.prepare_list()
    assert request.params == {"status": "awaiting_review", "mine": False, "limit": 50}


def test_prepare_get_builds_run_block_scoped_url() -> None:
    reviews = WorkflowReviews(client=MagicMock())
    request = reviews.prepare_get("run_1", "extract-1")
    assert request.method == "GET"
    assert request.url == "/workflows/reviews/run_1/extract-1"


def test_prepare_edit_posts_to_versions() -> None:
    reviews = WorkflowReviews(client=MagicMock())
    request = reviews.prepare_edit(
        "run_1",
        "extract-1",
        snapshot={"total": 200},
        version_stamp=3,
        origin="agent_edit",
        note="bumped total",
        command_id="cmd_1",
    )
    assert request.method == "POST"
    assert request.url == "/workflows/reviews/run_1/extract-1/versions"
    assert request.data == {
        "snapshot": {"total": 200},
        "version_stamp": 3,
        "origin": "agent_edit",
        "note": "bumped total",
        "command_id": "cmd_1",
    }


def test_prepare_decision_posts_verdict() -> None:
    reviews = WorkflowReviews(client=MagicMock())
    request = reviews.prepare_decision(
        "run_1",
        "extract-1",
        verdict="rejected",
        version_stamp=3,
        reason="wrong vendor",
    )
    assert request.method == "POST"
    assert request.url == "/workflows/reviews/run_1/extract-1/decision"
    assert request.data["verdict"] == "rejected"
    assert request.data["version_stamp"] == 3
    assert request.data["reason"] == "wrong vendor"


def test_prepare_claim_and_release() -> None:
    reviews = WorkflowReviews(client=MagicMock())
    claim = reviews.prepare_claim("run_1", "extract-1", version_stamp=3, ttl_seconds=600)
    assert claim.method == "POST"
    assert claim.url == "/workflows/reviews/run_1/extract-1/claim"
    assert claim.data == {"version_stamp": 3, "ttl_seconds": 600}

    release = reviews.prepare_release("run_1", "extract-1", version_stamp=3)
    assert release.method == "POST"
    assert release.url == "/workflows/reviews/run_1/extract-1/release"
    assert release.data == {"version_stamp": 3}


# ---------------------------------------------------------------------------
# Sync resource — full round trip through a mocked client
# ---------------------------------------------------------------------------


def test_list_parses_queue_response() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {"data": [_QUEUE_ITEM], "has_more": False}

    result = WorkflowReviews(client=client).list(workflow_id="wf_1")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/reviews"
    assert isinstance(result, ReviewQueueResponse)
    assert result.data[0].block_id == "extract-1"


def test_get_parses_overlay() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _OVERLAY

    overlay = WorkflowReviews(client=client).get("run_1", "extract-1")

    request = client._prepared_request.call_args.args[0]
    assert request.url == "/workflows/reviews/run_1/extract-1"
    assert overlay.version_stamp == 3


def test_approve_sends_approved_verdict() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {"submission_status": "accepted", "overlay": _OVERLAY}

    resp = WorkflowReviews(client=client).approve(
        "run_1", "extract-1", version_stamp=3, edited_output={"total": 150}
    )

    request = client._prepared_request.call_args.args[0]
    assert request.url == "/workflows/reviews/run_1/extract-1/decision"
    assert request.data["verdict"] == "approved"
    assert request.data["edited_output"] == {"total": 150}
    assert resp.submission_status == "accepted"


def test_reject_requires_reason_and_sends_rejected_verdict() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {"submission_status": "accepted", "overlay": _OVERLAY}

    WorkflowReviews(client=client).reject("run_1", "extract-1", version_stamp=3, reason="bad data")

    request = client._prepared_request.call_args.args[0]
    assert request.data["verdict"] == "rejected"
    assert request.data["reason"] == "bad data"


def test_escalate_sends_escalated_verdict_and_target() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {"submission_status": "accepted", "overlay": _OVERLAY}

    WorkflowReviews(client=client).escalate(
        "run_1", "extract-1", version_stamp=3, reason="unsure", escalate_to="senior_team"
    )

    request = client._prepared_request.call_args.args[0]
    assert request.data["verdict"] == "escalated"
    assert request.data["escalate_to"] == "senior_team"


def test_edit_posts_version_and_returns_overlay() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _OVERLAY

    overlay = WorkflowReviews(client=client).edit(
        "run_1", "extract-1", snapshot={"total": 200}, version_stamp=3
    )

    request = client._prepared_request.call_args.args[0]
    assert request.url == "/workflows/reviews/run_1/extract-1/versions"
    assert request.data["origin"] == "human_edit"
    assert isinstance(overlay, ReviewOverlay)


def test_reject_requires_reason_keyword() -> None:
    reviews = WorkflowReviews(client=MagicMock())
    with pytest.raises(TypeError):
        reviews.reject("run_1", "extract-1", version_stamp=3)  # type: ignore[call-arg]


# ---------------------------------------------------------------------------
# wait_for — polling semantics
# ---------------------------------------------------------------------------


def test_wait_for_returns_when_awaiting_review() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _OVERLAY

    overlay = WorkflowReviews(client=client).wait_for(
        "run_1", "extract-1", timeout=5.0, poll_interval=0.01
    )
    assert overlay.status == "awaiting_review"


def test_wait_for_times_out_when_overlay_never_appears() -> None:
    client = MagicMock()
    client._prepared_request.side_effect = NotFoundError("nope", status_code=404)

    with pytest.raises(TimeoutError):
        WorkflowReviews(client=client).wait_for(
            "run_1", "extract-1", timeout=0.05, poll_interval=0.01
        )


# ---------------------------------------------------------------------------
# Async resource
# ---------------------------------------------------------------------------


@pytest.mark.asyncio
async def test_async_get_parses_overlay() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(return_value=_OVERLAY)

    overlay = await AsyncWorkflowReviews(client=client).get("run_1", "extract-1")

    request = client._prepared_request.call_args.args[0]
    assert request.url == "/workflows/reviews/run_1/extract-1"
    assert overlay.version_stamp == 3


@pytest.mark.asyncio
async def test_async_approve_sends_approved_verdict() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(
        return_value={"submission_status": "already_applied", "overlay": _OVERLAY}
    )

    resp = await AsyncWorkflowReviews(client=client).approve(
        "run_1", "extract-1", version_stamp=3
    )

    request = client._prepared_request.call_args.args[0]
    assert request.data["verdict"] == "approved"
    assert resp.submission_status == "already_applied"


@pytest.mark.asyncio
async def test_async_wait_for_times_out() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(side_effect=NotFoundError("nope", status_code=404))

    with pytest.raises(TimeoutError):
        await AsyncWorkflowReviews(client=client).wait_for(
            "run_1", "extract-1", timeout=0.05, poll_interval=0.01
        )
