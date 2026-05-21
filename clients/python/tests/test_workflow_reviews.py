"""Smoke tests for `client.workflows.reviews.*`."""

from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock

import pytest

from retab.exceptions import NotFoundError
from retab.resources.workflows.reviews.client import AsyncWorkflowReviews, WorkflowReviews
from retab.types.workflows import (
    Actor,
    OutputVersion,
    ReviewDecision,
    ReviewOverlay,
    ReviewQueueItem,
    ReviewQueueResponse,
    SubmitDecisionResponse,
)

_NOW = "2026-05-01T14:30:00Z"
_VERSION_ID = "a" * 64
_CHILD_VERSION_ID = "b" * 64

_ACTOR_HUMAN = {"kind": "human", "id": "user_1", "display_name": "Ada"}
_ACTOR_MODEL = {"kind": "model", "id": "gpt", "display_name": "Extractor"}
_ACTOR_AGENT = {"kind": "agent", "id": "agent_1", "display_name": "Reviewer Agent"}

_OUTPUT_VERSION = {
    "parent_id": None,
    "author": _ACTOR_MODEL,
    "origin": "model_output",
    "snapshot": {"output": {"total": 100, "currency": "USD"}},
    "note": None,
    "created_at": _NOW,
}

_DECISION = {
    "verdict": "approved",
    "version_id": _VERSION_ID,
    "decided_by": _ACTOR_HUMAN,
    "decided_at": _NOW,
    "reason": None,
}

_OVERLAY = {
    "_id": "brun_1",
    "organization_id": "org_x",
    "workflow_id": "wf_1",
    "workflow_version_id": "wv_1",
    "workflow_run_id": "run_1",
    "block_id": "extract-1",
    "block_run_id": "brun_1",
    "runtime_block_id": None,
    "block_type": "extract",
    "triggered_by": {"kind": "low_confidence", "threshold": 0.8},
    "awaiting_since": _NOW,
    "priority": 5,
    "versions_by_id": {
        _VERSION_ID: _OUTPUT_VERSION,
        _CHILD_VERSION_ID: {
            **_OUTPUT_VERSION,
            "parent_id": _VERSION_ID,
            "author": _ACTOR_AGENT,
            "origin": "agent_created",
            "snapshot": {"output": {"total": 150, "currency": "USD"}},
        },
    },
    "decision": _DECISION,
}

_QUEUE_ITEM = {
    "_id": "brun_1",
    "organization_id": "org_x",
    "workflow_id": "wf_1",
    "workflow_version_id": "wv_1",
    "workflow_run_id": "run_1",
    "block_id": "extract-1",
    "block_run_id": "brun_1",
    "block_type": "extract",
    "triggered_by": {"kind": "low_confidence"},
    "awaiting_since": _NOW,
    "priority": 5,
}

# Real backend responses strip organization_id and runtime_block_id from public
# review payloads (see overlay_api_models._strip_public_overlay_fields).
_PUBLIC_OVERLAY = {
    "_id": "brun_1",
    "workflow_id": "wf_1",
    "workflow_version_id": "wv_1",
    "workflow_run_id": "run_1",
    "block_id": "extract-1",
    "block_run_id": "brun_1",
    "block_type": "extract",
    "triggered_by": {"kind": "low_confidence", "threshold": 0.8},
    "awaiting_since": _NOW,
    "priority": 5,
    "versions_by_id": {
        _VERSION_ID: _OUTPUT_VERSION,
        _CHILD_VERSION_ID: {
            **_OUTPUT_VERSION,
            "parent_id": _VERSION_ID,
            "author": _ACTOR_AGENT,
            "origin": "agent_created",
            "snapshot": {"output": {"total": 150, "currency": "USD"}},
        },
    },
    "decision": _DECISION,
}

_PUBLIC_QUEUE_ITEM = {
    "_id": "brun_1",
    "workflow_id": "wf_1",
    "workflow_version_id": "wv_1",
    "workflow_run_id": "run_1",
    "block_id": "extract-1",
    "block_run_id": "brun_1",
    "block_type": "extract",
    "triggered_by": {"kind": "low_confidence"},
    "awaiting_since": _NOW,
    "priority": 5,
}


def test_review_overlay_round_trips() -> None:
    overlay = ReviewOverlay.model_validate(_OVERLAY)
    assert overlay.id == "brun_1"
    assert set(overlay.versions_by_id) == {_VERSION_ID, _CHILD_VERSION_ID}
    assert overlay.versions_by_id[_CHILD_VERSION_ID].parent_id == _VERSION_ID
    assert overlay.decision is not None
    assert overlay.decision.version_id == _VERSION_ID
    dumped = overlay.model_dump(by_alias=True)
    assert dumped["_id"] == "brun_1"
    assert ReviewOverlay.model_validate(dumped).decision is not None


def test_review_queue_item_round_trips_without_overlay_history() -> None:
    item = ReviewQueueItem.model_validate(_QUEUE_ITEM)
    assert item.id == "brun_1"
    dumped = item.model_dump()
    assert "versions_by_id" not in dumped
    assert "decision" not in dumped


def test_review_queue_response_round_trips() -> None:
    resp = ReviewQueueResponse.model_validate(
        {
            "data": [_QUEUE_ITEM],
            "list_metadata": {"before": None, "after": "brun_xxx"},
        }
    )
    assert resp.list_metadata.before is None
    assert resp.list_metadata.after == "brun_xxx"
    assert resp.has_more is True
    assert resp.data[0].block_id == "extract-1"


def test_submit_decision_response_round_trips() -> None:
    resp = SubmitDecisionResponse.model_validate({"submission_status": "accepted", "review": _OVERLAY})
    assert resp.submission_status == "accepted"
    assert resp.review.decision is not None


def test_public_overlay_parses_without_organization_id() -> None:
    overlay = ReviewOverlay.model_validate(_PUBLIC_OVERLAY)
    assert overlay.id == "brun_1"
    assert overlay.organization_id is None
    assert overlay.runtime_block_id is None
    assert overlay.decision is not None
    assert overlay.decision.version_id == _VERSION_ID


def test_public_queue_item_parses_without_organization_id() -> None:
    item = ReviewQueueItem.model_validate(_PUBLIC_QUEUE_ITEM)
    assert item.id == "brun_1"
    assert item.organization_id is None


def test_public_queue_response_round_trips() -> None:
    resp = ReviewQueueResponse.model_validate(
        {
            "data": [_PUBLIC_QUEUE_ITEM],
            "list_metadata": {"before": None, "after": None},
        }
    )
    assert resp.list_metadata.after is None
    assert resp.has_more is False
    assert resp.data[0].organization_id is None


def test_submit_decision_response_defaults_resume_fields() -> None:
    resp = SubmitDecisionResponse.model_validate({"submission_status": "accepted", "review": _PUBLIC_OVERLAY})
    assert resp.resume_status == "resumed"
    assert resp.resume_error is None


def test_submit_decision_response_accepts_pending_resume_with_error() -> None:
    resp = SubmitDecisionResponse.model_validate(
        {
            "submission_status": "accepted_pending_resume",
            "review": _PUBLIC_OVERLAY,
            "resume_status": "failed",
            "resume_error": "Workflow run not found",
        }
    )
    assert resp.submission_status == "accepted_pending_resume"
    assert resp.resume_status == "failed"
    assert resp.resume_error == "Workflow run not found"


def test_submit_decision_response_ignores_unknown_fields() -> None:
    resp = SubmitDecisionResponse.model_validate(
        {
            "submission_status": "accepted",
            "review": _PUBLIC_OVERLAY,
            "future_server_field": "anything",
        }
    )
    assert resp.submission_status == "accepted"
    assert not hasattr(resp, "future_server_field")


def test_actor_kind_is_neutral_data_not_a_branch() -> None:
    for raw in (_ACTOR_MODEL, _ACTOR_AGENT, _ACTOR_HUMAN):
        actor = Actor.model_validate(raw)
        assert actor.kind in {"model", "agent", "human"}


def test_output_version_and_decision_round_trip() -> None:
    version = OutputVersion.model_validate(_OUTPUT_VERSION)
    assert version.parent_id is None
    assert version.snapshot == {"output": {"total": 100, "currency": "USD"}}

    decision = ReviewDecision.model_validate(_DECISION)
    assert decision.verdict == "approved"
    assert decision.version_id == _VERSION_ID


def test_prepare_list_builds_get_with_params() -> None:
    reviews = WorkflowReviews(client=MagicMock())
    request = reviews.prepare_list(workflow_id="wf_1", limit=10)
    assert request.method == "GET"
    assert request.url == "/workflows/reviews"
    assert request.params == {"limit": 10, "workflow_id": "wf_1", "decision": "none"}


def test_prepare_list_can_include_decided_reviews() -> None:
    reviews = WorkflowReviews(client=MagicMock())
    request = reviews.prepare_list(decision="any")
    assert request.params == {"limit": 50, "decision": "any"}


def test_prepare_get_builds_run_block_scoped_url() -> None:
    reviews = WorkflowReviews(client=MagicMock())
    request = reviews.prepare_get("run_1", "extract-1")
    assert request.method == "GET"
    assert request.url == "/workflows/reviews/run_1/extract-1"


def test_prepare_create_version_posts_snapshot_to_versions() -> None:
    reviews = WorkflowReviews(client=MagicMock())
    request = reviews.prepare_create_version(
        "run_1",
        "extract-1",
        snapshot={"category": "Invoice"},
        parent_id=_VERSION_ID,
        origin="agent_created",
        note="changed category",
    )
    assert request.method == "POST"
    assert request.url == "/workflows/reviews/run_1/extract-1/versions"
    assert request.data == {
        "snapshot": {"category": "Invoice"},
        "parent_id": _VERSION_ID,
        "origin": "agent_created",
        "note": "changed category",
    }


def test_prepare_decision_posts_verdict_and_version_id() -> None:
    reviews = WorkflowReviews(client=MagicMock())
    request = reviews.prepare_decision(
        "run_1",
        "extract-1",
        verdict="rejected",
        version_id=_VERSION_ID,
        reason="wrong vendor",
    )
    assert request.method == "POST"
    assert request.url == "/workflows/reviews/run_1/extract-1/decision"
    assert request.data == {"verdict": "rejected", "version_id": _VERSION_ID, "reason": "wrong vendor"}


def test_list_parses_queue_response() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "data": [_QUEUE_ITEM],
        "list_metadata": {"before": None, "after": None},
    }

    result = WorkflowReviews(client=client).list(workflow_id="wf_1")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/reviews"
    assert isinstance(result, ReviewQueueResponse)
    assert result.has_more is False


def test_list_forwards_before_and_after_cursors() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "data": [_QUEUE_ITEM],
        "list_metadata": {"before": None, "after": "brun_xxx"},
    }

    result = WorkflowReviews(client=client).list(after="brun_prev", limit=25)

    request = client._prepared_request.call_args.args[0]
    assert request.params == {"limit": 25, "decision": "none", "after": "brun_prev"}
    assert result.has_more is True
    assert result.list_metadata.after == "brun_xxx"


def test_get_parses_overlay() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _OVERLAY

    overlay = WorkflowReviews(client=client).get("run_1", "extract-1")

    request = client._prepared_request.call_args.args[0]
    assert request.url == "/workflows/reviews/run_1/extract-1"
    assert overlay.versions_by_id[_VERSION_ID].origin == "model_output"


def test_approve_sends_approved_verdict() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {"submission_status": "accepted", "review": _OVERLAY}

    resp = WorkflowReviews(client=client).approve("run_1", "extract-1", version_id=_VERSION_ID)

    request = client._prepared_request.call_args.args[0]
    assert request.url == "/workflows/reviews/run_1/extract-1/decision"
    assert request.data == {"verdict": "approved", "version_id": _VERSION_ID}
    assert resp.submission_status == "accepted"


def test_reject_requires_reason_and_sends_rejected_verdict() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {"submission_status": "accepted", "review": _OVERLAY}

    WorkflowReviews(client=client).reject("run_1", "extract-1", version_id=_VERSION_ID, reason="bad data")

    request = client._prepared_request.call_args.args[0]
    assert request.data == {"verdict": "rejected", "version_id": _VERSION_ID, "reason": "bad data"}


def test_create_version_posts_version_and_returns_overlay() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _OVERLAY

    overlay = WorkflowReviews(client=client).create_version(
        "run_1",
        "extract-1",
        snapshot={"category": "Invoice"},
        parent_id=_VERSION_ID,
    )

    request = client._prepared_request.call_args.args[0]
    assert request.url == "/workflows/reviews/run_1/extract-1/versions"
    assert request.data["parent_id"] == _VERSION_ID
    assert request.data["origin"] == "human_created"
    assert isinstance(overlay, ReviewOverlay)


def test_reject_requires_reason_keyword() -> None:
    reviews = WorkflowReviews(client=MagicMock())
    with pytest.raises(TypeError):
        reviews.reject("run_1", "extract-1", version_id=_VERSION_ID)  # type: ignore[call-arg]


def test_wait_for_returns_when_decision_absent() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {**_OVERLAY, "decision": None}

    overlay = WorkflowReviews(client=client).wait_for("run_1", "extract-1", timeout=5.0, poll_interval=0.01)
    assert overlay.decision is None


def test_wait_for_times_out_when_overlay_never_appears() -> None:
    client = MagicMock()
    client._prepared_request.side_effect = NotFoundError("nope", status_code=404)

    with pytest.raises(TimeoutError):
        WorkflowReviews(client=client).wait_for("run_1", "extract-1", timeout=0.05, poll_interval=0.01)


@pytest.mark.asyncio
async def test_async_get_parses_overlay() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(return_value=_OVERLAY)

    overlay = await AsyncWorkflowReviews(client=client).get("run_1", "extract-1")

    request = client._prepared_request.call_args.args[0]
    assert request.url == "/workflows/reviews/run_1/extract-1"
    assert overlay.decision is not None


@pytest.mark.asyncio
async def test_async_approve_sends_approved_verdict() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(return_value={"submission_status": "already_applied", "review": _OVERLAY})

    resp = await AsyncWorkflowReviews(client=client).approve("run_1", "extract-1", version_id=_VERSION_ID)

    request = client._prepared_request.call_args.args[0]
    assert request.data["version_id"] == _VERSION_ID
    assert resp.submission_status == "already_applied"


@pytest.mark.asyncio
async def test_async_wait_for_times_out() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(side_effect=NotFoundError("nope", status_code=404))

    with pytest.raises(TimeoutError):
        await AsyncWorkflowReviews(client=client).wait_for("run_1", "extract-1", timeout=0.05, poll_interval=0.01)
