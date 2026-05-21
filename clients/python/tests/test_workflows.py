from datetime import date
from unittest.mock import AsyncMock, MagicMock

import pytest

from retab.resources.workflows.blocks.client import WorkflowBlocks
from retab.resources.workflows.client import AsyncWorkflows, Workflows
from retab.resources.workflows.edges.client import WorkflowEdges
from retab.resources.workflows.runs.client import AsyncWorkflowRuns, WorkflowRuns
from retab.resources.workflows.specs.client import AsyncWorkflowSpecs, WorkflowSpecs
from retab.types.mime import FileRef, MIMEData
from retab.types.workflows.model import (
    DeclarativeApplyResponse,
    DeclarativeExportResponse,
    DeclarativePlanResponse,
    DeclarativeValidationResponse,
    WorkflowRun,
    WorkflowBlock,
    WorkflowWithEntities,
    Workflow,
    WorkflowBlockCreateRequest,
    WorkflowBlockUpdateRequest,
    WorkflowEdgeCreateRequest,
    StepExecutionResponse,
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
        "list_metadata": {"before": None, "after": "workflow_after"},
    }

    result = Workflows(client=client).list(
        limit=5,
        order="asc",
        sort_by="updated_at",
        fields="id,name",
        after="workflow_before",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows"
    assert request.params == {
        "limit": 5,
        "order": "asc",
        "sort_by": "updated_at",
        "fields": "id,name",
        "after": "workflow_before",
    }
    assert result.list_metadata.after == "workflow_after"


@pytest.mark.asyncio
async def test_async_workflows_get_uses_detail_route() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(
        return_value={
            "id": "workflow_123",
            "name": "Test Workflow",
            "description": "",
            "published": None,
            "email_trigger": {"allowed_senders": [], "allowed_domains": []},
            "created_at": "2026-03-12T10:00:00Z",
            "updated_at": "2026-03-12T10:00:00Z",
        }
    )

    workflow = await AsyncWorkflows(client=client).get("workflow_123")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/workflow_123"
    assert workflow.id == "workflow_123"


@pytest.mark.asyncio
async def test_async_workflows_list_uses_paginated_route() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(
        return_value={
            "data": [],
            "list_metadata": {"before": None, "after": "workflow_after"},
        }
    )

    result = await AsyncWorkflows(client=client).list(
        limit=5,
        order="asc",
        sort_by="updated_at",
        fields="id,name",
        after="workflow_before",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows"
    assert request.params == {
        "limit": 5,
        "order": "asc",
        "sort_by": "updated_at",
        "fields": "id,name",
        "after": "workflow_before",
    }
    assert result.list_metadata.after == "workflow_after"


def test_workflows_exposes_specs_subresource() -> None:
    workflows = Workflows(client=MagicMock())

    assert isinstance(workflows.specs, WorkflowSpecs)


def test_async_workflows_exposes_specs_subresource() -> None:
    workflows = AsyncWorkflows(client=MagicMock())

    assert isinstance(workflows.specs, AsyncWorkflowSpecs)


def test_removed_workflow_methods_are_not_exposed() -> None:
    workflows = Workflows(client=MagicMock())

    assert not hasattr(workflows, "duplicate")
    assert not hasattr(workflows, "get_entities")
    assert not hasattr(workflows, "get_resolved_schemas")
    assert not hasattr(workflows, "list_snapshots")
    assert not hasattr(workflows.blocks, "config_history")
    assert not hasattr(workflows.blocks, "get_resolved_schemas")
    assert not hasattr(workflows.blocks, "create_batch")
    assert not hasattr(workflows.blocks, "list_simulations")
    assert not hasattr(workflows.blocks, "simulate")
    assert not hasattr(workflows.edges, "create_batch")
    assert not hasattr(workflows.edges, "delete_all")
    assert not hasattr(workflows.experiments, "duplicate")
    assert not hasattr(workflows.experiments, "list_eligible_blocks")
    assert not hasattr(workflows.reviews, "wait_for")


def test_workflow_specs_validate_uses_spec_validate_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "workflow_id": "wf_1",
        "block_count": 2,
        "edge_count": 1,
        "is_valid": True,
        "diagnostics": {"issues": []},
    }

    response = WorkflowSpecs(client=client).validate("apiVersion: workflows.retab.com/v1alpha2\n")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows/spec/validate"
    assert request.data == {"yaml_definition": "apiVersion: workflows.retab.com/v1alpha2\n"}
    assert isinstance(response, DeclarativeValidationResponse)
    assert response.is_valid is True


def test_workflow_specs_plan_uses_spec_plan_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "workflow_id": "wf_1",
        "action": "noop",
        "block_count": 2,
        "edge_count": 1,
        "diagnostics": {"issues": []},
        "format_version": "workflows-plan/v1",
        "summary": {"add": 0, "change": 1, "destroy": 0, "replace": 0, "noop": 0, "total": 1, "has_changes": True},
        "resource_changes": [
            {
                "address": "workflow.wf_1.block.block_extract",
                "target": "block",
                "target_id": "block_extract",
                "name": "Extract",
                "type": "extract",
                "actions": ["update"],
                "summary": "Update block 'Extract'",
                "change": {
                    "before": {"config": {"model": "old"}},
                    "after": {"config": {"model": "new"}},
                    "before_sensitive": {},
                    "after_sensitive": {},
                    "field_changes": [
                        {
                            "path": ["config", "model"],
                            "path_display": "config.model",
                            "action": "update",
                            "before": "old",
                            "after": "new",
                        }
                    ],
                },
            }
        ],
        "rendered_plan": "Plan: 0 to add, 1 to change, 0 to destroy.",
    }

    response = WorkflowSpecs(client=client).plan("spec: {}\n")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows/spec/plan"
    assert request.data == {"yaml_definition": "spec: {}\n"}
    assert isinstance(response, DeclarativePlanResponse)
    assert response.summary.change == 1
    assert response.resource_changes[0].change.field_changes[0].path_display == "config.model"
    assert "1 to change" in response.rendered_plan


def test_workflow_specs_apply_uses_spec_apply_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "workflow_id": "wf_1",
        "created": False,
        "block_count": 2,
        "edge_count": 1,
        "diagnostics": {"issues": []},
        "format_version": "workflows-plan/v1",
        "summary": {"add": 0, "change": 0, "destroy": 0, "replace": 0, "noop": 1, "total": 0, "has_changes": False},
        "resource_changes": [],
        "rendered_plan": "No changes. Infrastructure is up-to-date.",
    }

    response = WorkflowSpecs(client=client).apply("spec: {}\n")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows/spec/apply"
    assert request.data == {"yaml_definition": "spec: {}\n"}
    assert isinstance(response, DeclarativeApplyResponse)
    assert response.summary.noop == 1
    assert response.resource_changes == []


def test_workflow_specs_export_uses_spec_export_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "workflow_id": "wf_1",
        "yaml_definition": "apiVersion: workflows.retab.com/v1alpha2\n",
    }

    response = WorkflowSpecs(client=client).export("wf_1")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/wf_1/spec"
    assert isinstance(response, DeclarativeExportResponse)
    assert response.yaml_definition.startswith("apiVersion:")


@pytest.mark.asyncio
async def test_async_workflow_specs_validate_uses_spec_validate_route() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(
        return_value={
            "workflow_id": "wf_1",
            "block_count": 2,
            "edge_count": 1,
            "is_valid": True,
            "diagnostics": {"issues": []},
        }
    )

    response = await AsyncWorkflowSpecs(client=client).validate("spec: {}\n")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows/spec/validate"
    assert response.workflow_id == "wf_1"


def test_workflow_run_ignores_legacy_steps_payload() -> None:
    """The run object no longer embeds a step roster.

    Older servers may still return a ``steps`` field in the run payload; the
    SDK must ignore it (extra="ignore" on the model). Steps are fetched
    separately via ``client.workflows.steps.list(run_id)``.
    """
    run = WorkflowRun.model_validate(
        {
            "id": "run_123",
            "organization_id": "org_123",
            "workflow": {
                "workflow_id": "workflow_123",
                "version_id": "ver_0123456789abcdef0123456789abcdef",
                "name_at_run_time": "Classifier Workflow",
            },
            "trigger": {"type": "manual"},
            "lifecycle": {"status": "running"},
            "timing": {
                "created_at": "2026-03-13T10:00:00Z",
                "started_at": "2026-03-13T10:00:00Z",
            },
            "steps": [
                {
                    "block_id": "classifier-1",
                    "step_id": "classifier-1",
                    "block_type": "classifier",
                    "block_label": "Classifier",
                    "status": "completed",
                }
            ],
        }
    )

    # The legacy ``steps`` payload is silently dropped — it is no longer a
    # field on the model.
    assert "steps" not in run.model_dump()


def test_workflow_run_v2_typed_fields() -> None:
    """v2 nested fields parse and round-trip with the typed sub-models."""
    run = WorkflowRun.model_validate(
        {
            "id": "run_789",
            "organization_id": "org_1",
            "workflow": {
                "workflow_id": "wf_1",
                "version_id": "ver_abcdef0123456789abcdef0123456789",
                "name_at_run_time": "Test",
                "requested_version": "ver_abcdef0123456789abcdef0123456789",
            },
            "trigger": {"type": "api", "api_key_id": "ak_1"},
            "lifecycle": {"status": "completed"},
            "timing": {
                "created_at": "2026-03-13T10:00:00Z",
                "started_at": "2026-03-13T10:00:00Z",
                "completed_at": "2026-03-13T10:00:05Z",
                "accumulated_review_waiting_ms": 5000,
            },
            "inputs": {"documents": {}, "json_data": {"json-1": {"key": "value"}}},
        }
    )
    assert run.workflow.workflow_id == "wf_1"
    assert run.workflow.version_id == "ver_abcdef0123456789abcdef0123456789"
    assert run.workflow.name_at_run_time == "Test"
    assert run.trigger.type == "api"
    assert run.lifecycle.status == "completed"
    assert run.inputs.json_data == {"json-1": {"key": "value"}}
    assert run.timing.accumulated_review_waiting_ms == 5000
    assert not hasattr(run.timing, "duration_ms")
    assert not hasattr(run.timing, "active_duration_ms")
    assert "duration_ms" not in run.timing.model_dump()
    assert "active_duration_ms" not in run.timing.model_dump()

    # Defaults: inputs default to empty
    run2 = WorkflowRun.model_validate(
        {
            "id": "run_000",
            "organization_id": "org_1",
            "workflow": {
                "workflow_id": "wf_1",
                "version_id": "ver_0123456789abcdef0123456789abcdef",
                "name_at_run_time": "T",
            },
            "trigger": {"type": "manual"},
            "lifecycle": {"status": "pending"},
            "timing": {"created_at": "2026-01-01T00:00:00Z"},
        }
    )
    assert run2.inputs.documents == {}
    assert run2.inputs.json_data == {}
    assert run2.timing.accumulated_review_waiting_ms == 0


def test_workflow_with_entities_parsing() -> None:
    """WorkflowWithEntities parses blocks and edges and exposes start_document_blocks."""
    wfe = WorkflowWithEntities.model_validate(
        {
            "workflow": {
                "id": "wf_1",
                "name": "Test Workflow",
                "created_at": "2026-01-01T00:00:00Z",
                "updated_at": "2026-01-01T00:00:00Z",
            },
            "blocks": [
                {"id": "start-1", "workflow_id": "wf_1", "organization_id": "org_1", "draft_version": "draft_1", "type": "start-document", "label": "Document Input"},
                {"id": "extract-1", "workflow_id": "wf_1", "organization_id": "org_1", "draft_version": "draft_1", "type": "extract", "label": "Extract"},
                {"id": "json-1", "workflow_id": "wf_1", "organization_id": "org_1", "draft_version": "draft_1", "type": "start_json", "label": "JSON Input"},
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
        }
    )

    assert wfe.workflow.id == "wf_1"
    assert len(wfe.blocks) == 3
    assert len(wfe.edges) == 1

    start_document_blocks = wfe.start_document_blocks
    assert len(start_document_blocks) == 1
    assert start_document_blocks[0].id == "start-1"

    json_blocks = wfe.start_json_blocks
    assert len(json_blocks) == 1
    assert json_blocks[0].id == "json-1"

    edge = wfe.edges[0]
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


def test_workflows_update_accepts_email_trigger_policy() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "id": "workflow_123",
        "name": "Test Workflow",
        "description": "",
        "published": None,
        "email_trigger": {
            "allowed_senders": ["ops@example.com"],
            "allowed_domains": ["example.com"],
        },
        "created_at": "2026-03-12T10:00:00Z",
        "updated_at": "2026-03-12T10:00:00Z",
    }

    workflow = Workflows(client=client).update(
        "workflow_123",
        email_trigger={
            "allowed_senders": ["ops@example.com"],
            "allowed_domains": ["example.com"],
        },
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "PATCH"
    assert request.url == "/workflows/workflow_123"
    assert request.data == {
        "email_trigger": {
            "allowed_senders": ["ops@example.com"],
            "allowed_domains": ["example.com"],
        }
    }
    assert workflow.email_trigger.allowed_senders == ["ops@example.com"]


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
    assert request.url == "/workflows/blocks?workflow_id=wf_1"
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
    assert request.url == "/workflows/blocks/extract-1?workflow_id=wf_1"
    assert request.data == {"label": "Renamed", "position_x": 200.0}
    assert block.label == "Renamed"


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

    assert block.draft_version == "draft_1"


def test_workflow_block_exposes_resolved_schema_sidecar() -> None:
    block = WorkflowBlock.model_validate(
        {
            "id": "extract-1",
            "workflow_id": "wf_1",
            "organization_id": "org_1",
            "draft_version": "draft_1",
            "type": "extract",
            "label": "Extract",
            "resolved_schemas": {
                "input_schemas": {},
                "output_schemas": {
                    "output-json-0": {
                        "type": "object",
                        "properties": {"invoice_number": {"type": "string"}},
                    }
                },
            },
        }
    )

    assert block.resolved_schemas is not None
    assert block.resolved_schemas.output_schemas["output-json-0"]["properties"]["invoice_number"]["type"] == "string"


def test_step_execution_response_ignores_removed_payload_schema_fields() -> None:
    removed_payload_key = "raw" + "_" + "output"
    step_execution = StepExecutionResponse.model_validate(
        {
            "block_id": "extract-1",
            "block_type": "extract",
            "block_label": "Extract",
            "lifecycle": {"status": "completed"},
            "output": {
                "data": {"invoice_number": "INV-001"},
                "json_schema": {"type": "object"},
            },
            removed_payload_key: {
                "json_schema": {"type": "object"},
            },
            "json_schema": {"type": "object"},
            "handle_outputs": {
                "output-json-0": {
                    "type": "json",
                    "data": {"invoice_number": "INV-001"},
                }
            },
        }
    )

    dumped = step_execution.model_dump()
    assert "output" not in dumped
    assert removed_payload_key not in dumped
    assert "json_schema" not in dumped
    assert step_execution.get_json_output() == {"invoice_number": "INV-001"}


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
    assert request.url == "/workflows/edges?workflow_id=wf_1"
    assert request.data == {
        "id": "edge-1",
        "source_block": "start-1",
        "target_block": "extract-1",
        "source_handle": "output-file-0",
        "target_handle": "input-file-0",
    }
    assert edge.id == "edge-1"


def test_workflows_publish_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "id": "wf_1",
        "name": "Test",
        "published": {
            "version_id": "ver_0123456789abcdef0123456789abcdef",
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
    assert wf.published_version_id == "ver_0123456789abcdef0123456789abcdef"


def test_workflows_list_returns_typed_items() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "data": [
            {
                "id": "wf_1",
                "name": "Workflow A",
                "published": {
                    "version_id": "ver_0123456789abcdef0123456789abcdef",
                    "published_at": "2026-01-01T00:00:00Z",
                },
                "email_trigger": {"allowed_senders": [], "allowed_domains": []},
                "created_at": "2026-01-01T00:00:00Z",
                "updated_at": "2026-01-01T00:00:00Z",
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
        after="run_after",
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
    assert request.params["after"] == "run_after"


def test_workflow_runs_create_without_inputs_sends_json_body() -> None:
    request = WorkflowRuns(client=MagicMock()).prepare_create(workflow_id="wf_1")

    assert request.method == "POST"
    assert request.url == "/workflows/runs"
    assert request.data == {
        "workflow_id": "wf_1",
        "documents": {},
        "json_inputs": {},
        "version": "production",
    }


def test_workflow_runs_create_passes_file_refs_without_content() -> None:
    request = WorkflowRuns(client=MagicMock()).prepare_create(
        workflow_id="wf_1",
        documents={
            "start_1": FileRef(
                id="file_existing",
                filename="invoice.pdf",
                mime_type="application/pdf",
            )
        },
    )

    assert request.data["documents"] == {
        "start_1": {
            "id": "file_existing",
            "filename": "invoice.pdf",
            "mime_type": "application/pdf",
        }
    }
    assert "content" not in request.data["documents"]["start_1"]


def test_workflow_runs_create_keeps_mime_data_content_for_new_documents() -> None:
    request = WorkflowRuns(client=MagicMock()).prepare_create(
        workflow_id="wf_1",
        documents={
            "start_1": MIMEData(
                filename="note.txt",
                url="data:text/plain;base64,aGVsbG8=",
            )
        },
    )

    assert request.data["documents"] == {
        "start_1": {
            "filename": "note.txt",
            "content": "aGVsbG8=",
            "mime_type": "text/plain",
        }
    }


def _v2_run_payload(**overrides) -> dict:
    """Helper: build a minimal v2 WorkflowRun JSON payload for fixtures."""
    payload: dict = {
        "id": overrides.pop("id", "run_1"),
        "organization_id": "org_1",
        "workflow": {
            "workflow_id": "wf_1",
            "version_id": "ver_0123456789abcdef0123456789abcdef",
            "name_at_run_time": "Test",
        },
        "trigger": overrides.pop("trigger", {"type": "manual"}),
        "lifecycle": overrides.pop("lifecycle", {"status": "running"}),
        "timing": overrides.pop(
            "timing",
            {
                "created_at": "2026-01-01T00:00:00Z",
                "started_at": "2026-01-01T00:00:00Z",
            },
        ),
    }
    payload.update(overrides)
    return payload


def test_workflow_runs_cancel_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "run": _v2_run_payload(lifecycle={"status": "cancelled"}),
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
    client._prepared_request.return_value = _v2_run_payload(id="run_2", lifecycle={"status": "running"})

    run = WorkflowRuns(client=client).restart("run_1", command_id="cmd_2")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows/runs"
    assert request.data == {"restart_of": "run_1", "command_id": "cmd_2"}
    assert run.id == "run_2"


def test_workflow_runs_export_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "csv_data": "a,b\n1,2\n",
        "rows": 1,
        "columns": 2,
    }

    result = WorkflowRuns(client=client).export(
        workflow_id="wf_1",
        block_id="extract-1",
        export_source="outputs",
        selected_run_ids=["run_1", "run_2"],
        trigger_types=["api"],
        preferred_columns=["invoice_number", "total_amount"],
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/workflows/runs/export-payload"
    assert request.data == {
        "workflow_id": "wf_1",
        "block_id": "extract-1",
        "export_source": "outputs",
        "preferred_columns": ["invoice_number", "total_amount"],
        "selected_run_ids": ["run_1", "run_2"],
        "trigger_types": ["api"],
    }
    assert result.rows == 1
    assert result.columns == 2


def test_workflow_runs_do_not_expose_wait_for_completion() -> None:
    assert not hasattr(WorkflowRuns(client=MagicMock()), "wait_for_completion")
    assert not hasattr(AsyncWorkflowRuns(client=MagicMock()), "wait_for_completion")


def test_workflow_run_step_extracted_data() -> None:
    """WorkflowRunStep.extracted_data works with typed handle_outputs."""
    from retab.types.workflows.model import WorkflowRunStep

    step = WorkflowRunStep.model_validate(
        {
            "run_id": "run_1",
            "organization_id": "org_1",
            "block_id": "extract-1",
            "step_id": "extract-1",
            "block_type": "extract",
            "block_label": "Extract",
            "lifecycle": {"status": "completed"},
            "handle_outputs": {
                "output-json-0": {"type": "json", "data": {"total": 1234}},
            },
        }
    )
    assert step.extracted_data == {"total": 1234}
    dumped = step.model_dump()
    assert "status" not in dumped
    assert "terminal" not in dumped
    # handle_outputs should be typed as HandlePayload
    from retab.types.workflows.model import HandlePayload

    payload = step.handle_outputs["output-json-0"]
    assert isinstance(payload, HandlePayload)
    assert payload.type == "json"
    assert "input_document" not in WorkflowRunStep.model_fields
    assert "output_document" not in WorkflowRunStep.model_fields
    assert "split_documents" not in WorkflowRunStep.model_fields
    assert "status" not in WorkflowRunStep.model_fields
    assert "terminal" not in WorkflowRunStep.model_fields


def test_workflow_run_step_accepts_json_ref_handle_payload() -> None:
    from retab.types.workflows.model import WorkflowRunStep

    step = WorkflowRunStep.model_validate(
        {
            "run_id": "run_1",
            "organization_id": "org_1",
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
        }
    )

    payload = step.handle_outputs["output-json-0"]
    assert payload.type == "json_ref"
    assert payload.artifact_ref == {
        "operation": "workflow_step_json",
        "id": "artifact_123",
        "key": "output-json-0",
    }
    assert payload.preview == {"truncated": True}
