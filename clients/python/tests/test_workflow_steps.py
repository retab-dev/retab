from unittest.mock import AsyncMock, MagicMock

import pytest
from pydantic import ValidationError

from retab.resources.workflows.runs.steps.client import AsyncWorkflowSteps, WorkflowSteps
from retab.types.workflows.model import (
    ExtractStepOutput,
    ForEachSentinelStartStepOutput,
    parse_workflow_step_output,
)


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
    # handle_outputs values are now HandlePayload objects
    payload = step.handle_outputs["output-json-0"]
    assert payload.type == "json"
    assert payload.data == {"invoice_number": "INV-001"}
    # Convenience accessor
    assert step.extracted_data == {"invoice_number": "INV-001"}
    assert step.get_json_output("output-json-0") == {"invoice_number": "INV-001"}


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
        "outputs": {
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
    assert "extract-1" in result.outputs
    assert result.outputs["extract-1"].artifact is not None
    assert result.outputs["extract-1"].artifact.id == "ext_789"
    assert result.outputs["extract-1"].extracted_data == {"field": "value"}


def test_workflow_steps_get_no_json_output() -> None:
    """extracted_data returns None when there are no JSON outputs."""
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


def test_parse_workflow_step_output_extract_accepts_canonical_shape() -> None:
    parsed = parse_workflow_step_output(
        "extract",
        {
            "output": {"invoice_number": "INV-001"},
            "consensus": {
                "choices": [{"invoice_number": "INV-001"}],
                "likelihoods": {"invoice_number": 0.98},
            },
            "extraction_id": "ext_123",
        },
    )

    assert isinstance(parsed, ExtractStepOutput)
    assert parsed.output == {"invoice_number": "INV-001"}
    assert parsed.extracted_data == {"invoice_number": "INV-001"}
    assert parsed.consensus.likelihoods == {"invoice_number": 0.98}
    assert parsed.likelihoods == {"invoice_number": 0.98}
    assert parsed.consensus_details == [{"data": {"invoice_number": "INV-001"}}]


def test_parse_workflow_step_output_extract_rejects_legacy_shape() -> None:
    legacy_payload = {
        "extracted_data": {"invoice_number": "INV-002"},
        "likelihoods": {"invoice_number": 0.91},
        "consensus_details": [{"data": {"invoice_number": "INV-002"}}],
        "extraction_id": "ext_456",
    }

    with pytest.raises(ValidationError):
        ExtractStepOutput.model_validate(legacy_payload)

    parsed = parse_workflow_step_output("extract", legacy_payload)
    assert parsed == legacy_payload


def test_parse_workflow_step_output_for_each_partition_payload() -> None:
    parsed = parse_workflow_step_output(
        "for_each_sentinel_start",
        {
            "message": "Splitting document",
            "mr_id": "for_each-1",
            "current_index": 0,
            "total_items": 1,
            "max_iterations": 1,
            "is_first_iteration": True,
            "map_method": "split_by_key",
            "partition_id": "prtn_123",
            "all_item_keys": ["invoice_1"],
        },
    )

    assert isinstance(parsed, ForEachSentinelStartStepOutput)
    assert parsed.partition_id == "prtn_123"
