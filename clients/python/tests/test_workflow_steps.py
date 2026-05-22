import datetime
from unittest.mock import AsyncMock, MagicMock

import pytest

from retab.resources.workflows.artifacts import AsyncWorkflowArtifacts, WorkflowArtifacts
from retab.resources.workflows import AsyncWorkflows, Workflows
from retab.resources.workflows.runs import AsyncWorkflowRuns, WorkflowRuns
from retab.resources.workflows.steps import AsyncWorkflowSteps, WorkflowSteps
from retab.types.workflows import model as workflow_model
from retab.types.workflows.model import StepExecutionResponse, WorkflowRun


def _workflow_run_step_payload(**overrides) -> dict:
    payload = {
        "step_id": "step_extract_1",
        "run_id": "run_123",
        "block_id": "extract-1",
        "block_type": "extract",
        "block_label": "Extract",
        "lifecycle": {"status": "completed"},
        "handle_inputs": {},
        "handle_outputs": {},
    }
    payload.update(overrides)
    return payload


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


def test_workflow_steps_are_exposed_on_workflows_not_runs() -> None:
    client = MagicMock()

    workflows = Workflows(client=client)
    runs = WorkflowRuns(client=client)

    assert isinstance(workflows.steps, WorkflowSteps)
    with pytest.raises(AttributeError):
        object.__getattribute__(runs, "steps")


def test_async_workflow_steps_are_exposed_on_workflows_not_runs() -> None:
    client = MagicMock()

    workflows = AsyncWorkflows(client=client)
    runs = AsyncWorkflowRuns(client=client)

    assert isinstance(workflows.steps, AsyncWorkflowSteps)
    with pytest.raises(AttributeError):
        object.__getattribute__(runs, "steps")


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


def test_workflow_artifacts_get_uses_flat_artifact_id_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "operation": "extraction",
        "id": "ext_123",
    }

    artifact = WorkflowArtifacts(client=client).get("ext_123")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/artifacts/ext_123"
    assert artifact.id == "ext_123"
    assert artifact.operation == "extraction"


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
async def test_async_workflow_artifacts_get_uses_flat_artifact_id_route() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(
        return_value={
            "operation": "classification",
            "id": "clss_123",
        }
    )

    artifact = await AsyncWorkflowArtifacts(client=client).get("clss_123")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/artifacts/clss_123"
    assert artifact.id == "clss_123"


def test_workflow_steps_get_handle_outputs_typed() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _workflow_run_step_payload(
        handle_outputs={
            "output-json-0": {
                "type": "json",
                "data": {"invoice_number": "INV-001"},
            },
        },
    )

    step = WorkflowSteps(client=client).get("step_extract_1")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/steps/step_extract_1"
    assert step.step_id == "step_extract_1"
    assert step.run_id == "run_123"
    assert "output" not in step.model_dump()
    assert step.handle_outputs is not None
    payload = step.handle_outputs["output-json-0"]
    assert payload.type == "json"
    assert payload.data == {"invoice_number": "INV-001"}


def test_workflow_steps_get_accepts_json_ref_handle_payload() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _workflow_run_step_payload(
        step_id="step_function_1",
        block_id="function-1",
        block_type="function",
        block_label="Function",
        handle_outputs={
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
    )

    step = WorkflowSteps(client=client).get("step_function_1")

    assert step.handle_outputs is not None
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


def test_workflow_steps_get_accepts_loop_container_fields() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _workflow_run_step_payload(
        step_id="step_for_each_1",
        block_id="for_each-1",
        block_type="for_each",
        block_label="For each",
        loop_containers=[
            {
                "container_id": "for_each-1",
                "iteration": 2,
                "is_parallel": True,
                "parallel_item_index": 2,
            }
        ],
    )

    step = WorkflowSteps(client=client).get("step_for_each_1")

    assert step.loop_containers[0].iteration == 2
    assert step.loop_containers[0].is_parallel is True


@pytest.mark.asyncio
async def test_async_workflow_steps_get_handle_outputs_typed() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(
        return_value=_workflow_run_step_payload(
            step_id="step_start_json_1",
            block_id="start-json-1",
            block_type="start_json",
            block_label="Start JSON",
            handle_outputs={
                "output-json-0": {
                    "type": "json",
                    "data": {"payload": {"ok": True}},
                },
            },
        )
    )

    step = await AsyncWorkflowSteps(client=client).get("step_start_json_1")

    assert step.handle_outputs is not None
    payload = step.handle_outputs["output-json-0"]
    assert payload.type == "json"
    assert payload.data == {"payload": {"ok": True}}


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


def test_workflow_steps_get_requires_step_id() -> None:
    client = MagicMock()
    steps = WorkflowSteps(client=client)

    with pytest.raises(TypeError):
        steps.get()  # type: ignore[call-arg]
    with pytest.raises(TypeError, match="step_id is required"):
        steps.get("")  # type: ignore[arg-type]

    client._prepared_request.assert_not_called()


def test_workflow_steps_does_not_expose_removed_query_fetches() -> None:
    steps = WorkflowSteps(client=MagicMock())
    assert not hasattr(steps, "query")
    assert not hasattr(steps, "get_all")
    assert not hasattr(steps, "get_many")
    assert not hasattr(steps, "getAll")
    assert not hasattr(steps, "getMany")


def test_workflow_steps_get_accepts_non_json_handle_output() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _workflow_run_step_payload(
        step_id="step_parse_1",
        block_id="parse-1",
        block_type="parse",
        block_label="Parse",
        handle_outputs={
            "output-file-0": {
                "type": "file",
                "document": {"id": "file_123", "filename": "doc.pdf", "content": "...", "mime_type": "application/pdf"},
            },
        },
    )

    step = WorkflowSteps(client=client).get("step_parse_1")
    assert step.handle_outputs is not None
    assert step.handle_outputs["output-file-0"].type == "file"


def test_workflow_steps_get_accepts_empty_handle_outputs() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _workflow_run_step_payload(
        step_id="step_start_1",
        block_id="start-1",
        block_type="start_document",
        block_label="Start",
    )

    step = WorkflowSteps(client=client).get("step_start_1")
    assert step.handle_outputs == {}


def test_workflow_steps_get_uses_lifecycle_for_failed_step() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _workflow_run_step_payload(
        step_id="step_extract_1",
        block_id="extract-1",
        block_type="extract",
        block_label="Extract",
        lifecycle={"status": "error", "message": "boom"},
    )

    step = WorkflowSteps(client=client).get("step_extract_1")

    assert step.lifecycle.status == "error"
    assert "status" not in step.model_dump()
    assert "error" not in step.model_dump()
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
