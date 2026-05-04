"""Smoke tests for `client.workflows.tests.*` — wires + URL/method/body shape.

Mirrors the testing convention from `test_workflows.py`: mock the underlying
`client._prepared_request` so we can assert on the constructed `PreparedRequest`
without actually hitting the API.
"""

from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock

import pytest

from retab.resources.workflows.client import AsyncWorkflows, Workflows
from retab.resources.workflows.tests.client import (
    AsyncWorkflowTests,
    WorkflowTests,
)


_NOW = "2026-05-01T14:30:00Z"

_TEST_RESPONSE = {
    "id": "wfnodetest_abc",
    "workflow_id": "wf_abc123",
    "organization_id": "org_x",
    "target": {"type": "block", "block_id": "block_extract"},
    "source": {"type": "manual", "handle_inputs": {}},
    "name": "Q1 invoice total",
    "assertion": {
        "id": "assert_xyz",
        "target": {"output_handle_id": "output-json-0", "path": "total"},
        "condition": {"kind": "equals", "expected": 1234.56},
        "label": None,
    },
    "schema_drift": "fresh",
    "validation_status": "valid",
    "validation_issues": [],
    "latest_run_summary": None,
    "latest_passing_run_summary": None,
    "latest_failing_run_summary": None,
    "created_at": _NOW,
    "updated_at": _NOW,
}


# ---------------------------------------------------------------------------
# Wiring — `client.workflows.tests` and `tests.runs` are present
# ---------------------------------------------------------------------------


def test_workflows_client_exposes_tests_subresource() -> None:
    """The `tests` sub-resource MUST be reachable from `client.workflows`.
    Without this wire, the docs example `client.workflows.tests.create(...)`
    raises `AttributeError` at import time for callers.
    """
    client = MagicMock()
    workflows = Workflows(client=client)
    assert isinstance(workflows.tests, WorkflowTests)


def test_async_workflows_client_exposes_tests_subresource() -> None:
    client = MagicMock()
    workflows = AsyncWorkflows(client=client)
    assert isinstance(workflows.tests, AsyncWorkflowTests)


def test_tests_resource_exposes_runs_subresource() -> None:
    """`tests.runs.list(...)` and `tests.runs.get(...)` are the only way
    to read run records from the SDK — the runs sub-resource MUST be
    initialized when WorkflowTests is constructed.
    """
    client = MagicMock()
    tests = WorkflowTests(client=client)
    assert tests.runs is not None
    # And the same on the async side.
    async_tests = AsyncWorkflowTests(client=client)
    assert async_tests.runs is not None


# ---------------------------------------------------------------------------
# create — body must carry target / source / assertion in wire shape
# ---------------------------------------------------------------------------


def test_create_posts_to_block_tests_route_with_full_body() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _TEST_RESPONSE

    test = Workflows(client=client).tests.create(
        workflow_id="wf_abc123",
        target={"type": "block", "block_id": "block_extract"},
        source={"type": "manual", "handle_inputs": {}},
        assertion={
            "target": {"output_handle_id": "output-json-0", "path": "total"},
            "condition": {"kind": "equals", "expected": 1234.56},
        },
        name="Q1 invoice total",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows/wf_abc123/block-tests"
    # Pydantic dumps the discriminated union with `type` set explicitly.
    assert request.data["target"] == {"type": "block", "block_id": "block_extract"}
    assert request.data["source"] == {
        "type": "manual",
        "handle_inputs": {},
    }
    assert request.data["assertion"]["target"] == {
        "output_handle_id": "output-json-0",
        "path": "total",
    }
    assert request.data["name"] == "Q1 invoice total"
    assert test.id == "wfnodetest_abc"


def test_create_with_run_step_source_serializes_step_id() -> None:
    """The `run_step` variant must carry both `run_id` and `step_id` (the
    `step_id` is required for blocks executed inside a `for_each`)."""
    client = MagicMock()
    client._prepared_request.return_value = _TEST_RESPONSE

    Workflows(client=client).tests.create(
        workflow_id="wf_abc123",
        target={"type": "block", "block_id": "block_extract"},
        source={
            "type": "run_step",
            "run_id": "wfrun_xyz",
            "step_id": "block_extract",
        },
        assertion={
            "target": {"output_handle_id": "output-json-0", "path": "total"},
            "condition": {"kind": "exists"},
        },
    )

    request = client._prepared_request.call_args.args[0]
    assert request.data["source"] == {
        "type": "run_step",
        "run_id": "wfrun_xyz",
        "step_id": "block_extract",
    }
    # `name` was not provided — must NOT appear in the body (the route
    # validates with `extra="forbid"` only on the request model, but
    # cleaner request payloads are easier to debug from server logs).
    assert "name" not in request.data


def test_create_rejects_unknown_source_type() -> None:
    """An unknown source `type` discriminator must raise — the SDK relies on
    Pydantic's discriminated-union TypeAdapter to surface this as a typed
    `ValidationError` (subclass of ValueError) naming the invalid tag.
    """
    from pydantic import ValidationError

    client = MagicMock()
    with pytest.raises((ValueError, ValidationError)) as exc_info:
        Workflows(client=client).tests.create(
            workflow_id="wf_abc123",
            target={"type": "block", "block_id": "block_x"},
            source={"type": "from_thin_air", "handle_inputs": {}},  # type: ignore[arg-type]
            assertion={
                "target": {"output_handle_id": "output-json-0", "path": "total"},
                "condition": {"kind": "exists"},
            },
        )
    # The error must mention the offending tag so the caller can debug.
    assert "from_thin_air" in str(exc_info.value)


# ---------------------------------------------------------------------------
# get / list / delete — URLs and params
# ---------------------------------------------------------------------------


def test_get_uses_test_detail_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _TEST_RESPONSE

    Workflows(client=client).tests.get("wf_abc123", "wfnodetest_abc")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/wf_abc123/block-tests/wfnodetest_abc"


def test_list_uses_block_tests_route_with_filter() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {"tests": []}

    result = Workflows(client=client).tests.list(
        workflow_id="wf_abc123",
        target_block_id="block_extract",
        limit=25,
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/wf_abc123/block-tests"
    assert request.params == {"limit": 25, "target_block_id": "block_extract"}
    assert result.tests == []


def test_list_omits_target_block_id_when_unset() -> None:
    """The query param `target_block_id` must be omitted entirely (not
    `None`) when the caller doesn't supply it — otherwise the backend
    string-coerces and tries to match a literal `"None"`.
    """
    client = MagicMock()
    client._prepared_request.return_value = {"tests": []}

    Workflows(client=client).tests.list(workflow_id="wf_abc123")

    request = client._prepared_request.call_args.args[0]
    assert "target_block_id" not in request.params


def test_delete_uses_test_detail_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = None

    Workflows(client=client).tests.delete("wf_abc123", "wfnodetest_abc")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "DELETE"
    assert request.url == "/workflows/wf_abc123/block-tests/wfnodetest_abc"


# ---------------------------------------------------------------------------
# update — body MUST omit fields the caller didn't pass
# ---------------------------------------------------------------------------


def test_update_only_includes_fields_the_caller_passed() -> None:
    """The backend treats missing PATCH fields as `leave untouched`.
    Including a field with value `None` would either trip Pydantic
    validation (for required-non-null fields) or — worse — clear the
    field on the stored doc. The SDK MUST omit, not nullify.
    """
    client = MagicMock()
    client._prepared_request.return_value = _TEST_RESPONSE

    Workflows(client=client).tests.update(
        workflow_id="wf_abc123",
        test_id="wfnodetest_abc",
        name="renamed",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "PATCH"
    assert request.url == "/workflows/wf_abc123/block-tests/wfnodetest_abc"
    assert request.data == {"name": "renamed"}, (
        f"PATCH body must only carry the field the caller passed; got {request.data!r}"
    )


def test_update_with_assertion_serializes_assertion_only() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _TEST_RESPONSE

    Workflows(client=client).tests.update(
        workflow_id="wf_abc123",
        test_id="wfnodetest_abc",
        assertion={
            "target": {"output_handle_id": "output-json-0", "path": "vendor.name"},
            "condition": {"kind": "matches_regex", "expected": "^Acme.*"},
        },
    )

    request = client._prepared_request.call_args.args[0]
    assert "name" not in request.data
    assert "source" not in request.data
    assert request.data["assertion"]["condition"]["kind"] == "matches_regex"


# ---------------------------------------------------------------------------
# execute — three call patterns + n_consensus
# ---------------------------------------------------------------------------


def _execute_response(**overrides: object) -> dict:
    base = {
        "batch_id": "btbatch_q1z2",
        "job_id": "job_abc",
        "status": "queued",
        "workflow_id": "wf_abc123",
        "target": {"type": "block", "block_id": "block_extract"},
        "test_id": None,
        "total_tests": 4,
    }
    base.update(overrides)
    return base


def test_execute_with_test_id_only() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _execute_response()

    response = Workflows(client=client).tests.execute(
        workflow_id="wf_abc123",
        test_id="wfnodetest_abc",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows/wf_abc123/block-tests/execute"
    assert request.data == {"test_id": "wfnodetest_abc"}
    assert response.status == "queued"


def test_execute_with_target_and_consensus() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _execute_response()

    Workflows(client=client).tests.execute(
        workflow_id="wf_abc123",
        target={"type": "block", "block_id": "block_extract"},
        n_consensus=5,
    )

    request = client._prepared_request.call_args.args[0]
    assert request.data == {
        "target": {"type": "block", "block_id": "block_extract"},
        "n_consensus": 5,
    }


def test_execute_no_args_runs_every_test_in_workflow() -> None:
    """Empty body = run every test in the workflow. The backend
    distinguishes this from `target=None` because Pydantic accepts the
    empty dict as a valid request.
    """
    client = MagicMock()
    client._prepared_request.return_value = _execute_response()

    Workflows(client=client).tests.execute(workflow_id="wf_abc123")

    request = client._prepared_request.call_args.args[0]
    assert request.data == {}


# ---------------------------------------------------------------------------
# runs sub-resource
# ---------------------------------------------------------------------------


_RUN_RESPONSE = {
    "id": "wfnodetestrun_abc",
    "test_id": "wfnodetest_abc",
    "workflow_id": "wf_abc123",
    "organization_id": "org_x",
    "target": {"type": "block", "block_id": "block_extract"},
    "source": {"type": "manual", "handle_inputs": {}},
    "status": "passed",
    "execution_fingerprint": "f1",
    "started_at": _NOW,
    "completed_at": _NOW,
    "duration_ms": 1234,
    "outputs": {"output-json-0": {"total": 1234.56}},
    "warnings": [],
    "skipped": False,
}


def test_runs_list_uses_test_runs_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {"runs": []}

    result = Workflows(client=client).tests.runs.list(
        workflow_id="wf_abc123",
        test_id="wfnodetest_abc",
        limit=10,
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/wf_abc123/block-tests/wfnodetest_abc/runs"
    assert request.params == {"limit": 10}
    assert result.runs == []


def test_runs_get_uses_run_detail_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _RUN_RESPONSE

    run = Workflows(client=client).tests.runs.get(
        workflow_id="wf_abc123",
        test_id="wfnodetest_abc",
        run_id="wfnodetestrun_abc",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert (
        request.url
        == "/workflows/wf_abc123/block-tests/wfnodetest_abc/runs/wfnodetestrun_abc"
    )
    assert run.status == "passed"
    # The renamed `outputs` field (formerly `handle_outputs`) parses cleanly.
    assert run.outputs == {"output-json-0": {"total": 1234.56}}


# ---------------------------------------------------------------------------
# Async parity — at least one async call MUST behave the same way
# ---------------------------------------------------------------------------


@pytest.mark.asyncio
async def test_async_create_posts_to_block_tests_route() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(return_value=_TEST_RESPONSE)

    test = await AsyncWorkflows(client=client).tests.create(
        workflow_id="wf_abc123",
        target={"type": "block", "block_id": "block_extract"},
        source={"type": "manual", "handle_inputs": {}},
        assertion={
            "target": {"output_handle_id": "output-json-0", "path": "total"},
            "condition": {"kind": "equals", "expected": 1234.56},
        },
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows/wf_abc123/block-tests"
    assert test.id == "wfnodetest_abc"


@pytest.mark.asyncio
async def test_async_runs_get_returns_typed_record() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(return_value=_RUN_RESPONSE)

    run = await AsyncWorkflows(client=client).tests.runs.get(
        workflow_id="wf_abc123",
        test_id="wfnodetest_abc",
        run_id="wfnodetestrun_abc",
    )

    assert run.outputs == {"output-json-0": {"total": 1234.56}}
    # Re-validate the discriminated union on the source side.
    assert run.source.type == "manual"


# ---------------------------------------------------------------------------
# Gap-coverage round 2 — caught in second-pass review
# ---------------------------------------------------------------------------


def test_update_with_no_kwargs_produces_empty_patch_body() -> None:
    """Calling update() with only the path args (no name/assertion/source)
    must produce an empty PATCH body. The backend treats missing fields
    as "leave untouched"; including them as null would clear the field
    on the stored doc.
    """
    client = MagicMock()
    client._prepared_request.return_value = _TEST_RESPONSE

    Workflows(client=client).tests.update(
        workflow_id="wf_abc123",
        test_id="wfnodetest_abc",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.data == {}, (
        f"update() with no field kwargs must send an empty body; got {request.data!r}"
    )


def test_workflow_test_response_parses_legacy_doc_with_null_assertion() -> None:
    """Pre-rewrite tests in storage may have ``assertion=None``. The SDK's
    `WorkflowTest` model has `assertion: AssertionSpec | None = None`
    matching the backend's relaxed shape. Without this, listing a workflow
    with even one orphan would crash the client at `model_validate`.
    """
    from retab.types.workflows import WorkflowTest

    legacy_payload = dict(_TEST_RESPONSE)
    legacy_payload["assertion"] = None
    test = WorkflowTest.model_validate(legacy_payload)
    assert test.assertion is None


def test_async_runs_list_uses_test_runs_route() -> None:
    """Sync coverage exists at `test_runs_list_uses_test_runs_route`; pin
    the async equivalent so future refactors can't drop one without the
    other.
    """
    import asyncio

    client = MagicMock()
    client._prepared_request = AsyncMock(return_value={"runs": []})

    async def _go() -> None:
        await AsyncWorkflows(client=client).tests.runs.list(
            workflow_id="wf_abc123",
            test_id="wfnodetest_abc",
            limit=10,
        )

    asyncio.run(_go())

    request = client._prepared_request.call_args.args[0]
    assert request.url == "/workflows/wf_abc123/block-tests/wfnodetest_abc/runs"
    assert request.params == {"limit": 10}


def test_wait_for_completion_returns_parsed_batch_result_when_job_completes() -> None:
    """`wait_for_completion(job_id)` polls `client.jobs.retrieve(job_id)`,
    returns the parsed `BlockTestBatchExecutionResult` once the job lands
    in a terminal state.
    """
    from retab.types.jobs import Job, JobResponse
    from retab.types.workflows import BlockTestBatchExecutionResult

    client = MagicMock()
    job = Job(
        id="job_abc",
        status="completed",
        endpoint="/v1/workflows/wf/block-tests/execute",
        response=JobResponse(
            status_code=200,
            body={
                "workflow_id": "wf_abc123",
                "counts": {"passed": 3, "failed": 1},
                "results": [],
            },
        ),
        created_at=0,
        expires_at=0,
        organization_id="org_x",
    )
    client.jobs.retrieve = MagicMock(return_value=job)

    result = Workflows(client=client).tests.wait_for_completion(
        "job_abc",
        poll_interval_seconds=0.001,
        timeout_seconds=1.0,
    )

    assert isinstance(result, BlockTestBatchExecutionResult)
    assert result.workflow_id == "wf_abc123"
    assert result.counts.passed == 3


def test_wait_for_completion_raises_runtime_error_on_failed_job() -> None:
    """A `failed` / `cancelled` / `expired` job must surface as a
    `RuntimeError` with the backend's error payload, not silently
    return None or fall through.
    """
    from retab.types.jobs import Job, JobError

    client = MagicMock()
    job = Job(
        id="job_bad",
        status="failed",
        endpoint="/v1/workflows/wf/block-tests/execute",
        error=JobError(code="INTERNAL_ERROR", message="something went wrong"),
        created_at=0,
        expires_at=0,
        organization_id="org_x",
    )
    client.jobs.retrieve = MagicMock(return_value=job)

    with pytest.raises(RuntimeError, match="failed"):
        Workflows(client=client).tests.wait_for_completion(
            "job_bad",
            poll_interval_seconds=0.001,
            timeout_seconds=1.0,
        )


def test_wait_for_completion_times_out_when_job_never_terminates() -> None:
    """If the job stays `queued` / `in_progress` past the deadline, the
    helper raises `TimeoutError` with the job_id named in the message.
    """
    from retab.types.jobs import Job

    client = MagicMock()
    job = Job(
        id="job_stuck",
        status="in_progress",
        endpoint="/v1/workflows/wf/block-tests/execute",
        created_at=0,
        expires_at=0,
        organization_id="org_x",
    )
    client.jobs.retrieve = MagicMock(return_value=job)

    with pytest.raises(TimeoutError, match="job_stuck"):
        Workflows(client=client).tests.wait_for_completion(
            "job_stuck",
            poll_interval_seconds=0.01,
            timeout_seconds=0.05,
        )
