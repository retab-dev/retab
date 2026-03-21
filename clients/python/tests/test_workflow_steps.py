from unittest.mock import AsyncMock, MagicMock

import pytest

from retab.resources.workflows.runs.steps.client import AsyncWorkflowSteps, WorkflowSteps


def test_workflow_steps_list_uses_full_steps_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = [
        {
            "run_id": "run_123",
            "organization_id": "org_123",
            "node_id": "extract-1",
            "step_id": "extract-1",
            "node_type": "extract",
            "node_label": "Extract",
            "status": "completed",
        }
    ]

    steps = WorkflowSteps(client=client).list("run_123")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/runs/run_123/steps"
    assert len(steps) == 1
    assert steps[0].node_id == "extract-1"


@pytest.mark.asyncio
async def test_async_workflow_steps_list_uses_full_steps_route() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(return_value=[
        {
            "run_id": "run_123",
            "organization_id": "org_123",
            "node_id": "extract-1",
            "step_id": "extract-1",
            "node_type": "extract",
            "node_label": "Extract",
            "status": "completed",
        }
    ])

    steps = await AsyncWorkflowSteps(client=client).list("run_123")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/runs/run_123/steps"
    assert len(steps) == 1
    assert steps[0].node_id == "extract-1"


def test_workflow_steps_get_handle_outputs_typed() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "node_id": "extract-1",
        "node_type": "extract",
        "node_label": "Extract",
        "status": "completed",
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
    # handle_outputs values are now HandlePayload objects
    payload = step.handle_outputs["output-json-0"]
    assert payload.type == "json"
    assert payload.data == {"invoice_number": "INV-001"}
    # Convenience accessor
    assert step.extracted_data == {"invoice_number": "INV-001"}
    assert step.get_json_output("output-json-0") == {"invoice_number": "INV-001"}


@pytest.mark.asyncio
async def test_async_workflow_steps_get_handle_outputs_typed() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(return_value={
        "node_id": "start-json-1",
        "node_type": "start_json",
        "node_label": "Start JSON",
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


def test_workflow_steps_list_with_node_ids() -> None:
    """list() with node_ids filters the persisted step list and still returns WorkflowRunStep items."""
    client = MagicMock()
    client._prepared_request.return_value = [
        {
            "run_id": "run_123",
            "organization_id": "org_123",
            "node_id": "extract-1",
            "step_id": "extract-1",
            "node_type": "extract",
            "node_label": "Extract",
            "status": "completed",
        },
        {
            "run_id": "run_123",
            "organization_id": "org_123",
            "node_id": "parse-1",
            "step_id": "parse-1",
            "node_type": "parse",
            "node_label": "Parse",
            "status": "completed",
        },
    ]

    result = WorkflowSteps(client=client).list("run_123", node_ids=["extract-1"])

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/runs/run_123/steps"
    assert len(result) == 1
    assert result[0].node_id == "extract-1"


def test_workflow_steps_get_many_uses_batch_endpoint() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "outputs": {
            "extract-1": {
                "node_id": "extract-1",
                "node_type": "extract",
                "node_label": "Extract",
                "status": "completed",
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
    assert "extract-1" in result.outputs
    assert result.outputs["extract-1"].extracted_data == {"field": "value"}


def test_workflow_steps_get_no_json_output() -> None:
    """extracted_data returns None when there are no JSON outputs."""
    client = MagicMock()
    client._prepared_request.return_value = {
        "node_id": "parse-1",
        "node_type": "parse",
        "node_label": "Parse",
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
        "node_id": "start-1",
        "node_type": "start",
        "node_label": "Start",
        "status": "completed",
        "handle_outputs": None,
        "handle_inputs": None,
    }

    step = WorkflowSteps(client=client).get("run_123", "start-1")
    assert step.extracted_data is None
