from datetime import date
from unittest.mock import AsyncMock, MagicMock

import pytest

from retab.resources.workflows.blocks.client import WorkflowBlocks
from retab.resources.workflows.client import AsyncWorkflows, Workflows
from retab.resources.workflows.edges.client import WorkflowEdges
from retab.resources.workflows.runs.client import AsyncWorkflowRuns, WorkflowRuns
from retab.types.workflows.model import (
    WorkflowRun,
    WorkflowRunError,
    WorkflowBlock,
    WorkflowEdgeDoc,
    WorkflowSubflow,
    WorkflowWithEntities,
    Workflow,
    WorkflowBlockCreateRequest,
    WorkflowBlockUpdateRequest,
    WorkflowEdgeCreateRequest,
)


def test_workflows_get_uses_detail_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "id": "workflow_123",
        "name": "Test Workflow",
        "description": "",
        "published": None,
        "email_trigger": {"allowed_senders": [], "allowed_domains": []},
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
        "published": None,
        "email_trigger": {"allowed_senders": [], "allowed_domains": []},
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
            {"id": "start-1", "workflow_id": "wf_1", "organization_id": "org_1", "draft_version": "draft_1", "type": "start", "label": "Document Input"},
            {"id": "extract-1", "workflow_id": "wf_1", "organization_id": "org_1", "draft_version": "draft_1", "type": "extract", "label": "Extract"},
            {"id": "json-1", "workflow_id": "wf_1", "organization_id": "org_1", "draft_version": "draft_1", "type": "start_json", "label": "JSON Input"},
            {"id": "end-1", "workflow_id": "wf_1", "organization_id": "org_1", "draft_version": "draft_1", "type": "end", "label": "Output"},
        ],
        "edges": [
            {
                "id": "edge-1",
                "workflow_id": "wf_1",
                "organization_id": "org_1",
                "draft_version": "draft_1",
                "source_block": "start-1",
                "target_block": "extract-1",
                "source_handle": "output-file-0",
                "target_handle": "input-file-0",
            },
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
    assert edge.organization_id == "org_1"
    assert edge.draft_version == "draft_1"
    assert edge.source_block == "start-1"
    assert edge.target_block == "extract-1"


def test_workflows_create_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "id": "wf_new",
        "name": "My Workflow",
        "description": "A test",
        "published": None,
        "email_trigger": {"allowed_senders": [], "allowed_domains": []},
        "created_at": "2026-03-12T10:00:00Z",
        "updated_at": "2026-03-12T10:00:00Z",
    }

    wf = Workflows(client=client).create(name="My Workflow", description="A test")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows"
    assert request.data == {"name": "My Workflow", "description": "A test"}
    assert wf.id == "wf_new"


def test_workflow_blocks_create_accepts_typed_request() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "id": "extract-1",
        "workflow_id": "wf_1",
        "organization_id": "org_1",
        "draft_version": "draft_1",
        "type": "extract",
        "label": "Extract",
    }

    block = WorkflowBlocks(client=client).create(
        "wf_1",
        request=WorkflowBlockCreateRequest(
            id="extract-1",
            type="extract",
            label="Extract",
            position_x=120,
            position_y=80,
            config={"json_schema": {"type": "object"}},
        ),
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows/wf_1/blocks"
    assert request.data == {
        "id": "extract-1",
        "type": "extract",
        "label": "Extract",
        "position_x": 120.0,
        "position_y": 80.0,
        "config": {"json_schema": {"type": "object"}},
    }
    assert block.id == "extract-1"


def test_workflow_blocks_update_accepts_typed_request() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "id": "extract-1",
        "workflow_id": "wf_1",
        "organization_id": "org_1",
        "draft_version": "draft_1",
        "type": "extract",
        "label": "Renamed",
    }

    block = WorkflowBlocks(client=client).update(
        "wf_1",
        request=WorkflowBlockUpdateRequest(
            block_id="extract-1",
            label="Renamed",
            position_x=200,
        ),
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "PATCH"
    assert request.url == "/workflows/wf_1/blocks/extract-1"
    assert request.data == {"label": "Renamed", "position_x": 200.0}
    assert block.label == "Renamed"


def test_workflow_blocks_create_batch_accepts_typed_requests() -> None:
    client = MagicMock()
    client._prepared_request.return_value = [
        {"id": "start-1", "workflow_id": "wf_1", "organization_id": "org_1", "draft_version": "draft_1", "type": "start"},
        {"id": "extract-1", "workflow_id": "wf_1", "organization_id": "org_1", "draft_version": "draft_1", "type": "extract"},
    ]

    blocks = WorkflowBlocks(client=client).create_batch(
        "wf_1",
        [
            WorkflowBlockCreateRequest(id="start-1", type="start"),
            WorkflowBlockCreateRequest(id="extract-1", type="extract"),
        ],
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows/wf_1/blocks/batch"
    assert request.data == [
        {"id": "start-1", "type": "start", "label": "", "position_x": 0.0, "position_y": 0.0},
        {"id": "extract-1", "type": "extract", "label": "", "position_x": 0.0, "position_y": 0.0},
    ]
    assert [block.id for block in blocks] == ["start-1", "extract-1"]


def test_workflow_block_parses_live_editing_metadata() -> None:
    block = WorkflowBlock.model_validate(
        {
            "id": "extract-1",
            "workflow_id": "wf_1",
            "organization_id": "org_1",
            "draft_version": "draft_1",
            "type": "extract",
            "label": "Extract",
        }
    )

    assert block.organization_id == "org_1"
    assert block.draft_version == "draft_1"


def test_workflow_edges_create_accepts_typed_request() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "id": "edge-1",
        "workflow_id": "wf_1",
        "organization_id": "org_1",
        "draft_version": "draft_1",
        "source_block": "start-1",
        "target_block": "extract-1",
        "source_handle": "output-file-0",
        "target_handle": "input-file-0",
    }

    edge = WorkflowEdges(client=client).create(
        "wf_1",
        request=WorkflowEdgeCreateRequest(
            id="edge-1",
            source_block="start-1",
            target_block="extract-1",
            source_handle="output-file-0",
            target_handle="input-file-0",
        ),
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows/wf_1/edges"
    assert request.data == {
        "id": "edge-1",
        "source_block": "start-1",
        "target_block": "extract-1",
        "source_handle": "output-file-0",
        "target_handle": "input-file-0",
    }
    assert edge.id == "edge-1"


def test_workflow_edges_create_batch_accepts_typed_requests() -> None:
    client = MagicMock()
    client._prepared_request.return_value = [
        {
            "id": "edge-1",
            "workflow_id": "wf_1",
            "organization_id": "org_1",
            "draft_version": "draft_1",
            "source_block": "start-1",
            "target_block": "extract-1",
        },
    ]

    edges = WorkflowEdges(client=client).create_batch(
        "wf_1",
        [WorkflowEdgeCreateRequest(id="edge-1", source_block="start-1", target_block="extract-1")],
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows/wf_1/edges/batch"
    assert request.data == [
        {"id": "edge-1", "source_block": "start-1", "target_block": "extract-1"},
    ]
    assert [edge.id for edge in edges] == ["edge-1"]


def test_workflows_publish_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "id": "wf_1",
        "name": "Test",
        "published": {
            "snapshot_id": "snap_1",
            "published_at": "2026-03-12T10:00:00Z",
        },
        "email_trigger": {"allowed_senders": [], "allowed_domains": []},
        "created_at": "2026-03-12T10:00:00Z",
        "updated_at": "2026-03-12T10:00:00Z",
    }

    wf = Workflows(client=client).publish("wf_1", description="v1")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows/wf_1/publish"
    assert wf.published_snapshot_id == "snap_1"


def test_workflows_get_entities_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "workflow": {
            "id": "wf_1", "name": "Test",
            "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
        },
        "blocks": [{"id": "start-1", "workflow_id": "wf_1", "organization_id": "org_1", "draft_version": "draft_1", "type": "start"}],
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
                "published": {
                    "snapshot_id": "snap_1",
                    "published_at": "2026-01-01T00:00:00Z",
                },
                "email_trigger": {"allowed_senders": [], "allowed_domains": []},
                "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
            },
        ],
        "list_metadata": {"before": None, "after": None},
    }

    result = Workflows(client=client).list()
    assert len(result.data) == 1
    assert isinstance(result.data[0], Workflow)
    assert result.data[0].id == "wf_1"


def test_workflow_runs_list_serializes_pythonic_filters() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "data": [],
        "list_metadata": {"before": None, "after": None},
    }

    WorkflowRuns(client=client).list(
        workflow_id="wf_1",
        statuses=["completed", "error"],
        trigger_types=["api", "email"],
        from_date=date(2026, 1, 1),
        to_date=date(2026, 1, 31),
        fields=["id", "status"],
        after="cursor_1",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/runs"
    assert request.params["workflow_id"] == "wf_1"
    assert request.params["statuses"] == "completed,error"
    assert request.params["trigger_types"] == "api,email"
    assert request.params["from_date"] == "2026-01-01"
    assert request.params["to_date"] == "2026-01-31"
    assert request.params["fields"] == "id,status"
    assert request.params["after"] == "cursor_1"


def test_workflow_runs_cancel_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "run": {
            "id": "run_1",
            "workflow_id": "wf_1",
            "workflow_name": "Test",
            "organization_id": "org_1",
            "status": "cancelled",
            "started_at": "2026-01-01T00:00:00Z",
            "created_at": "2026-01-01T00:00:00Z",
            "updated_at": "2026-01-01T00:00:00Z",
        },
        "cancellation_status": "cancelled",
    }

    result = WorkflowRuns(client=client).cancel("run_1", command_id="cmd_1")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows/runs/run_1/cancel"
    assert request.data == {"command_id": "cmd_1"}
    assert result.cancellation_status == "cancelled"
    assert result.run.id == "run_1"


def test_workflow_runs_restart_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "id": "run_2",
        "workflow_id": "wf_1",
        "workflow_name": "Test",
        "organization_id": "org_1",
        "status": "running",
        "started_at": "2026-01-01T00:00:00Z",
        "created_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-01-01T00:00:00Z",
    }

    run = WorkflowRuns(client=client).restart("run_1", command_id="cmd_2")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows/runs/run_1/restart"
    assert request.data == {"command_id": "cmd_2"}
    assert run.id == "run_2"


def test_workflow_runs_submit_hil_decision_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "submission_status": "accepted",
        "decision": {
            "run_id": "run_1",
            "node_id": "hil-1",
            "node_status": "waiting_for_hil",
            "decision_received": True,
            "decision_applied": False,
            "approved": True,
            "modified_data": {"field": "value"},
            "payload_hash": "hash_1",
            "received_at": "2026-01-01T00:00:01Z",
            "applied_at": None,
        },
    }

    result = WorkflowRuns(client=client).submit_hil_decision(
        "run_1",
        node_id="hil-1",
        approved=True,
        modified_data={"field": "value"},
        command_id="cmd_3",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows/runs/run_1/hil-decisions"
    assert request.data == {
        "node_id": "hil-1",
        "approved": True,
        "modified_data": {"field": "value"},
        "command_id": "cmd_3",
    }
    assert result.submission_status == "accepted"
    assert result.decision.node_id == "hil-1"


def test_workflow_runs_submit_hil_decision_accepts_already_applied_status() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "submission_status": "already_applied",
        "decision": {
            "run_id": "run_1",
            "node_id": "hil-1",
            "node_status": "completed",
            "decision_received": True,
            "decision_applied": True,
            "approved": True,
            "modified_data": {"field": "value"},
            "payload_hash": "hash_1",
            "received_at": "2026-01-01T00:00:01Z",
            "applied_at": "2026-01-01T00:00:05Z",
        },
    }

    result = WorkflowRuns(client=client).submit_hil_decision(
        "run_1",
        node_id="hil-1",
        approved=True,
    )

    assert result.submission_status == "already_applied"
    assert result.decision.decision_applied is True


def test_workflow_runs_get_hil_decision_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "run_id": "run_1",
        "node_id": "hil-1",
        "node_status": "completed",
        "decision_received": True,
        "decision_applied": True,
        "approved": True,
        "modified_data": {"field": "value"},
        "payload_hash": "hash_1",
        "received_at": "2026-01-01T00:00:01Z",
        "applied_at": "2026-01-01T00:00:05Z",
    }

    result = WorkflowRuns(client=client).get_hil_decision("run_1", "hil-1")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/runs/run_1/hil-decisions/hil-1"
    assert result.decision_applied is True


def test_workflow_runs_export_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "csv_data": "a,b\n1,2\n",
        "rows": 1,
        "columns": 2,
    }

    result = WorkflowRuns(client=client).export(
        workflow_id="wf_1",
        node_id="extract-1",
        export_source="outputs",
        selected_run_ids=["run_1", "run_2"],
        trigger_types=["api"],
        preferred_columns=["invoice_number", "total_amount"],
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows/runs/export_payload"
    assert request.data == {
        "workflow_id": "wf_1",
        "node_id": "extract-1",
        "export_source": "outputs",
        "preferred_columns": ["invoice_number", "total_amount"],
        "selected_run_ids": ["run_1", "run_2"],
        "trigger_types": ["api"],
    }
    assert result.rows == 1
    assert result.columns == 2


def test_workflow_run_final_outputs_preserved() -> None:
    """final_outputs stays as the raw backend payload."""
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
    assert run.final_outputs == {
        "end-1": {
            "document": None,
            "data": {"invoice_number": "INV-001", "total": 1234.56},
        },
    }


def test_workflow_run_has_no_output_property() -> None:
    """The deprecated output convenience property is not exposed."""
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
    with pytest.raises(AttributeError):
        _ = run.output


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


def test_workflow_run_raise_for_status_cancelled() -> None:
    """raise_for_status raises WorkflowRunError on cancelled status."""
    run = WorkflowRun.model_validate({
        "id": "run_cancelled", "workflow_id": "wf_1", "workflow_name": "Test",
        "organization_id": "org_1", "status": "cancelled",
        "started_at": "2026-01-01T00:00:00Z",
        "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
    })
    with pytest.raises(WorkflowRunError) as exc_info:
        run.raise_for_status()
    assert "run_cancelled" in str(exc_info.value)
    assert "cancelled" in str(exc_info.value)
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


def test_wait_for_completion_returns_waiting_for_human(monkeypatch: pytest.MonkeyPatch) -> None:
    client = MagicMock()
    client._prepared_request.side_effect = [
        {
            "id": "run_1", "workflow_id": "wf_1", "workflow_name": "Test",
            "organization_id": "org_1", "status": "running",
            "started_at": "2026-01-01T00:00:00Z",
            "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
        },
        {
            "id": "run_1", "workflow_id": "wf_1", "workflow_name": "Test",
            "organization_id": "org_1", "status": "waiting_for_human",
            "waiting_for_node_ids": ["hil-1"],
            "started_at": "2026-01-01T00:00:00Z",
            "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:01Z",
        },
    ]
    monkeypatch.setattr("retab.resources.workflows.runs.client.time.sleep", lambda _: None)

    run = WorkflowRuns(client=client).wait_for_completion("run_1")

    assert run.status == "waiting_for_human"
    assert run.waiting_for_node_ids == ["hil-1"]


@pytest.mark.asyncio
async def test_async_wait_for_completion_returns_waiting_for_human(monkeypatch: pytest.MonkeyPatch) -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(side_effect=[
        {
            "id": "run_1", "workflow_id": "wf_1", "workflow_name": "Test",
            "organization_id": "org_1", "status": "running",
            "started_at": "2026-01-01T00:00:00Z",
            "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
        },
        {
            "id": "run_1", "workflow_id": "wf_1", "workflow_name": "Test",
            "organization_id": "org_1", "status": "waiting_for_human",
            "waiting_for_node_ids": ["hil-1"],
            "started_at": "2026-01-01T00:00:00Z",
            "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:01Z",
        },
    ])

    async def _no_sleep(_: float) -> None:
        return None

    monkeypatch.setattr("retab.resources.workflows.runs.client.asyncio.sleep", _no_sleep)

    run = await AsyncWorkflowRuns(client=client).wait_for_completion("run_1")

    assert run.status == "waiting_for_human"
    assert run.waiting_for_node_ids == ["hil-1"]


@pytest.mark.asyncio
async def test_async_wait_for_completion_awaits_async_callback() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(return_value={
        "id": "run_1", "workflow_id": "wf_1", "workflow_name": "Test",
        "organization_id": "org_1", "status": "completed",
        "started_at": "2026-01-01T00:00:00Z",
        "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
    })
    seen_run_ids: list[str] = []

    async def on_status(run: WorkflowRun) -> None:
        seen_run_ids.append(run.id)

    run = await AsyncWorkflowRuns(client=client).wait_for_completion("run_1", on_status=on_status)

    assert run.status == "completed"
    assert seen_run_ids == ["run_1"]




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
