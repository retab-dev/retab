"""Smoke tests for ``client.workflows.reviews.*``."""

from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock

import pytest

from retab.exceptions import NotFoundError
from retab.resources.workflows.reviews.client import AsyncWorkflowReviews, WorkflowReviews
from retab.types.workflows import (
    Actor,
    AppendVersionResponse,
    OutputVersion,
    Review,
    ReviewDecision,
    ReviewQueueResponse,
    ReviewSummary,
    SubmitDecisionResponse,
)

_NOW = "2026-05-01T14:30:00Z"
_REVIEW_ID = "rev_1"
_VERSION_ID = "a" * 64
_CHILD_VERSION_ID = "b" * 64

_ACTOR_HUMAN = {"kind": "human", "id": "user_1", "display_name": "Ada"}
_ACTOR_MODEL = {"kind": "model", "id": "gpt", "display_name": "Extractor"}
_ACTOR_AGENT = {"kind": "agent", "id": "agent_1", "display_name": "Reviewer Agent"}

_OUTPUT_VERSION = {
    "parent_id": None,
    "author": _ACTOR_MODEL,
    "snapshot": {"output": {"total": 100, "currency": "USD"}},
    "note": None,
    "created_at": _NOW,
}

_DECISION = {
    "verdict": "approved",
    "version_id": _VERSION_ID,
    "author": _ACTOR_HUMAN,
    "decided_at": _NOW,
    "reason": None,
}

_REVIEW = {
    "id": _REVIEW_ID,
    "workflow_id": "wf_1",
    "workflow_version_id": "wv_1",
    "workflow_run_id": "run_1",
    "block_id": "extract-1",
    "step_id": "step_1",
    "parent_step_id": None,
    "iteration_key": None,
    "block_type": "extract",
    "triggered_by": {"kind": "low_confidence", "threshold": 0.8},
    "awaiting_since": _NOW,
    "priority": 5,
    "versions": {
        _VERSION_ID: _OUTPUT_VERSION,
        _CHILD_VERSION_ID: {
            **_OUTPUT_VERSION,
            "parent_id": _VERSION_ID,
            "author": _ACTOR_AGENT,
            "snapshot": {"output": {"total": 150, "currency": "USD"}},
        },
    },
    "decision": _DECISION,
}

_SUMMARY = {
    "id": _REVIEW_ID,
    "workflow_id": "wf_1",
    "workflow_run_id": "run_1",
    "block_id": "extract-1",
    "step_id": "step_1",
    "parent_step_id": None,
    "iteration_key": None,
    "block_type": "extract",
    "triggered_by": {"kind": "low_confidence"},
    "awaiting_since": _NOW,
    "priority": 5,
    "seed_version_id": _VERSION_ID,
    "version_count": 2,
    "decision": None,
}


def test_review_round_trips() -> None:
    review = Review.model_validate(_REVIEW)
    assert review.id == _REVIEW_ID
    assert set(review.versions) == {_VERSION_ID, _CHILD_VERSION_ID}
    assert review.versions[_CHILD_VERSION_ID].parent_id == _VERSION_ID
    assert review.decision is not None
    assert review.decision.version_id == _VERSION_ID


def test_review_summary_round_trips_without_full_history() -> None:
    item = ReviewSummary.model_validate(_SUMMARY)
    assert item.id == _REVIEW_ID
    assert item.seed_version_id == _VERSION_ID
    dumped = item.model_dump()
    assert "versions" not in dumped


def test_review_queue_response_round_trips() -> None:
    resp = ReviewQueueResponse.model_validate({"data": [_SUMMARY], "list_metadata": {"before": None, "after": _REVIEW_ID}})
    assert resp.has_more is True
    assert resp.data[0].step_id == "step_1"


def test_submit_decision_response_accepts_pending_resume() -> None:
    resp = SubmitDecisionResponse.model_validate(
        {
            "submission_status": "accepted",
            "review": _REVIEW,
            "resume_status": "pending",
            "resume_error": "Temporal signal queued for retry",
        }
    )
    assert resp.resume_status == "pending"
    assert resp.resume_error == "Temporal signal queued for retry"


def test_append_version_response_round_trips() -> None:
    resp = AppendVersionResponse.model_validate({"append_status": "accepted", "version_id": _CHILD_VERSION_ID, "review": _REVIEW})
    assert resp.version_id == _CHILD_VERSION_ID
    assert resp.review.versions[_CHILD_VERSION_ID].parent_id == _VERSION_ID


def test_actor_kind_is_neutral_data_not_a_branch() -> None:
    for raw in (_ACTOR_MODEL, _ACTOR_AGENT, _ACTOR_HUMAN):
        actor = Actor.model_validate(raw)
        assert actor.kind in {"model", "agent", "human"}


def test_output_version_and_decision_round_trip() -> None:
    version = OutputVersion.model_validate(_OUTPUT_VERSION)
    assert version.parent_id is None
    assert version.author.kind == "model"

    decision = ReviewDecision.model_validate(_DECISION)
    assert decision.verdict == "approved"


def test_prepare_list_builds_get_with_hard_cutover_filters() -> None:
    reviews = WorkflowReviews(client=MagicMock())
    request = reviews.prepare_list(
        workflow_id="wf_1",
        run_id="run_1",
        block_id="extract-1",
        step_id="step_1",
        iteration_key="0",
        limit=10,
        decision_status="decided",
    )
    assert request.method == "GET"
    assert request.url == "/workflows/reviews"
    assert request.params == {
        "limit": 10,
        "decision_status": "decided",
        "workflow_id": "wf_1",
        "run_id": "run_1",
        "block_id": "extract-1",
        "step_id": "step_1",
        "iteration_key": "0",
    }


def test_prepare_get_builds_review_id_url() -> None:
    request = WorkflowReviews(client=MagicMock()).prepare_get(_REVIEW_ID)
    assert request.method == "GET"
    assert request.url == f"/workflows/reviews/{_REVIEW_ID}"


def test_prepare_append_version_posts_parent_version_id() -> None:
    request = WorkflowReviews(client=MagicMock()).prepare_append_version(
        _REVIEW_ID,
        snapshot={"category": "Invoice"},
        parent_version_id=_VERSION_ID,
        note="changed category",
    )
    assert request.method == "POST"
    assert request.url == f"/workflows/reviews/{_REVIEW_ID}/versions"
    assert request.data == {
        "snapshot": {"category": "Invoice"},
        "parent_version_id": _VERSION_ID,
        "note": "changed category",
    }


def test_prepare_approve_and_reject_use_split_endpoints() -> None:
    reviews = WorkflowReviews(client=MagicMock())
    approve = reviews.prepare_approve(_REVIEW_ID, version_id=_VERSION_ID)
    reject = reviews.prepare_reject(_REVIEW_ID, version_id=_VERSION_ID, reason="wrong vendor")
    assert approve.url == f"/workflows/reviews/{_REVIEW_ID}/approve"
    assert approve.data == {"version_id": _VERSION_ID}
    assert reject.url == f"/workflows/reviews/{_REVIEW_ID}/reject"
    assert reject.data == {"version_id": _VERSION_ID, "reason": "wrong vendor"}


def test_list_get_approve_append_parse_responses() -> None:
    client = MagicMock()
    client._prepared_request.side_effect = [
        {"data": [_SUMMARY], "list_metadata": {"before": None, "after": None}},
        _REVIEW,
        {"submission_status": "accepted", "review": _REVIEW, "resume_status": "resumed"},
        {"append_status": "accepted", "version_id": _CHILD_VERSION_ID, "review": _REVIEW},
    ]
    reviews = WorkflowReviews(client=client)

    assert isinstance(reviews.list(workflow_id="wf_1"), ReviewQueueResponse)
    assert isinstance(reviews.get(_REVIEW_ID), Review)
    assert reviews.approve(_REVIEW_ID, version_id=_VERSION_ID).resume_status == "resumed"
    assert (
        reviews.append_version(
            _REVIEW_ID,
            snapshot={"category": "Invoice"},
            parent_version_id=_VERSION_ID,
        ).version_id
        == _CHILD_VERSION_ID
    )


def test_reject_requires_reason_keyword() -> None:
    reviews = WorkflowReviews(client=MagicMock())
    with pytest.raises(TypeError):
        reviews.reject(_REVIEW_ID, version_id=_VERSION_ID)  # type: ignore[call-arg]


def test_wait_for_returns_when_decision_absent() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {**_REVIEW, "decision": None}
    review = WorkflowReviews(client=client).wait_for(_REVIEW_ID, timeout=5.0, poll_interval=0.01)
    assert review.decision is None


def test_wait_for_times_out_when_review_never_appears() -> None:
    client = MagicMock()
    client._prepared_request.side_effect = NotFoundError("nope", status_code=404)
    with pytest.raises(TimeoutError):
        WorkflowReviews(client=client).wait_for(_REVIEW_ID, timeout=0.05, poll_interval=0.01)


@pytest.mark.asyncio
async def test_async_get_and_approve() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(side_effect=[_REVIEW, {"submission_status": "already_applied", "review": _REVIEW}])
    reviews = AsyncWorkflowReviews(client=client)
    assert (await reviews.get(_REVIEW_ID)).id == _REVIEW_ID
    assert (await reviews.approve(_REVIEW_ID, version_id=_VERSION_ID)).submission_status == "already_applied"


@pytest.mark.asyncio
async def test_async_wait_for_times_out() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(side_effect=NotFoundError("nope", status_code=404))
    with pytest.raises(TimeoutError):
        await AsyncWorkflowReviews(client=client).wait_for(_REVIEW_ID, timeout=0.05, poll_interval=0.01)
