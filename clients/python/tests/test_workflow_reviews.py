# pyright: reportAttributeAccessIssue=false
"""Smoke tests for ``client.workflows.reviews.*``."""

from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock

import pytest

from retab.resources.workflows.reviews import AsyncWorkflowReviews, WorkflowReviews
from retab.types.pagination import PaginatedList
from retab.types.workflows import (
    Actor,
    Review,
    ReviewDecision,
    ReviewVersion,
    SubmitDecisionResponse,
)

_NOW = "2026-05-01T14:30:00Z"
_REVIEW_ID = "rev_1"
_VERSION_ID = "rvr_AAAAAAAAAAAAAAAAAAAAAAAAAA"
_CHILD_VERSION_ID = "rvr_BBBBBBBBBBBBBBBBBBBBBBBBBB"

_ACTOR_HUMAN = {"kind": "human", "id": "user_1", "display_name": "Ada"}
_ACTOR_MODEL = {"kind": "model", "id": "gpt", "display_name": "Extractor"}
_ACTOR_AGENT = {"kind": "agent", "id": "agent_1", "display_name": "Reviewer Agent"}

_OUTPUT_VERSION = {
    "id": _VERSION_ID,
    "review_id": _REVIEW_ID,
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
    "created_at": _NOW,
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
    "triggered_by": {"kind": "confidence_lt", "threshold": 0.8},
    "created_at": _NOW,
    "decision": _DECISION,
}

_CHILD_OUTPUT_VERSION = {
    **_OUTPUT_VERSION,
    "id": _CHILD_VERSION_ID,
    "parent_id": _VERSION_ID,
    "author": _ACTOR_AGENT,
    "snapshot": {"output": {"total": 150, "currency": "USD"}},
}

_QUEUE_ROW = {
    "id": _REVIEW_ID,
    "workflow_id": "wf_1",
    "workflow_version_id": "wv_1",
    "workflow_run_id": "run_1",
    "block_id": "extract-1",
    "step_id": "step_1",
    "parent_step_id": None,
    "iteration_key": None,
    "block_type": "extract",
    "triggered_by": {"kind": "confidence_lt", "threshold": 0.8},
    "created_at": _NOW,
    "decision": None,
}


def test_review_round_trips() -> None:
    review = Review.model_validate(_REVIEW)
    assert review.id == _REVIEW_ID
    assert "versions" not in review.model_dump()
    assert review.decision is not None
    assert review.decision.version_id == _VERSION_ID


def test_review_queue_row_round_trips_without_full_history() -> None:
    item = Review.model_validate(_QUEUE_ROW)
    assert item.id == _REVIEW_ID
    dumped = item.model_dump()
    assert "versions" not in dumped


def test_review_queue_response_round_trips() -> None:
    resp = PaginatedList[Review].model_validate({"data": [_QUEUE_ROW], "list_metadata": {"before": None, "after": _REVIEW_ID}})
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


def test_review_version_round_trips() -> None:
    version = ReviewVersion.model_validate(_CHILD_OUTPUT_VERSION)
    assert version.id == _CHILD_VERSION_ID
    assert version.review_id == _REVIEW_ID
    assert version.parent_id == _VERSION_ID


def test_review_version_list_response_round_trips() -> None:
    resp = PaginatedList[ReviewVersion].model_validate(
        {
            "data": [_OUTPUT_VERSION, _CHILD_OUTPUT_VERSION],
            "list_metadata": {"before": _VERSION_ID, "after": _CHILD_VERSION_ID},
        }
    )
    assert resp.has_more is True
    assert resp.data[1].id == _CHILD_VERSION_ID


def test_actor_kind_is_neutral_data_not_a_branch() -> None:
    for raw in (_ACTOR_MODEL, _ACTOR_AGENT, _ACTOR_HUMAN):
        actor = Actor.model_validate(raw)
        assert actor.kind in {"model", "agent", "human"}


def test_output_version_and_decision_round_trip() -> None:
    version = ReviewVersion.model_validate(_OUTPUT_VERSION)
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
    assert request.url == "/v1/workflows/reviews"
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
    assert request.url == f"/v1/workflows/reviews/{_REVIEW_ID}"


def test_review_versions_prepare_create_get_list() -> None:
    versions = WorkflowReviews(client=MagicMock()).versions
    create = versions.prepare_create(
        review_id=_REVIEW_ID,
        snapshot={"category": "Invoice"},
        parent_id=_VERSION_ID,
    )
    get = versions.prepare_get(_CHILD_VERSION_ID)
    list_request = versions.prepare_list(review_id=_REVIEW_ID, limit=25, after=_VERSION_ID)
    assert create.method == "POST"
    assert create.url == "/v1/workflows/reviews/versions"
    assert create.data == {
        "review_id": _REVIEW_ID,
        "snapshot": {"category": "Invoice"},
        "parent_id": _VERSION_ID,
    }
    assert get.method == "GET"
    assert get.url == f"/v1/workflows/reviews/versions/{_CHILD_VERSION_ID}"
    assert list_request.method == "GET"
    assert list_request.url == "/v1/workflows/reviews/versions"
    assert list_request.params == {
        "review_id": _REVIEW_ID,
        "limit": 25,
        "after": _VERSION_ID,
    }


def test_prepare_approve_and_reject_use_split_endpoints() -> None:
    reviews = WorkflowReviews(client=MagicMock())
    approve = reviews.prepare_approve(_REVIEW_ID, version_id=_VERSION_ID)
    reject = reviews.prepare_reject(_REVIEW_ID, version_id=_VERSION_ID, reason="wrong vendor")
    assert approve.url == f"/v1/workflows/reviews/{_REVIEW_ID}/approve"
    assert approve.data == {"version_id": _VERSION_ID}
    assert reject.url == f"/v1/workflows/reviews/{_REVIEW_ID}/reject"
    assert reject.data == {"version_id": _VERSION_ID, "reason": "wrong vendor"}


def test_list_get_approve_parse_responses() -> None:
    client = MagicMock()
    client._prepared_request.side_effect = [
        {"data": [_QUEUE_ROW], "list_metadata": {"before": None, "after": None}},
        _REVIEW,
        {"submission_status": "accepted", "review": _REVIEW, "resume_status": "resumed"},
    ]
    reviews = WorkflowReviews(client=client)

    queue = reviews.list(workflow_id="wf_1")
    assert isinstance(queue, PaginatedList)
    assert queue.data[0].id == _REVIEW_ID
    assert isinstance(reviews.get(_REVIEW_ID), Review)
    assert reviews.approve(_REVIEW_ID, version_id=_VERSION_ID).resume_status == "resumed"


def test_review_versions_create_get_list_parse_responses() -> None:
    client = MagicMock()
    client._prepared_request.side_effect = [
        _CHILD_OUTPUT_VERSION,
        _CHILD_OUTPUT_VERSION,
        {
            "data": [_OUTPUT_VERSION, _CHILD_OUTPUT_VERSION],
            "list_metadata": {"before": None, "after": None},
        },
    ]
    versions = WorkflowReviews(client=client).versions
    assert (
        versions.create(
            review_id=_REVIEW_ID,
            parent_id=_VERSION_ID,
            snapshot={"category": "Invoice"},
        ).id
        == _CHILD_VERSION_ID
    )
    assert versions.get(_CHILD_VERSION_ID).parent_id == _VERSION_ID
    assert versions.list(review_id=_REVIEW_ID).data[0].id == _VERSION_ID


def test_review_versions_create_requires_parent_id() -> None:
    versions = WorkflowReviews(client=MagicMock()).versions
    with pytest.raises(TypeError):
        versions.create(review_id=_REVIEW_ID, snapshot={"category": "Invoice"})  # type: ignore[call-arg]

    request = versions.prepare_create(
        review_id=_REVIEW_ID,
        parent_id="",
        snapshot={"category": "Invoice"},
    )
    assert request.data == {
        "review_id": _REVIEW_ID,
        "parent_id": "",
        "snapshot": {"category": "Invoice"},
    }


def test_reject_requires_reason_keyword() -> None:
    reviews = WorkflowReviews(client=MagicMock())
    with pytest.raises(TypeError):
        reviews.reject(_REVIEW_ID, version_id=_VERSION_ID)  # type: ignore[call-arg]


def test_reviews_do_not_expose_wait_for() -> None:
    assert not hasattr(WorkflowReviews(client=MagicMock()), "wait_for")
    assert not hasattr(AsyncWorkflowReviews(client=MagicMock()), "wait_for")


@pytest.mark.asyncio
async def test_async_get_and_approve() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(side_effect=[_REVIEW, {"submission_status": "accepted", "review": _REVIEW}])
    reviews = AsyncWorkflowReviews(client=client)
    assert (await reviews.get(_REVIEW_ID)).id == _REVIEW_ID
    assert (await reviews.approve(_REVIEW_ID, version_id=_VERSION_ID)).submission_status == "accepted"
