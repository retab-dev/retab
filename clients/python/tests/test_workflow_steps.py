import datetime
from unittest.mock import AsyncMock, MagicMock

import pytest

from retab.resources.workflows.runs.steps.client import AsyncWorkflowSteps, WorkflowSteps
from retab.types.workflows import model as workflow_model
from retab.types.workflows.model import StepExecutionResponse, WorkflowRun


def test_workflow_steps_list_uses_full_steps_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = [
        {
            "run_id": "run_123",
            "organization_id": "org_123",
            "block_id": "extract-1",
            "step_id": "extract-1",
            "block_type": "extract",
            "block_label": "Extract",
            "status": "completed",
        }
    ]

    steps = WorkflowSteps(client=client).list("run_123")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/runs/run_123/steps"
    assert len(steps) == 1
    assert steps[0].block_id == "extract-1"


@pytest.mark.asyncio
async def test_async_workflow_steps_list_uses_full_steps_route() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(return_value=[
        {
            "run_id": "run_123",
            "organization_id": "org_123",
            "block_id": "extract-1",
            "step_id": "extract-1",
            "block_type": "extract",
            "block_label": "Extract",
            "status": "completed",
        }
    ])

    steps = await AsyncWorkflowSteps(client=client).list("run_123")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/runs/run_123/steps"
    assert len(steps) == 1
    assert steps[0].block_id == "extract-1"


def test_workflow_steps_get_handle_outputs_typed() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "block_id": "extract-1",
        "block_type": "extract",
        "block_label": "Extract",
        "status": "completed",
        "artifact": {
            "operation": "extraction",
            "id": "ext_123",
        },
        "artifacts": [
            {
                "operation": "extraction",
                "id": "ext_123",
            },
            {
                "operation": "extract",
                "id": "run_123_extract-1",
            },
        ],
        "artifact_view": {
            "block_type": "extract",
            "artifact": {
                "operation": "extraction",
                "id": "ext_123",
            },
            "artifacts": [
                {
                    "operation": "extraction",
                    "id": "ext_123",
                },
                {
                    "operation": "extract",
                    "id": "run_123_extract-1",
                },
            ],
            "data": {
                "output": {"invoice_number": "INV-001"},
                "extraction_id": "ext_123",
            },
        },
        "output": {"removed": True},
        "handle_outputs": {
            "output-json-0": {
                "type": "json",
                "data": {"invoice_number": "INV-001"},
            },
        },
        "handle_inputs": None,
    }

    step = WorkflowSteps(client=client).get("run_123", "extract-1")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/runs/run_123/steps/extract-1"
    assert step.handle_outputs is not None
    assert step.artifact is not None
    assert step.artifact.operation == "extraction"
    assert step.artifact.id == "ext_123"
    assert [artifact.model_dump() for artifact in step.artifacts] == [
        {"operation": "extraction", "id": "ext_123"},
        {"operation": "extract", "id": "run_123_extract-1"},
    ]
    assert step.artifact_view is not None
    assert step.artifact_view.data == {
        "output": {"invoice_number": "INV-001"},
        "extraction_id": "ext_123",
    }
    assert "output" not in step.model_dump()
    # handle_outputs values are now HandlePayload objects
    payload = step.handle_outputs["output-json-0"]
    assert payload.type == "json"
    assert payload.data == {"invoice_number": "INV-001"}
    # Convenience accessor
    assert step.extracted_data == {"invoice_number": "INV-001"}
    assert step.get_json_output("output-json-0") == {"invoice_number": "INV-001"}


def test_workflow_step_sdk_does_not_export_removed_payload_response_names() -> None:
    response_name = "Step" + "Output" + "Response"
    batch_response_name = "Step" + "Outputs" + "BatchResponse"

    assert not hasattr(workflow_model, response_name)
    assert not hasattr(workflow_model, batch_response_name)


def test_workflow_steps_get_accepts_partition_artifact() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "block_id": "for_each-1",
        "block_type": "for_each",
        "block_label": "For Each",
        "status": "completed",
        "artifact": {
            "operation": "partition",
            "id": "prtn_123",
        },
        "handle_outputs": None,
        "handle_inputs": None,
    }

    step = WorkflowSteps(client=client).get("run_123", "for_each-1")

    assert step.artifact is not None
    assert step.artifact.operation == "partition"
    assert step.artifact.id == "prtn_123"


@pytest.mark.asyncio
async def test_async_workflow_steps_get_handle_outputs_typed() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(return_value={
        "block_id": "start-json-1",
        "block_type": "start_json",
        "block_label": "Start JSON",
        "status": "completed",
        "handle_outputs": {
            "output-json-0": {
                "type": "json",
                "data": {"payload": {"ok": True}},
            },
        },
        "handle_inputs": {},
    })

    step = await AsyncWorkflowSteps(client=client).get("run_123", "start-json-1")

    assert step.handle_outputs is not None
    payload = step.handle_outputs["output-json-0"]
    assert payload.type == "json"
    assert payload.data == {"payload": {"ok": True}}
    assert step.extracted_data == {"payload": {"ok": True}}


def test_workflow_steps_list_with_block_ids() -> None:
    """list() with block_ids filters the persisted step list and still returns WorkflowRunStep items."""
    client = MagicMock()
    client._prepared_request.return_value = [
        {
            "run_id": "run_123",
            "organization_id": "org_123",
            "block_id": "extract-1",
            "step_id": "extract-1",
            "block_type": "extract",
            "block_label": "Extract",
            "status": "completed",
            "artifact": {"operation": "extraction", "id": "ext_123"},
        },
        {
            "run_id": "run_123",
            "organization_id": "org_123",
            "block_id": "parse-1",
            "step_id": "parse-1",
            "block_type": "parse",
            "block_label": "Parse",
            "status": "completed",
        },
    ]

    result = WorkflowSteps(client=client).list("run_123", block_ids=["extract-1"])

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/runs/run_123/steps"
    assert len(result) == 1
    assert result[0].block_id == "extract-1"
    assert result[0].artifact is not None
    assert result[0].artifact.operation == "extraction"


def test_workflow_steps_get_many_uses_batch_endpoint() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "executions": {
            "extract-1": {
                "block_id": "extract-1",
                "block_type": "extract",
                "block_label": "Extract",
                "status": "completed",
                "artifact": {
                    "operation": "extraction",
                    "id": "ext_789",
                },
                "handle_outputs": {
                    "output-json-0": {
                        "type": "json",
                        "data": {"field": "value"},
                    },
                },
                "handle_inputs": None,
            },
        },
    }

    result = WorkflowSteps(client=client).get_many("run_123", ["extract-1"])

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows/runs/run_123/steps/batch"
    assert "extract-1" in result.executions
    assert result.executions["extract-1"].artifact is not None
    assert result.executions["extract-1"].artifact.id == "ext_789"
    assert result.executions["extract-1"].extracted_data == {"field": "value"}


def test_workflow_steps_get_no_json_output() -> None:
    """extracted_data returns None when there are no JSON handle outputs."""
    client = MagicMock()
    client._prepared_request.return_value = {
        "block_id": "parse-1",
        "block_type": "parse",
        "block_label": "Parse",
        "status": "completed",
        "handle_outputs": {
            "output-file-0": {
                "type": "file",
                "document": {"id": "file_123", "filename": "doc.pdf", "content": "...", "mime_type": "application/pdf"},
            },
        },
        "handle_inputs": None,
    }

    step = WorkflowSteps(client=client).get("run_123", "parse-1")
    assert step.extracted_data is None
    assert step.get_json_output("output-file-0") is None


def test_workflow_steps_get_empty_handle_outputs() -> None:
    """extracted_data returns None when handle_outputs is None."""
    client = MagicMock()
    client._prepared_request.return_value = {
        "block_id": "start-1",
        "block_type": "start",
        "block_label": "Start",
        "status": "completed",
        "handle_outputs": None,
        "handle_inputs": None,
    }

    step = WorkflowSteps(client=client).get("run_123", "start-1")
    assert step.extracted_data is None


def test_step_execution_response_carries_error_on_failed_step() -> None:
    """StepExecutionResponse now carries `error` for parity with WorkflowRunStep / StepStatus."""
    client = MagicMock()
    client._prepared_request.return_value = {
        "block_id": "extract-1",
        "block_type": "extract",
        "block_label": "Extract",
        "status": "error",
        "error": "LLM returned malformed JSON",
        "artifact": None,
        "handle_outputs": None,
        "handle_inputs": None,
    }

    step = WorkflowSteps(client=client).get("run_123", "extract-1")

    assert step.status == "error"
    assert step.error == "LLM returned malformed JSON"


def test_step_execution_response_error_defaults_to_none_for_successful_step() -> None:
    response = StepExecutionResponse.model_validate(
        {
            "block_id": "extract-1",
            "block_type": "extract",
            "block_label": "Extract",
            "status": "completed",
        }
    )
    assert response.error is None


def _minimal_run_payload(**overrides) -> dict:
    now = datetime.datetime(2026, 1, 1, tzinfo=datetime.timezone.utc)
    payload = {
        "id": "run_123",
        "workflow_id": "wf_1",
        "workflow_name": "Test",
        "organization_id": "org_1",
        "status": "completed",
        "started_at": now,
        "created_at": now,
        "updated_at": now,
        "waiting_for_block_ids": [],
    }
    payload.update(overrides)
    return payload


def test_workflow_run_is_terminal_true_for_each_terminal_status() -> None:
    for status in ("completed", "error", "cancelled"):
        run = WorkflowRun.model_validate(_minimal_run_payload(status=status))
        assert run.is_terminal, f"{status} should be terminal"


def test_workflow_run_is_terminal_false_for_non_terminal_statuses() -> None:
    for status in ("pending", "running", "waiting_for_human"):
        run = WorkflowRun.model_validate(_minimal_run_payload(status=status))
        assert not run.is_terminal, f"{status} should not be terminal"
