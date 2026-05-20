"""Smoke tests for `client.workflows.experiments.*`, `client.workflows.diagnose`, and
`client.workflows.blocks.simulate`.

Mirrors the existing pattern from `test_workflow_tests.py`: mock the
underlying `client._prepared_request` so we can assert on the constructed
`PreparedRequest` without hitting the API.
"""

from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock

import pytest

from retab.resources.workflows.client import AsyncWorkflows, Workflows
from retab.resources.workflows.experiments.client import (
    AsyncExperimentRuns,
    AsyncWorkflowExperiments,
    ExperimentRuns,
    WorkflowExperiments,
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


_EXPERIMENT_RESPONSE = {
    "id": "exp_abc",
    "workflow_id": "wf_abc123",
    "block_id": "block_extract",
    "n_consensus": 5,
    "document_count": 3,
    "name": "Q1 invoices",
    "last_run_id": None,
    "created_at": _NOW,
    "updated_at": _NOW,
    "status": "draft",
    "block_kind": "extract",
    "score": None,
    "is_stale": False,
    "schema_drift": "unknown",
    "schema_drift_detail": None,
}


# ---------------------------------------------------------------------------
# Wiring — `client.workflows.experiments` and `experiments.runs` are present
# ---------------------------------------------------------------------------


def test_workflows_client_exposes_experiments_subresource() -> None:
    client = MagicMock()
    workflows = Workflows(client=client)
    assert isinstance(workflows.experiments, WorkflowExperiments)


def test_async_workflows_client_exposes_experiments_subresource() -> None:
    client = MagicMock()
    workflows = AsyncWorkflows(client=client)
    assert isinstance(workflows.experiments, AsyncWorkflowExperiments)


def test_experiments_resource_exposes_runs_subresource() -> None:
    """`experiments.runs.create(...)` is the only path to trigger a run from
    the SDK — runs sub-resource MUST be initialized when WorkflowExperiments is.
    """
    client = MagicMock()
    experiments = WorkflowExperiments(client=client)
    assert isinstance(experiments.runs, ExperimentRuns)
    async_experiments = AsyncWorkflowExperiments(client=client)
    assert isinstance(async_experiments.runs, AsyncExperimentRuns)


# ---------------------------------------------------------------------------
# create — body must carry block_id, name, and the chosen document source
# ---------------------------------------------------------------------------


def test_experiments_create_posts_to_workflow_experiments_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _EXPERIMENT_RESPONSE

    experiment = Workflows(client=client).experiments.create(
        workflow_id="wf_abc123",
        block_id="block_extract",
        name="Q1 invoices",
        document_captures=[{"workflow_run_id": "wfr_1"}, {"workflow_run_id": "wfr_2", "step_id": "for_each-0"}],
        n_consensus=5,
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows/wf_abc123/experiments"
    assert request.data["block_id"] == "block_extract"
    assert request.data["name"] == "Q1 invoices"
    assert request.data["n_consensus"] == 5
    assert request.data["document_captures"] == [
        {"workflow_run_id": "wfr_1"},
        {"workflow_run_id": "wfr_2", "step_id": "for_each-0"},
    ]
    # ``documents`` not passed → omitted from the body (backend treats
    # missing as "no explicit docs", not "empty list").
    assert "documents" not in request.data
    assert experiment.id == "exp_abc"


def test_experiments_create_with_explicit_documents_serializes_handle_inputs() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _EXPERIMENT_RESPONSE

    Workflows(client=client).experiments.create(
        workflow_id="wf_abc123",
        block_id="block_extract",
        name="Manual set",
        documents=[
            {
                "handle_inputs": {
                    "input-file-0": {"type": "file", "document": {"id": "doc_1", "filename": "a.pdf"}},
                },
            }
        ],
    )

    request = client._prepared_request.call_args.args[0]
    assert request.data["documents"] == [
        {
            "handle_inputs": {
                "input-file-0": {"type": "file", "document": {"id": "doc_1", "filename": "a.pdf"}},
            },
        }
    ]


# ---------------------------------------------------------------------------
# update — only fields the caller passed end up in the PATCH body
# ---------------------------------------------------------------------------


def test_experiments_update_sends_only_provided_fields() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {**_EXPERIMENT_RESPONSE, "name": "Renamed"}

    Workflows(client=client).experiments.update(
        workflow_id="wf_abc123",
        experiment_id="exp_abc",
        name="Renamed",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "PATCH"
    assert request.url == "/workflows/wf_abc123/experiments/exp_abc"
    assert request.data == {"name": "Renamed"}


def test_experiments_update_passing_n_consensus_only() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _EXPERIMENT_RESPONSE

    Workflows(client=client).experiments.update(
        workflow_id="wf_abc123",
        experiment_id="exp_abc",
        n_consensus=7,
    )

    request = client._prepared_request.call_args.args[0]
    assert request.data == {"n_consensus": 7}


# ---------------------------------------------------------------------------
# list / get / delete / duplicate / cancel — URL + method shape
# ---------------------------------------------------------------------------


def test_experiments_list_uses_get() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "data": [_EXPERIMENT_RESPONSE],
        "list_metadata": {"before": None, "after": None},
    }

    page = Workflows(client=client).experiments.list(workflow_id="wf_abc123")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/wf_abc123/experiments"
    assert len(page.data) == 1
    assert page.data[0].id == "exp_abc"
    assert page.list_metadata.before is None
    assert page.list_metadata.after is None


def test_experiments_get_uses_detail_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _EXPERIMENT_RESPONSE

    Workflows(client=client).experiments.get(workflow_id="wf_abc123", experiment_id="exp_abc")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/wf_abc123/experiments/exp_abc"


def test_experiments_delete_uses_delete_method() -> None:
    client = MagicMock()
    client._prepared_request.return_value = None

    Workflows(client=client).experiments.delete(workflow_id="wf_abc123", experiment_id="exp_abc")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "DELETE"
    assert request.url == "/workflows/wf_abc123/experiments/exp_abc"


def test_experiments_duplicate_posts_to_duplicate_subroute() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {**_EXPERIMENT_RESPONSE, "id": "exp_copy"}

    Workflows(client=client).experiments.duplicate(workflow_id="wf_abc123", experiment_id="exp_abc")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows/wf_abc123/experiments/exp_abc/duplicate"


# ---------------------------------------------------------------------------
# runs — create, list, get, cancel, results, metrics
# ---------------------------------------------------------------------------


def _experiment_run_response(**overrides: object) -> dict:
    base = {
        "id": "exprun_1",
        "workflow": _WORKFLOW_REF,
        "trigger": _TRIGGER,
        "lifecycle": _PENDING,
        "timing": _TIMING,
        "experiment_id": "exp_abc",
        "block_id": "block_extract",
        "block_kind": "extract",
        "n_consensus": 5,
        "definition_fingerprint": "deadbeef",
        "documents_fingerprint": "docbeef",
        "document_count": 3,
        "total_document_count": 3,
        "completed_document_count": 0,
        "error_count": 0,
    }
    base.update(overrides)
    return base


_EXPERIMENT_RESULT = {
    "id": "expresult_1",
    "run_id": "exprun_1",
    "experiment_id": "exp_abc",
    "document_id": "expdoc_1",
    "lifecycle": _COMPLETED,
    "timing": _TIMING,
    "block_kind": "extract",
    "handle_inputs": {"input-file-0": {"type": "file"}},
    "artifact": {"operation": "extraction", "id": "ext_1"},
    "attempt": 1,
}


def test_experiments_runs_create_posts_to_run_subroute() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _experiment_run_response()

    run = Workflows(client=client).experiments.runs.create(
        workflow_id="wf_abc123",
        experiment_id="exp_abc",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows/wf_abc123/experiments/exp_abc/runs"
    assert request.data == {}
    assert run.id == "exprun_1"
    assert run.lifecycle.status == "pending"


def test_experiments_runs_create_default_body_is_empty_dict() -> None:
    """No overrides → empty body. Backend's ``RunExperimentRequest`` has
    all-optional fields, so an empty body is the canonical no-op shape."""
    client = MagicMock()
    client._prepared_request.return_value = _experiment_run_response()

    Workflows(client=client).experiments.runs.create(
        workflow_id="wf_abc123",
        experiment_id="exp_abc",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.data == {}


def test_experiments_runs_create_does_not_expose_retry_or_stale_flags() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _experiment_run_response(
        id="exprun_2",
        lifecycle=_COMPLETED,
    )

    Workflows(client=client).experiments.runs.create(
        workflow_id="wf_abc123",
        experiment_id="exp_abc",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.data == {}
    assert not hasattr(Workflows(client=client).experiments.runs, "run_document")


def test_experiments_runs_list_uses_canonical_runs_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "data": [],
        "list_metadata": {"before": None, "after": None},
    }

    page = Workflows(client=client).experiments.runs.list(workflow_id="wf_abc123", experiment_id="exp_abc")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/experiments/runs"
    assert request.params == {
        "limit": 20,
        "workflow_id": "wf_abc123",
        "experiment_id": "exp_abc",
    }
    assert page.data == []
    assert page.list_metadata.before is None
    assert page.list_metadata.after is None


def test_experiments_runs_list_exposes_workflow_run_style_filters() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "data": [],
        "list_metadata": {"before": None, "after": None},
    }

    Workflows(client=client).experiments.runs.list(
        workflow_id="wf_abc123",
        experiment_id="exp_abc",
        block_id="block_extract",
        status="completed",
        statuses=["completed", "error"],
        exclude_status="cancelled",
        trigger_type="api",
        trigger_types=["api", "manual_run"],
        from_date="2026-05-01",
        to_date="2026-05-18",
        sort_by="created_at",
        fields=["id", "lifecycle"],
        before="exprun_before",
        after="exprun_after",
        limit=10,
        order="asc",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.url == "/workflows/experiments/runs"
    assert request.params == {
        "workflow_id": "wf_abc123",
        "experiment_id": "exp_abc",
        "block_id": "block_extract",
        "status": "completed",
        "statuses": "completed,error",
        "exclude_status": "cancelled",
        "trigger_type": "api",
        "trigger_types": "api,manual_run",
        "from_date": "2026-05-01",
        "to_date": "2026-05-18",
        "sort_by": "created_at",
        "fields": "id,lifecycle",
        "before": "exprun_before",
        "after": "exprun_after",
        "limit": 10,
        "order": "asc",
    }


def test_experiments_hard_cutover_removes_legacy_content_and_metrics_aliases() -> None:
    experiments = Workflows(client=MagicMock()).experiments

    assert "cancel" not in dir(experiments)
    assert "get_metrics" not in dir(experiments)
    assert "get_content" not in dir(experiments)
    assert "get_content" not in dir(experiments.runs)
    assert "cancel_document" not in dir(experiments.runs)
    assert "get_job" not in dir(experiments.runs)
    assert "wait_for_completion" not in dir(experiments.runs)
    assert "get" in dir(experiments.runs)


def test_experiments_runs_get_uses_run_id_first_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _experiment_run_response(lifecycle=_COMPLETED)

    run = Workflows(client=client).experiments.runs.get("exprun_1")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/experiments/runs/exprun_1"
    assert run.lifecycle.status == "completed"
    assert run.timing.started_at is not None


def test_experiments_runs_cancel_uses_run_id_first_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _experiment_run_response(lifecycle={"status": "cancelled"})

    run = Workflows(client=client).experiments.runs.cancel("exprun_1")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows/experiments/runs/exprun_1/cancel"
    assert request.data == {}
    assert run.lifecycle.status == "cancelled"


def test_experiments_runs_results_list_uses_run_id_first_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "data": [_EXPERIMENT_RESULT],
        "list_metadata": {"before": None, "after": None},
    }

    page = Workflows(client=client).experiments.runs.results.list("exprun_1")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/experiments/runs/exprun_1/results"
    assert page.data[0].document_id == "expdoc_1"


def test_experiments_runs_results_get_uses_document_id_child_key() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _EXPERIMENT_RESULT

    result = Workflows(client=client).experiments.runs.results.get(
        "exprun_1",
        "expdoc_1",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/experiments/runs/exprun_1/results/expdoc_1"
    assert result.document_id == "expdoc_1"


def test_experiments_runs_metrics_summary_view_default() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "experiment_id": "exp_abc",
        "run_id": "exprun_1",
        "view": "summary",
        "definition_fingerprint": "deadbeef",
        "block_kind": "extract",
        "score": 0.83,
        "documents": [],
    }

    Workflows(client=client).experiments.runs.metrics.get("exprun_1")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/experiments/runs/exprun_1/metrics"
    assert request.params == {"view": "summary", "include_prior": True}


def test_experiments_runs_metrics_by_target_view_passes_target_path() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "run_id": "exprun_1",
        "view": "by_target",
        "target": "total",
        "score": 0.9,
        "documents": [],
    }

    Workflows(client=client).experiments.runs.metrics.get(
        "exprun_1",
        view="by_target",
        target_path="total",
        include_prior=False,
    )

    request = client._prepared_request.call_args.args[0]
    assert request.params == {
        "view": "by_target",
        "include_prior": False,
        "target_path": "total",
    }


def test_experiments_runs_metrics_returns_stale_error_envelope() -> None:
    """When the latest run is stale the backend returns the
    ``error="stale_metrics"`` envelope instead of a view payload."""
    client = MagicMock()
    client._prepared_request.return_value = {
        "error": "stale_metrics",
        "experiment_id": "exp_abc",
        "stale_reasons": ["config_changed"],
        "last_run": {
            "run_id": "exprun_1",
            "definition_fingerprint": "old",
            "score": 0.5,
            "created_at": _NOW,
        },
        "current_config_fingerprint": "new",
        "message": "Metrics are stale; rerun the experiment.",
    }

    result = Workflows(client=client).experiments.runs.metrics.get("exprun_1")

    # Discriminated union accepts the error envelope by ``error`` field.
    assert getattr(result, "error", None) == "stale_metrics"


# ---------------------------------------------------------------------------
# eligible-blocks
# ---------------------------------------------------------------------------


def test_experiments_list_eligible_blocks_uses_eligible_blocks_subroute() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {"blocks": []}

    Workflows(client=client).experiments.list_eligible_blocks("wf_abc123")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/wf_abc123/experiments/eligible-blocks"


def test_experiments_do_not_expose_run_batch_surface() -> None:
    experiments = Workflows(client=MagicMock()).experiments

    assert "prepare_run_batch" not in dir(experiments)
    assert "run_batch" not in dir(experiments)


# ---------------------------------------------------------------------------
# Diagnose — fetches entities first, then POSTs the graph
# ---------------------------------------------------------------------------


def test_workflows_diagnose_fetches_entities_then_posts_graph() -> None:
    client = MagicMock()

    entities_payload = {
        "workflow": {
            "id": "wf_abc123",
            "name": "Test",
            "description": "",
            "published": None,
            "email_trigger": {"allowed_senders": [], "allowed_domains": []},
            "created_at": _NOW,
            "updated_at": _NOW,
        },
        "blocks": [
            {
                "id": "start-1",
                "workflow_id": "wf_abc123",
                "organization_id": "org_x",
                "type": "start-document",
                "label": "Start",
                "position_x": 0,
                "position_y": 0,
                "config": None,
            },
        ],
        "edges": [],
    }
    diagnose_payload = {
        "is_valid": True,
        "issues": [
            {
                "severity": "warning",
                "code": "MISSING_HIL_PREDICATE",
                "message": "Review gate needs a predicate",
                "block_id": "extract_1",
            }
        ],
        "suggestions": [],
        "stats": {
            "total_blocks": 1,
            "total_edges": 0,
            "block_types": {"start": 1},
            "start_document_blocks": 1,
        },
    }

    # Two backend calls: one to fetch entities, one to diagnose. Order matters.
    client._prepared_request.side_effect = [entities_payload, diagnose_payload]

    result = Workflows(client=client).diagnose("wf_abc123")

    assert result.is_valid is True
    assert result.issues[0].severity == "warning"
    assert result.issues[0].code == "MISSING_HIL_PREDICATE"
    # Verify the second call was the diagnose-graph POST with the
    # unsaved-shape blocks/edges payload the backend expects.
    diagnose_call = client._prepared_request.call_args_list[1].args[0]
    assert diagnose_call.method == "POST"
    assert diagnose_call.url == "/workflows/wf_abc123/diagnose-graph"
    assert diagnose_call.data["re_propagate"] is True
    assert diagnose_call.data["blocks"][0]["id"] == "start-1"
    assert diagnose_call.data["blocks"][0]["type"] == "start"
    assert "position" in diagnose_call.data["blocks"][0]


@pytest.mark.asyncio
async def test_async_workflows_diagnose_fetches_entities_then_posts_graph() -> None:
    client = MagicMock()

    entities_payload = {
        "workflow": {
            "id": "wf_abc123",
            "name": "Test",
            "description": "",
            "published": None,
            "email_trigger": {"allowed_senders": [], "allowed_domains": []},
            "created_at": _NOW,
            "updated_at": _NOW,
        },
        "blocks": [],
        "edges": [],
    }
    diagnose_payload = {
        "is_valid": False,
        "issues": [{"severity": "error", "code": "NO_START_BLOCK", "message": "no start"}],
        "suggestions": [],
        "stats": {"total_blocks": 0, "total_edges": 0, "block_types": {}, "start_document_blocks": 0},
    }
    client._prepared_request = AsyncMock(side_effect=[entities_payload, diagnose_payload])

    result = await AsyncWorkflows(client=client).diagnose("wf_abc123")

    assert result.is_valid is False
    assert result.issues[0].code == "NO_START_BLOCK"


# ---------------------------------------------------------------------------
# Block simulate — keyed by run_id, NOT workflow_id
# ---------------------------------------------------------------------------


_SIMULATION_RESPONSE = {
    "id": "sim_1",
    "organization_id": "org_x",
    "workflow_id": "wf_abc123",
    "run_id": "wfr_1",
    "block_id": "block_extract",
    "block_type": "extract",
    "success": True,
    "handle_inputs": {"input-file-0": {"type": "file"}},
    "handle_outputs": {"output-json-0": {"type": "json", "data": {"total": 1234.56}}},
    "duration_ms": 412.0,
    "skipped": False,
}


def test_blocks_simulate_uses_runs_steps_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _SIMULATION_RESPONSE

    sim = Workflows(client=client).blocks.simulate(
        run_id="wfr_1",
        block_id="block_extract",
        n_consensus=5,
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    # Route is keyed by run_id (under /workflows/runs/...), not workflow_id —
    # exact URL match is what the regression check guards.
    assert request.url == "/workflows/runs/wfr_1/steps/block_extract/simulate"
    assert request.params == {"n_consensus": 5}
    assert sim.id == "sim_1"
    assert sim.success is True


def test_blocks_simulate_omits_check_eligibility_when_default() -> None:
    """Default ``check_eligibility=True`` should NOT appear in the query string —
    sending it would be redundant with the server default and clutter logs."""
    client = MagicMock()
    client._prepared_request.return_value = _SIMULATION_RESPONSE

    Workflows(client=client).blocks.simulate(
        run_id="wfr_1",
        block_id="block_extract",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.params is None


def test_blocks_simulate_with_check_eligibility_false_passes_param() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _SIMULATION_RESPONSE

    Workflows(client=client).blocks.simulate(
        run_id="wfr_1",
        block_id="block_extract",
        check_eligibility=False,
    )

    request = client._prepared_request.call_args.args[0]
    assert request.params == {"check_eligibility": False}


def test_blocks_simulate_with_step_id_for_for_each() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _SIMULATION_RESPONSE

    Workflows(client=client).blocks.simulate(
        run_id="wfr_1",
        block_id="block_inside_for_each",
        step_id="for_each-0/extract-0",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.params == {"step_id": "for_each-0/extract-0"}


@pytest.mark.asyncio
async def test_async_blocks_simulate_uses_runs_steps_route() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(return_value=_SIMULATION_RESPONSE)

    await AsyncWorkflows(client=client).blocks.simulate(
        run_id="wfr_1",
        block_id="block_extract",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.url == "/workflows/runs/wfr_1/steps/block_extract/simulate"
