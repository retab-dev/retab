# pyright: reportArgumentType=false, reportCallIssue=false
"""Smoke tests for `client.workflows.tests.*` — wires + URL/method/body shape.

Mirrors the testing convention from `test_workflows.py`: mock the underlying
`client._prepared_request` so we can assert on the constructed `PreparedRequest`
without actually hitting the API.
"""

from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock

import pytest

from retab.resources.workflows import AsyncWorkflows, Workflows
from retab.resources.workflows.tests import (
    AsyncWorkflowTests,
    WorkflowTests,
)


_NOW = "2026-05-01T14:30:00Z"
_WORKFLOW_REF = {
    "workflow_id": "wf_abc123",
    "version_id": "draft_1",
    "name_at_run_time": "Q1 workflow",
    "requested_version": "draft",
}
_TRIGGER = {"type": "api"}
_PENDING = {"status": "pending"}
_COMPLETED = {"status": "completed"}
_TIMING = {"created_at": _NOW, "started_at": _NOW, "completed_at": _NOW}

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
    "schema_drift": "none",
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


def test_create_posts_to_tests_route_with_full_body() -> None:
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
    assert request.url == "/v1/workflows/tests"
    assert request.data["workflow_id"] == "wf_abc123"
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


def test_create_with_manual_source_serializes_typed_handle_inputs() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _TEST_RESPONSE

    Workflows(client=client).tests.create(
        workflow_id="wf_abc123",
        target={"type": "block", "block_id": "block_extract"},
        source={
            "type": "manual",
            "handle_inputs": {"input-json-0": {"type": "json", "data": {"subtotal": 100}}},
        },
        assertion={
            "target": {"output_handle_id": "output-json-0", "path": "total"},
            "condition": {"kind": "exists"},
        },
    )

    request = client._prepared_request.call_args.args[0]
    assert request.data["source"] == {
        "type": "manual",
        "handle_inputs": {"input-json-0": {"type": "json", "data": {"subtotal": 100}}},
    }


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

    Workflows(client=client).tests.get("wfnodetest_abc")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/v1/workflows/tests/wfnodetest_abc"


def test_list_uses_tests_route_with_filter() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "data": [],
        "list_metadata": {"before": None, "after": None},
    }

    result = Workflows(client=client).tests.list(
        workflow_id="wf_abc123",
        target_block_id="block_extract",
        limit=25,
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/v1/workflows/tests"
    assert request.params == {"limit": 25, "workflow_id": "wf_abc123", "target_block_id": "block_extract"}
    assert result.data == []
    assert result.list_metadata.before is None
    assert result.list_metadata.after is None


def test_list_omits_target_block_id_when_unset() -> None:
    """The query param `target_block_id` must be omitted entirely (not
    `None`) when the caller doesn't supply it — otherwise the backend
    string-coerces and tries to match a literal `"None"`.
    """
    client = MagicMock()
    client._prepared_request.return_value = {
        "data": [],
        "list_metadata": {"before": None, "after": None},
    }

    Workflows(client=client).tests.list(workflow_id="wf_abc123")

    request = client._prepared_request.call_args.args[0]
    assert "target_block_id" not in request.params


def test_delete_uses_test_detail_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = None

    Workflows(client=client).tests.delete("wfnodetest_abc")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "DELETE"
    assert request.url == "/v1/workflows/tests/wfnodetest_abc"


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
        test_id="wfnodetest_abc",
        name="renamed",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "PATCH"
    assert request.url == "/v1/workflows/tests/wfnodetest_abc"
    assert request.data == {"name": "renamed"}, f"PATCH body must only carry the field the caller passed; got {request.data!r}"


def test_update_with_assertion_serializes_assertion_only() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _TEST_RESPONSE

    Workflows(client=client).tests.update(
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
# runs.create — three call patterns + n_consensus
# ---------------------------------------------------------------------------


def _run_response(**overrides: object) -> dict:
    base = {
        "id": "wftestrun_q1z2",
        "workflow": _WORKFLOW_REF,
        "trigger": _TRIGGER,
        "lifecycle": _PENDING,
        "timing": _TIMING,
        "target": {"type": "block", "block_id": "block_extract"},
        "test_id": None,
        "total_tests": 4,
        "counts": {"queued": 4},
    }
    base.update(overrides)
    return base


def test_runs_create_with_test_id_only() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _run_response()

    response = Workflows(client=client).tests.runs.create(
        "wf_abc123",
        test_id="wfnodetest_abc",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/v1/workflows/tests/runs"
    assert request.data == {
        "workflow_id": "wf_abc123",
        "test_id": "wfnodetest_abc",
    }
    assert response.lifecycle.status == "pending"
    assert response.id == "wftestrun_q1z2"


def test_runs_create_with_target_and_consensus() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _run_response()

    Workflows(client=client).tests.runs.create(
        "wf_abc123",
        target={"type": "block", "block_id": "block_extract"},
        n_consensus=5,
    )

    request = client._prepared_request.call_args.args[0]
    assert request.data == {
        "workflow_id": "wf_abc123",
        "target": {"type": "block", "block_id": "block_extract"},
        "n_consensus": 5,
    }


def test_runs_create_no_args_runs_every_test_in_workflow() -> None:
    """Only workflow_id in the body = run every test in the workflow."""
    client = MagicMock()
    client._prepared_request.return_value = _run_response()

    Workflows(client=client).tests.runs.create("wf_abc123")

    request = client._prepared_request.call_args.args[0]
    assert request.data == {"workflow_id": "wf_abc123"}


# ---------------------------------------------------------------------------
# runs sub-resource
# ---------------------------------------------------------------------------


_RUN_RESPONSE = {
    **_run_response(lifecycle=_COMPLETED, counts={"passed": 1}),
}

_RESULT_RESPONSE = {
    "id": "wfresult_abc",
    "run_id": "wftestrun_q1z2",
    "test_id": "wfnodetest_abc",
    "workflow_id": "wf_abc123",
    "lifecycle": _COMPLETED,
    "timing": _TIMING,
    "target": {"type": "block", "block_id": "block_extract"},
    "source": {"type": "manual", "handle_inputs": {}},
    "execution_fingerprint": "f1",
    "outputs": {"output-json-0": {"total": 1234.56}},
    "warnings": [],
    "skipped": False,
    "verdict": "passed",
}


def test_workflow_test_result_accepts_bare_string_verdict() -> None:
    """Regression: the server emits ``verdict`` as one of the bare strings
    ``"passed" | "failed" | "blocked"`` (or ``None``). The SDK type previously
    declared it as ``dict[str, Any] | None``, which made every list+get call
    against real data raise ``ValidationError``."""
    from retab.types.workflows.tests import WorkflowTestResult

    for verdict in ("passed", "failed", "blocked", None):
        payload = {**_RESULT_RESPONSE, "verdict": verdict}
        parsed = WorkflowTestResult.model_validate(payload)
        assert parsed.verdict == verdict


def test_runs_list_uses_canonical_runs_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "data": [],
        "list_metadata": {"before": None, "after": None},
    }

    result = Workflows(client=client).tests.runs.list(
        workflow_id="wf_abc123",
        test_id="wfnodetest_abc",
        limit=10,
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/v1/workflows/tests/runs"
    assert request.params == {
        "limit": 10,
        "workflow_id": "wf_abc123",
        "test_id": "wfnodetest_abc",
    }
    assert result.data == []
    assert result.list_metadata.before is None
    assert result.list_metadata.after is None


def test_runs_get_uses_run_id_first_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _RUN_RESPONSE

    run = Workflows(client=client).tests.runs.get("wftestrun_q1z2")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/v1/workflows/tests/runs/wftestrun_q1z2"
    assert run.lifecycle.status == "completed"


def test_runs_cancel_uses_run_id_first_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _run_response(lifecycle={"status": "cancelled"})

    run = Workflows(client=client).tests.runs.cancel("wftestrun_q1z2")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/v1/workflows/tests/runs/wftestrun_q1z2/cancel"
    assert request.data == {}
    assert run.lifecycle.status == "cancelled"


def test_workflow_test_run_lifecycle_uses_test_status_values() -> None:
    from pydantic import ValidationError

    from retab.types.workflows import WorkflowTestRun

    run = WorkflowTestRun.model_validate(_run_response(lifecycle={"status": "error", "message": "boom"}))
    assert run.lifecycle.status == "error"
    assert run.lifecycle.message == "boom"

    with pytest.raises(ValidationError):
        WorkflowTestRun.model_validate(_run_response(lifecycle={"status": "awaiting_review"}))


def test_workflow_test_result_matches_run_record_shape() -> None:
    from retab.types.workflows import WorkflowTestResult

    result = WorkflowTestResult.model_validate(
        {
            **_RESULT_RESPONSE,
            "run_id": None,
            "lifecycle": None,
            "timing": None,
        }
    )

    assert result.workflow_id == "wf_abc123"
    assert result.run_id is None
    assert result.lifecycle is None
    assert result.timing is None


def test_workflow_test_drift_enums_match_openapi() -> None:
    from pydantic import ValidationError

    from retab.types.workflows import WorkflowTest

    valid = WorkflowTest.model_validate(
        {
            **_TEST_RESPONSE,
            "assertion_drift_status": "valid",
            "schema_drift": "partial",
        }
    )
    assert valid.assertion_drift_status == "valid"
    assert valid.schema_drift == "partial"

    with pytest.raises(ValidationError):
        WorkflowTest.model_validate({**_TEST_RESPONSE, "schema_drift": "fresh"})

    with pytest.raises(ValidationError):
        WorkflowTest.model_validate({**_TEST_RESPONSE, "assertion_drift_status": "unknown"})


def test_runs_results_list_uses_run_id_first_results_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "data": [_RESULT_RESPONSE],
        "list_metadata": {"before": None, "after": None},
    }

    result = Workflows(client=client).tests.results.list("wftestrun_q1z2")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/v1/workflows/tests/results"
    assert request.params == {"run_id": "wftestrun_q1z2", "limit": 20}
    assert result.data[0].test_id == "wfnodetest_abc"
    assert result.data[0].outputs == {"output-json-0": {"total": 1234.56}}


def test_runs_results_get_uses_flat_result_id_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _RESULT_RESPONSE

    result = Workflows(client=client).tests.results.get("wfresult_abc")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/v1/workflows/tests/results/wfresult_abc"
    assert result.id == "wfresult_abc"
    assert result.test_id == "wfnodetest_abc"


def test_tests_hard_cutover_removes_legacy_execute_and_scoped_run_aliases() -> None:
    client = MagicMock()
    tests = Workflows(client=client).tests

    assert not hasattr(tests, "execute")
    assert not hasattr(tests.runs, "get_execution")
    assert hasattr(tests, "results")
    assert hasattr(tests.results, "get")
    assert not callable(tests.results)


# ---------------------------------------------------------------------------
# Async parity — at least one async call MUST behave the same way
# ---------------------------------------------------------------------------


@pytest.mark.asyncio
async def test_async_create_posts_to_tests_route() -> None:
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
    assert request.url == "/v1/workflows/tests"
    assert request.data["workflow_id"] == "wf_abc123"
    assert test.id == "wfnodetest_abc"


@pytest.mark.asyncio
async def test_async_runs_get_returns_typed_record() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(return_value=_RUN_RESPONSE)

    run = await AsyncWorkflows(client=client).tests.runs.get("wftestrun_q1z2")

    assert run.id == "wftestrun_q1z2"
    assert run.lifecycle.status == "completed"


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
        test_id="wfnodetest_abc",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.data == {}, f"update() with no field kwargs must send an empty body; got {request.data!r}"


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


@pytest.mark.asyncio
async def test_async_runs_list_uses_test_runs_route() -> None:
    """Sync coverage exists at `test_runs_list_uses_test_runs_route`; pin
    the async equivalent so future refactors can't drop one without the
    other.
    """
    client = MagicMock()
    client._prepared_request = AsyncMock(
        return_value={"data": [], "list_metadata": {"before": None, "after": None}},
    )

    await AsyncWorkflows(client=client).tests.runs.list(
        workflow_id="wf_abc123",
        test_id="wfnodetest_abc",
        limit=10,
    )

    request = client._prepared_request.call_args.args[0]
    assert request.url == "/v1/workflows/tests/runs"
    assert request.params == {
        "limit": 10,
        "workflow_id": "wf_abc123",
        "test_id": "wfnodetest_abc",
    }


def test_workflow_tests_do_not_expose_wait_for_completion() -> None:
    client = MagicMock()

    assert not hasattr(Workflows(client=client).tests, "wait_for_completion")
    assert not hasattr(AsyncWorkflows(client=client).tests, "wait_for_completion")
