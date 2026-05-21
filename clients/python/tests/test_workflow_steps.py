import datetime
from unittest.mock import AsyncMock, MagicMock

import pytest

from retab.resources.workflows.runs.steps.client import AsyncWorkflowSteps, WorkflowSteps
from retab.resources.workflows.artifacts.client import AsyncWorkflowArtifacts, WorkflowArtifacts
from retab.types.workflows import model as workflow_model
from retab.types.workflows.model import StepExecutionResponse, WorkflowRun


def test_workflow_steps_list_uses_full_steps_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "data": [
            {
                "run_id": "run_123",
                "organization_id": "org_123",
                "block_id": "extract-1",
                "step_id": "extract-1",
                "block_type": "extract",
                "block_label": "Extract",
                "lifecycle": {"status": "completed"},
            }
        ],
        "list_metadata": {"before": None, "after": None},
    }

    steps = WorkflowSteps(client=client).list("run_123")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/steps?run_id=run_123"
    assert len(steps) == 1
    assert steps[0].block_id == "extract-1"
    assert steps.list_metadata.before is None
    assert steps.list_metadata.after is None


def test_workflow_artifacts_get_accepts_ref_and_returns_flattened_record() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "operation": "review_trigger_evaluation",
        "id": "heval_123",
        "requires_human_review": True,
    }
    artifact_ref = workflow_model.StepArtifactRef(
        operation="review_trigger_evaluation",
        id="heval_123",
    )

    artifact = WorkflowArtifacts(client=client).get(artifact_ref)

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/artifacts/review_trigger_evaluation/heval_123"
    assert artifact.operation == "review_trigger_evaluation"
    assert artifact.id == "heval_123"
    assert artifact.model_dump() == {"operation": "review_trigger_evaluation", "id": "heval_123"}


def test_workflow_artifacts_list_uses_run_scoped_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "data": [
            {"operation": "conditional_evaluation", "id": "ceval_123"},
        ],
        "list_metadata": {"before": None, "after": None},
    }

    artifacts = WorkflowArtifacts(client=client).list(
        "run_123",
        operation="conditional_evaluation",
        block_id="conditional-1",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/artifacts"
    assert request.params == {
        "run_id": "run_123",
        "operation": "conditional_evaluation",
        "block_id": "conditional-1",
    }
    assert len(artifacts) == 1
    assert artifacts[0].operation == "conditional_evaluation"
    assert artifacts.list_metadata.before is None
    assert artifacts.list_metadata.after is None


@pytest.mark.asyncio
async def test_async_workflow_steps_list_uses_full_steps_route() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(
        return_value={
            "data": [
                {
                    "run_id": "run_123",
                    "organization_id": "org_123",
                    "block_id": "extract-1",
                    "step_id": "extract-1",
                    "block_type": "extract",
                    "block_label": "Extract",
                    "lifecycle": {"status": "completed"},
                }
            ],
            "list_metadata": {"before": None, "after": None},
        }
    )

    steps = await AsyncWorkflowSteps(client=client).list("run_123")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/steps?run_id=run_123"
    assert len(steps) == 1
    assert steps[0].block_id == "extract-1"
    assert steps.list_metadata.before is None
    assert steps.list_metadata.after is None


@pytest.mark.asyncio
async def test_async_workflow_artifacts_get_accepts_operation_and_id() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(
        return_value={
            "operation": "function_invocation",
            "id": "fninv_123",
            "output": {"ok": True},
        }
    )

    artifact = await AsyncWorkflowArtifacts(client=client).get(
        "function_invocation",
        "fninv_123",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/artifacts/function_invocation/fninv_123"
    assert artifact.operation == "function_invocation"
    assert artifact.model_dump() == {"operation": "function_invocation", "id": "fninv_123"}


def test_workflow_steps_get_handle_outputs_typed() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "block_id": "extract-1",
        "block_type": "extract",
        "block_label": "Extract",
        "lifecycle": {"status": "completed"},
        "artifact": {
            "operation": "extraction",
            "id": "ext_123",
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
    assert request.url == "/workflows/steps/extract-1?run_id=run_123"
    assert step.handle_outputs is not None
    assert step.artifact is not None
    assert step.artifact.model_dump() == {"operation": "extraction", "id": "ext_123"}
    assert "output" not in step.model_dump()
    # The plural artifacts list and API-specific artifact_view are gone.
    assert "artifacts" not in step.model_dump()
    assert "artifact_view" not in step.model_dump()
    assert "metadata" not in step.model_dump()
    # handle_outputs values are now HandlePayload objects
    payload = step.handle_outputs["output-json-0"]
    assert payload.type == "json"
    assert payload.data == {"invoice_number": "INV-001"}
    # Convenience accessor
    assert step.extracted_data == {"invoice_number": "INV-001"}
    assert step.get_json_output("output-json-0") == {"invoice_number": "INV-001"}


def test_workflow_steps_get_accepts_json_ref_handle_payload() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "run_id": "run_123",
        "organization_id": "org_123",
        "block_id": "function-1",
        "step_id": "function-1",
        "block_type": "function",
        "block_label": "Function",
        "lifecycle": {"status": "completed"},
        "handle_outputs": {
            "output-json-0": {
                "type": "json_ref",
                "artifact_ref": {
                    "operation": "workflow_step_json",
                    "id": "artifact_123",
                    "key": "output-json-0",
                },
                "preview": {"truncated": True},
            },
        },
        "handle_inputs": {},
    }

    step = WorkflowSteps(client=client).get("run_123", "function-1")

    payload = step.handle_outputs["output-json-0"]
    assert payload.type == "json_ref"
    assert payload.artifact_ref == {
        "operation": "workflow_step_json",
        "id": "artifact_123",
        "key": "output-json-0",
    }
    assert payload.preview == {"truncated": True}


def test_workflow_step_sdk_does_not_export_removed_payload_response_names() -> None:
    response_name = "Step" + "Output" + "Response"
    batch_response_name = "Step" + "Outputs" + "BatchResponse"

    assert not hasattr(workflow_model, response_name)
    assert not hasattr(workflow_model, batch_response_name)
    assert not hasattr(workflow_model, "StepExecutionsBatchResponse")
    assert not hasattr(workflow_model, "StepExecutionStatus")
    assert not hasattr(workflow_model, "TerminalState")


def test_workflow_steps_get_accepts_partition_artifact() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "block_id": "for_each-1",
        "block_type": "for_each",
        "block_label": "For Each",
        "lifecycle": {"status": "completed"},
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
    client._prepared_request = AsyncMock(
        return_value={
            "block_id": "start-json-1",
            "block_type": "start_json",
            "block_label": "Start JSON",
            "lifecycle": {"status": "completed"},
            "handle_outputs": {
                "output-json-0": {
                    "type": "json",
                    "data": {"payload": {"ok": True}},
                },
            },
            "handle_inputs": {},
        }
    )

    step = await AsyncWorkflowSteps(client=client).get("run_123", "start-json-1")

    assert step.handle_outputs is not None
    payload = step.handle_outputs["output-json-0"]
    assert payload.type == "json"
    assert payload.data == {"payload": {"ok": True}}
    assert step.extracted_data == {"payload": {"ok": True}}


def test_workflow_steps_list_with_block_ids() -> None:
    """list() with block_ids filters the persisted step list and still returns WorkflowRunStep items."""
    client = MagicMock()
    client._prepared_request.return_value = {
        "data": [
            {
                "run_id": "run_123",
                "organization_id": "org_123",
                "block_id": "extract-1",
                "step_id": "extract-1",
                "block_type": "extract",
                "block_label": "Extract",
                "lifecycle": {"status": "completed"},
                "artifact": {"operation": "extraction", "id": "ext_123"},
            },
            {
                "run_id": "run_123",
                "organization_id": "org_123",
                "block_id": "parse-1",
                "step_id": "parse-1",
                "block_type": "parse",
                "block_label": "Parse",
                "lifecycle": {"status": "completed"},
            },
        ],
        "list_metadata": {"before": None, "after": None},
    }

    result = WorkflowSteps(client=client).list("run_123", block_ids=["extract-1"])

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/steps?run_id=run_123"
    assert len(result) == 1
    assert result[0].block_id == "extract-1"
    assert result[0].artifact is not None
    assert result[0].artifact.operation == "extraction"


def test_workflow_steps_get_requires_block_id() -> None:
    client = MagicMock()
    steps = WorkflowSteps(client=client)

    with pytest.raises(TypeError):
        steps.get("run_123")  # type: ignore[call-arg]
    with pytest.raises(TypeError, match="block_id is required"):
        steps.get("run_123", "")  # type: ignore[arg-type]

    client._prepared_request.assert_not_called()


def test_workflow_steps_only_exposes_get_for_full_execution_fetches() -> None:
    steps = WorkflowSteps(client=MagicMock())
    assert not hasattr(steps, "get_all")
    assert not hasattr(steps, "get_many")
    assert not hasattr(steps, "getAll")
    assert not hasattr(steps, "getMany")


def test_workflow_steps_get_no_json_output() -> None:
    """extracted_data returns None when there are no JSON handle outputs."""
    client = MagicMock()
    client._prepared_request.return_value = {
        "block_id": "parse-1",
        "block_type": "parse",
        "block_label": "Parse",
        "lifecycle": {"status": "completed"},
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
        "block_type": "start-document",
        "block_label": "Start",
        "lifecycle": {"status": "completed"},
        "handle_outputs": None,
        "handle_inputs": None,
    }

    step = WorkflowSteps(client=client).get("run_123", "start-1")
    assert step.extracted_data is None


def test_step_execution_response_uses_lifecycle_for_failed_step() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "block_id": "extract-1",
        "block_type": "extract",
        "block_label": "Extract",
        "lifecycle": {
            "status": "error",
            "message": "LLM returned malformed JSON",
            "stage": "execution",
        },
        "artifact": None,
        "handle_outputs": None,
        "handle_inputs": None,
    }

    step = WorkflowSteps(client=client).get("run_123", "extract-1")

    assert step.lifecycle.status == "error"
    assert step.lifecycle.message == "LLM returned malformed JSON"
    assert "error" not in step.model_dump()
    assert "status" not in step.model_dump()
    assert "terminal" not in step.model_dump()


def test_step_execution_response_has_no_compatibility_error_field() -> None:
    response = StepExecutionResponse.model_validate(
        {
            "block_id": "extract-1",
            "block_type": "extract",
            "block_label": "Extract",
            "lifecycle": {"status": "completed"},
        }
    )
    assert "error" not in response.model_dump()
    assert "status" not in response.model_dump()
    assert "terminal" not in response.model_dump()


def _minimal_run_payload(**overrides) -> dict:
    """Build a minimal v2 ``WorkflowRun`` payload.

    ``overrides`` may set ``lifecycle`` directly; otherwise pass
    ``lifecycle_kind`` (and any extra lifecycle fields like ``message``)
    as a convenience.
    """
    now = datetime.datetime(2026, 1, 1, tzinfo=datetime.timezone.utc).isoformat()
    lifecycle_kind = overrides.pop("lifecycle_kind", "completed")
    lifecycle: dict = {"status": lifecycle_kind}
    if lifecycle_kind == "error":
        lifecycle.setdefault("message", overrides.pop("error_message", "boom"))
    payload: dict = {
        "id": "run_123",
        "organization_id": "org_1",
        "workflow": {
            "workflow_id": "wf_1",
            "version_id": "wv_1",
            "name_at_run_time": "Test",
        },
        "trigger": {"type": "manual"},
        "lifecycle": overrides.pop("lifecycle", lifecycle),
        "timing": {"created_at": now, "started_at": now, "completed_at": now},
        "inputs": {"documents": {}, "json_data": {}},
    }
    payload.update(overrides)
    return payload
