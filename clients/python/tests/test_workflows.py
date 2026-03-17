from unittest.mock import AsyncMock, MagicMock

import pytest

from retab.resources.workflows.client import AsyncWorkflows, Workflows
from retab.types.workflows.model import (
    WorkflowRun,
    WorkflowRunError,
    WorkflowBlock,
    WorkflowEdge,
    WorkflowSubflow,
    WorkflowWithEntities,
    Workflow,
    FinalNodeOutput,
)


def test_workflows_get_uses_detail_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "id": "workflow_123",
        "name": "Test Workflow",
        "description": "",
        "is_published": False,
        "email_senders_whitelist": [],
        "email_domains_whitelist": [],
        "created_at": "2026-03-12T10:00:00Z",
        "updated_at": "2026-03-12T10:00:00Z",
    }

    workflow = Workflows(client=client).get("workflow_123")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/workflow_123"
    assert workflow.id == "workflow_123"


def test_workflows_list_uses_paginated_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "data": [],
        "list_metadata": {"before": None, "after": "cursor_1"},
    }

    result = Workflows(client=client).list(
        limit=5,
        order="asc",
        sort_by="updated_at",
        fields="id,name",
        after="cursor_0",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows"
    assert request.params == {
        "limit": 5,
        "order": "asc",
        "sort_by": "updated_at",
        "fields": "id,name",
        "after": "cursor_0",
    }
    assert result.list_metadata.after == "cursor_1"


@pytest.mark.asyncio
async def test_async_workflows_get_uses_detail_route() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(return_value={
        "id": "workflow_123",
        "name": "Test Workflow",
        "description": "",
        "is_published": False,
        "email_senders_whitelist": [],
        "email_domains_whitelist": [],
        "created_at": "2026-03-12T10:00:00Z",
        "updated_at": "2026-03-12T10:00:00Z",
    })

    workflow = await AsyncWorkflows(client=client).get("workflow_123")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/workflow_123"
    assert workflow.id == "workflow_123"


@pytest.mark.asyncio
async def test_async_workflows_list_uses_paginated_route() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(return_value={
        "data": [],
        "list_metadata": {"before": None, "after": "cursor_1"},
    })

    result = await AsyncWorkflows(client=client).list(
        limit=5,
        order="asc",
        sort_by="updated_at",
        fields="id,name",
        after="cursor_0",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows"
    assert request.params == {
        "limit": 5,
        "order": "asc",
        "sort_by": "updated_at",
        "fields": "id,name",
        "after": "cursor_0",
    }
    assert result.list_metadata.after == "cursor_1"


def test_workflow_run_accepts_newer_step_node_types() -> None:
    run = WorkflowRun.model_validate(
        {
            "id": "run_123",
            "workflow_id": "workflow_123",
            "workflow_name": "Classifier Workflow",
            "organization_id": "org_123",
            "status": "running",
            "started_at": "2026-03-13T10:00:00Z",
            "steps": [
                {
                    "node_id": "classifier-1",
                    "node_type": "classifier",
                    "node_label": "Classifier",
                    "status": "completed",
                }
            ],
            "created_at": "2026-03-13T10:00:00Z",
            "updated_at": "2026-03-13T10:00:00Z",
        }
    )

    assert run.steps[0].node_type == "classifier"


def test_workflow_run_accepts_skipped_step_statuses() -> None:
    run = WorkflowRun.model_validate(
        {
            "id": "run_456",
            "workflow_id": "workflow_123",
            "workflow_name": "Branching Workflow",
            "organization_id": "org_123",
            "status": "running",
            "started_at": "2026-03-13T10:00:00Z",
            "steps": [
                {
                    "node_id": "extract-1",
                    "node_type": "extract",
                    "node_label": "Extract",
                    "status": "completed",
                },
                {
                    "node_id": "extract-2",
                    "node_type": "extract",
                    "node_label": "Skipped branch",
                    "status": "skipped",
                },
            ],
            "created_at": "2026-03-13T10:00:00Z",
            "updated_at": "2026-03-13T10:00:00Z",
        }
    )

    assert [step.status for step in run.steps] == ["completed", "skipped"]


def test_workflow_run_enriched_fields() -> None:
    """New fields on WorkflowRun are parsed when present and default when absent."""
    run = WorkflowRun.model_validate({
        "id": "run_789",
        "workflow_id": "wf_1",
        "workflow_name": "Test",
        "organization_id": "org_1",
        "status": "completed",
        "started_at": "2026-03-13T10:00:00Z",
        "created_at": "2026-03-13T10:00:00Z",
        "updated_at": "2026-03-13T10:00:00Z",
        "trigger_type": "api",
        "config_snapshot_id": "snap_abc",
        "cost_summary": {"total": 0.05, "input_cost": 0.03, "output_cost": 0.02},
        "input_json_data": {"json-1": {"key": "value"}},
        "human_waiting_duration_ms": 5000,
    })
    assert run.trigger_type == "api"
    assert run.config_snapshot_id == "snap_abc"
    assert run.cost_summary == {"total": 0.05, "input_cost": 0.03, "output_cost": 0.02}
    assert run.input_json_data == {"json-1": {"key": "value"}}
    assert run.human_waiting_duration_ms == 5000

    # Without the new fields, defaults apply
    run2 = WorkflowRun.model_validate({
        "id": "run_000", "workflow_id": "wf_1", "workflow_name": "T",
        "organization_id": "org_1", "status": "pending",
        "started_at": "2026-01-01T00:00:00Z",
        "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
    })
    assert run2.trigger_type is None
    assert run2.cost_summary is None
    assert run2.human_waiting_duration_ms == 0


def test_workflow_with_entities_parsing() -> None:
    """WorkflowWithEntities parses blocks, edges, subflows and exposes start_nodes."""
    wfe = WorkflowWithEntities.model_validate({
        "workflow": {
            "id": "wf_1", "name": "Test Workflow",
            "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
        },
        "blocks": [
            {"id": "start-1", "workflow_id": "wf_1", "type": "start", "label": "Document Input"},
            {"id": "extract-1", "workflow_id": "wf_1", "type": "extract", "label": "Extract"},
            {"id": "json-1", "workflow_id": "wf_1", "type": "start_json", "label": "JSON Input"},
            {"id": "end-1", "workflow_id": "wf_1", "type": "end", "label": "Output"},
        ],
        "edges": [
            {"id": "edge-1", "workflow_id": "wf_1", "source_block": "start-1", "target_block": "extract-1",
             "source_handle": "output-file-0", "target_handle": "input-file-0"},
        ],
        "subflows": [],
    })

    assert wfe.workflow.id == "wf_1"
    assert len(wfe.blocks) == 4
    assert len(wfe.edges) == 1

    start_nodes = wfe.start_nodes
    assert len(start_nodes) == 1
    assert start_nodes[0].id == "start-1"

    json_nodes = wfe.start_json_nodes
    assert len(json_nodes) == 1
    assert json_nodes[0].id == "json-1"

    edge = wfe.edges[0]
    assert edge.source_block == "start-1"
    assert edge.target_block == "extract-1"


def test_workflows_create_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "id": "wf_new",
        "name": "My Workflow",
        "description": "A test",
        "is_published": False,
        "email_senders_whitelist": [],
        "email_domains_whitelist": [],
        "created_at": "2026-03-12T10:00:00Z",
        "updated_at": "2026-03-12T10:00:00Z",
    }

    wf = Workflows(client=client).create(name="My Workflow", description="A test")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows"
    assert request.data == {"name": "My Workflow", "description": "A test"}
    assert wf.id == "wf_new"


def test_workflows_publish_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "id": "wf_1",
        "name": "Test",
        "is_published": True,
        "published_snapshot_id": "snap_1",
        "email_senders_whitelist": [],
        "email_domains_whitelist": [],
        "created_at": "2026-03-12T10:00:00Z",
        "updated_at": "2026-03-12T10:00:00Z",
    }

    wf = Workflows(client=client).publish("wf_1", description="v1")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows/wf_1/publish"
    assert wf.is_published is True


def test_workflows_get_entities_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "workflow": {
            "id": "wf_1", "name": "Test",
            "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
        },
        "blocks": [{"id": "start-1", "workflow_id": "wf_1", "type": "start"}],
        "edges": [],
        "subflows": [],
    }

    wfe = Workflows(client=client).get_entities("wf_1")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/wf_1/entities"
    assert isinstance(wfe, WorkflowWithEntities)
    assert len(wfe.start_nodes) == 1


def test_workflows_list_returns_typed_items() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "data": [
            {
                "id": "wf_1", "name": "Workflow A",
                "is_published": True,
                "email_senders_whitelist": [],
                "email_domains_whitelist": [],
                "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
            },
        ],
        "list_metadata": {"before": None, "after": None},
    }

    result = Workflows(client=client).list()
    assert len(result.data) == 1
    assert isinstance(result.data[0], Workflow)
    assert result.data[0].id == "wf_1"


def test_workflow_run_output_single_end_node() -> None:
    """output property returns Dict[str, FinalNodeOutput] for single end node."""
    run = WorkflowRun.model_validate({
        "id": "run_1", "workflow_id": "wf_1", "workflow_name": "Test",
        "organization_id": "org_1", "status": "completed",
        "started_at": "2026-01-01T00:00:00Z",
        "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
        "final_outputs": {
            "end-1": {
                "document": None,
                "data": {"invoice_number": "INV-001", "total": 1234.56},
            },
        },
    })
    output = run.output
    assert output is not None
    assert "end-1" in output
    assert isinstance(output["end-1"], FinalNodeOutput)
    assert output["end-1"].data == {"invoice_number": "INV-001", "total": 1234.56}
    assert output["end-1"].document is None


def test_workflow_run_output_multiple_end_nodes() -> None:
    """output property returns Dict[str, FinalNodeOutput] for multiple end nodes."""
    run = WorkflowRun.model_validate({
        "id": "run_2", "workflow_id": "wf_1", "workflow_name": "Test",
        "organization_id": "org_1", "status": "completed",
        "started_at": "2026-01-01T00:00:00Z",
        "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
        "final_outputs": {
            "end-1": {"document": None, "data": {"count": 3}},
            "end-2": {"document": {"id": "f1", "filename": "out.pdf", "mime_type": "application/pdf"}, "data": {"count": 5}},
        },
    })
    output = run.output
    assert output is not None
    assert len(output) == 2
    assert output["end-1"].data == {"count": 3}
    assert output["end-2"].data == {"count": 5}
    assert output["end-2"].document == {"id": "f1", "filename": "out.pdf", "mime_type": "application/pdf"}


def test_workflow_run_output_none() -> None:
    """output property returns None when no final_outputs."""
    run = WorkflowRun.model_validate({
        "id": "run_3", "workflow_id": "wf_1", "workflow_name": "Test",
        "organization_id": "org_1", "status": "running",
        "started_at": "2026-01-01T00:00:00Z",
        "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
    })
    assert run.output is None


def test_workflow_run_raise_for_status_error() -> None:
    """raise_for_status raises WorkflowRunError on error status."""
    run = WorkflowRun.model_validate({
        "id": "run_err", "workflow_id": "wf_1", "workflow_name": "Test",
        "organization_id": "org_1", "status": "error",
        "error": "Node extract-1 failed: invalid schema",
        "started_at": "2026-01-01T00:00:00Z",
        "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
    })
    with pytest.raises(WorkflowRunError) as exc_info:
        run.raise_for_status()
    assert "run_err" in str(exc_info.value)
    assert "invalid schema" in str(exc_info.value)
    assert exc_info.value.run is run


def test_workflow_run_raise_for_status_ok() -> None:
    """raise_for_status is silent on completed status."""
    run = WorkflowRun.model_validate({
        "id": "run_ok", "workflow_id": "wf_1", "workflow_name": "Test",
        "organization_id": "org_1", "status": "completed",
        "started_at": "2026-01-01T00:00:00Z",
        "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
    })
    run.raise_for_status()  # Should not raise




def test_step_status_extracted_data() -> None:
    """StepStatus.extracted_data works on inline steps in WorkflowRun."""
    run = WorkflowRun.model_validate({
        "id": "run_1", "workflow_id": "wf_1", "workflow_name": "T",
        "organization_id": "org_1", "status": "completed",
        "started_at": "2026-01-01T00:00:00Z",
        "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
        "steps": [
            {
                "node_id": "extract-1", "node_type": "extract", "node_label": "Extract",
                "status": "completed",
                "handle_outputs": {
                    "output-json-0": {"type": "json", "data": {"invoice": "INV-001"}},
                },
            },
            {
                "node_id": "parse-1", "node_type": "parse", "node_label": "Parse",
                "status": "completed",
                "handle_outputs": {
                    "output-file-0": {"type": "file"},
                },
            },
        ],
    })
    assert run.steps[0].extracted_data == {"invoice": "INV-001"}
    assert run.steps[1].extracted_data is None


def test_workflow_run_step_extracted_data() -> None:
    """WorkflowRunStep.extracted_data works with typed handle_outputs."""
    from retab.types.workflows.model import WorkflowRunStep

    step = WorkflowRunStep.model_validate({
        "run_id": "run_1", "organization_id": "org_1",
        "node_id": "extract-1", "step_id": "extract-1",
        "node_type": "extract", "node_label": "Extract",
        "status": "completed",
        "handle_outputs": {
            "output-json-0": {"type": "json", "data": {"total": 1234}},
        },
    })
    assert step.extracted_data == {"total": 1234}
    # handle_outputs should be typed as HandlePayload
    from retab.types.workflows.model import HandlePayload
    payload = step.handle_outputs["output-json-0"]
    assert isinstance(payload, HandlePayload)
    assert payload.type == "json"
