# pyright: reportAttributeAccessIssue=false, reportArgumentType=false, reportCallIssue=false
from datetime import date
from unittest.mock import AsyncMock, MagicMock

import pytest
from pydantic import ValidationError

from retab.resources.workflows.blocks import WorkflowBlocks
from retab.resources.workflows import AsyncWorkflows, Workflows
from retab.resources.workflows.edges import WorkflowEdges
from retab.resources.workflows.runs import AsyncWorkflowRuns, WorkflowRuns
from retab.resources.workflows.blocks.executions import AsyncWorkflowBlockExecutions, WorkflowBlockExecutions
from retab.types.workflows.blocks import WorkflowBlockVersion
from retab.types.workflows.edges import WorkflowEdgeVersion
from retab.resources.workflows.spec import AsyncWorkflowSpec, WorkflowSpec
from retab.types.mime import FileRef, MIMEData
from retab.types.workflows.model import (
    DeclarativeApplyResponse,
    DeclarativeExportResponse,
    DeclarativePlanResponse,
    DeclarativeValidationResponse,
    WorkflowRun,
    WorkflowBlock,
    Workflow,
    WorkflowBlockCreateRequest,
    UpdateWorkflowBlockRequest,
    WorkflowEdgeCreateRequest,
    WorkflowGraphVersion,
    StepExecutionResponse,
    StoredBlockExecution,
)
from retab.types.workflows.steps import PublicHandlePayload

# Whole module is unit (pure offline; no server/credentials needed).
pytestmark = pytest.mark.unit

INVOICE_WORKFLOW_YAML = """apiVersion: workflows.retab.com/v1alpha2
kind: Workflow
metadata:
  id: wf_invoice_validation
  name: Invoice Validation Workflow
spec:
  blocks:
    start:
      type: start_json
      label: Invoice JSON
      config:
        json_schema:
          type: object
          properties:
            invoice_id:
              type: string
            line_items:
              type: array
              items:
                type: object
                properties:
                  description:
                    type: string
                  amount:
                    type: number
                required:
                  - description
                  - amount
            tax_rate:
              type: number
            stated_total:
              type: number
          required:
            - invoice_id
            - line_items
            - tax_rate
            - stated_total
    validate_total:
      type: function
      label: Validate Invoice Total
      config:
        output_schema:
          type: object
          properties:
            invoice_id:
              type: string
            subtotal:
              type: number
            computed_total:
              type: number
            is_valid:
              type: boolean
          required:
            - invoice_id
            - subtotal
            - computed_total
            - is_valid
        code: |
          from models import Input, Output

          def transform(input_data: Input) -> Output:
              subtotal = sum(item.amount for item in input_data.line_items)
              computed_total = round(subtotal + subtotal * input_data.tax_rate, 2)
              return Output(
                  invoice_id=input_data.invoice_id,
                  subtotal=subtotal,
                  computed_total=computed_total,
                  is_valid=abs(computed_total - input_data.stated_total) <= 0.01,
              )
  edges:
    - from:
        block: start
        handle: output-json-0
      to:
        block: validate_total
        handle: input-json-0
"""


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
    assert request.url == "/v1/workflows/workflow_123"
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
        after="workflow_before",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/v1/workflows"
    assert request.params == {
        "limit": 5,
        "order": "asc",
        "sort_by": "updated_at",
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
    assert request.url == "/v1/workflows/workflow_123"
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
        after="workflow_before",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/v1/workflows"
    assert request.params == {
        "limit": 5,
        "order": "asc",
        "sort_by": "updated_at",
        "after": "workflow_before",
    }
    assert result.list_metadata.after == "workflow_after"


def test_workflow_version_methods_use_version_routes() -> None:
    workflow_client = MagicMock()
    workflow_client._prepared_request.return_value = {
        "id": "ver_abcdefghijklmnopqrstuvwxyz012345",
        "workflow_id": "wrk_1",
        "organization_id": "org_1",
        "environment_id": "env_1",
        "blocks": [],
        "edges": [],
        "block_version_ids": [],
        "edge_version_ids": [],
        "created_at": "2026-01-01T00:00:00Z",
    }

    workflow_version = Workflows(client=workflow_client).get_version(
        "ver_abcdefghijklmnopqrstuvwxyz012345",
        workflow_id="wrk_1",
    )
    workflow_request = workflow_client._prepared_request.call_args.args[0]
    assert workflow_request.method == "GET"
    assert workflow_request.url == "/v1/workflows/versions/ver_abcdefghijklmnopqrstuvwxyz012345"
    assert workflow_request.params == {"workflow_id": "wrk_1"}
    assert isinstance(workflow_version, WorkflowGraphVersion)

    block_client = MagicMock()
    block_client._prepared_request.return_value = {
        "id": "bver_abcdefghijklmnopqrstuvwxyz012345",
        "block_id": "blk_1",
        "workflow_id": "wrk_1",
        "organization_id": "org_1",
        "environment_id": "env_1",
        "workflow_version_id": "ver_abcdefghijklmnopqrstuvwxyz012345",
        "type": "extract",
        "label": "Extract",
        "position_x": 0,
        "position_y": 0,
        "config_hash": "hash",
        "created_at": "2026-01-01T00:00:00Z",
    }
    block_version = WorkflowBlocks(client=block_client).get_version("bver_abcdefghijklmnopqrstuvwxyz012345")
    block_request = block_client._prepared_request.call_args.args[0]
    assert block_request.method == "GET"
    assert block_request.url == "/v1/workflows/blocks/versions/bver_abcdefghijklmnopqrstuvwxyz012345"
    assert isinstance(block_version, WorkflowBlockVersion)

    edge_client = MagicMock()
    edge_client._prepared_request.return_value = {
        "id": "ever_abcdefghijklmnopqrstuvwxyz012345",
        "edge_id": "edg_1",
        "workflow_id": "wrk_1",
        "organization_id": "org_1",
        "environment_id": "env_1",
        "workflow_version_id": "ver_abcdefghijklmnopqrstuvwxyz012345",
        "source": "blk_1",
        "target": "blk_2",
        "created_at": "2026-01-01T00:00:00Z",
    }
    edge_version = WorkflowEdges(client=edge_client).get_version("ever_abcdefghijklmnopqrstuvwxyz012345")
    edge_request = edge_client._prepared_request.call_args.args[0]
    assert edge_request.method == "GET"
    assert edge_request.url == "/v1/workflows/edges/versions/ever_abcdefghijklmnopqrstuvwxyz012345"
    assert isinstance(edge_version, WorkflowEdgeVersion)


def test_workflows_exposes_spec_subresource() -> None:
    workflows = Workflows(client=MagicMock())

    assert isinstance(workflows.spec, WorkflowSpec)


def test_async_workflows_exposes_spec_subresource() -> None:
    workflows = AsyncWorkflows(client=MagicMock())

    assert isinstance(workflows.spec, AsyncWorkflowSpec)


def test_workflows_blocks_exposes_executions_subresource() -> None:
    workflows = Workflows(client=MagicMock())

    assert isinstance(workflows.blocks.executions, WorkflowBlockExecutions)


def test_async_workflows_blocks_exposes_executions_subresource() -> None:
    workflows = AsyncWorkflows(client=MagicMock())

    assert isinstance(workflows.blocks.executions, AsyncWorkflowBlockExecutions)


def test_removed_workflow_methods_are_not_exposed() -> None:
    workflows = Workflows(client=MagicMock())

    assert not hasattr(workflows, "duplicate")
    assert not hasattr(workflows, "get_entities")
    assert not hasattr(workflows, "get_resolved_schemas")
    assert not hasattr(workflows, "list_snapshots")
    assert not hasattr(workflows.blocks, "config_history")
    assert not hasattr(workflows.blocks, "get_resolved_schemas")
    assert not hasattr(workflows.blocks, "create_batch")
    assert not hasattr(workflows.blocks, "list_executions")
    assert not hasattr(workflows.blocks, "execute")
    assert not hasattr(workflows.blocks.executions, "execute")
    assert not hasattr(workflows.edges, "create_batch")
    assert not hasattr(workflows.edges, "delete_all")
    assert not hasattr(workflows.experiments, "duplicate")
    assert not hasattr(workflows.experiments, "list_eligible_blocks")
    assert not hasattr(workflows.reviews, "wait_for")


def test_workflow_block_executions_create_uses_top_level_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "id": "sim_1",
        "workflow_id": "wf_1",
        "source_run_id": "run_1",
        "block_id": "block_1",
        "block_type": "extract",
        "lifecycle": {"status": "completed"},
        "created_at": "2026-03-12T10:00:00Z",
    }

    block_execution = Workflows(client=client).blocks.executions.create(
        run_id="run_1",
        block_id="block_1",
        step_id="step_1",
        n_consensus=5,
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/v1/workflows/blocks/executions"
    assert request.data == {
        "run_id": "run_1",
        "block_id": "block_1",
        "step_id": "step_1",
        "n_consensus": 5,
        "check_eligibility": True,
    }
    assert "workflow_id" not in request.data
    assert isinstance(block_execution, StoredBlockExecution)
    assert block_execution.id == "sim_1"


def test_workflow_block_executions_list_uses_top_level_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "data": [
            {
                "id": "sim_1",
                "workflow_id": "wf_1",
                "source_run_id": "run_1",
                "block_id": "block_1",
                "block_type": "extract",
                "lifecycle": {"status": "completed"},
                "created_at": "2026-03-12T10:00:00Z",
            }
        ],
        "list_metadata": {"before": None, "after": None},
    }

    result = Workflows(client=client).blocks.executions.list(
        run_id="run_1",
        block_id="block_1",
        limit=10,
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/v1/workflows/blocks/executions"
    assert request.params == {
        "run_id": "run_1",
        "block_id": "block_1",
        "limit": 10,
        "order": "desc",
    }
    assert result.data[0].id == "sim_1"


@pytest.mark.asyncio
async def test_async_workflow_block_executions_create_uses_top_level_route() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(
        return_value={
            "id": "sim_1",
            "workflow_id": "wf_1",
            "source_run_id": "run_1",
            "block_id": "block_1",
            "block_type": "extract",
            "lifecycle": {"status": "completed"},
            "created_at": "2026-03-12T10:00:00Z",
        }
    )

    block_execution = await AsyncWorkflows(client=client).blocks.executions.create(
        run_id="run_1",
        block_id="block_1",
        step_id="step_1",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/v1/workflows/blocks/executions"
    assert request.data == {
        "run_id": "run_1",
        "block_id": "block_1",
        "step_id": "step_1",
        "check_eligibility": True,
    }
    assert block_execution.id == "sim_1"


def test_workflow_spec_validate_uses_spec_validate_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "workflow_id": "wf_1",
        "block_count": 2,
        "edge_count": 1,
        "is_valid": True,
        "diagnostics": {"issues": []},
    }

    response = WorkflowSpec(client=client).validate(INVOICE_WORKFLOW_YAML)

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/v1/workflows/spec/validate"
    assert request.data == {"yaml_definition": INVOICE_WORKFLOW_YAML}
    assert isinstance(response, DeclarativeValidationResponse)
    assert response.is_valid is True


def test_workflow_spec_plan_uses_spec_plan_route() -> None:
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

    response = Workflows(client=client).plan(INVOICE_WORKFLOW_YAML)

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/v1/workflows/spec/plan"
    assert request.data == {"yaml_definition": INVOICE_WORKFLOW_YAML}
    assert isinstance(response, DeclarativePlanResponse)
    assert response.summary is not None
    assert response.summary.change == 1
    assert response.resource_changes is not None
    field_changes = response.resource_changes[0].change.field_changes
    assert field_changes is not None
    assert field_changes[0].path_display == "config.model"
    assert response.rendered_plan is not None
    assert "1 to change" in response.rendered_plan


def test_workflow_plan_with_workflow_id_uses_existing_workflow_route() -> None:
    """Passing workflow_id folds plan onto the existing-workflow route."""
    client = MagicMock()
    client._prepared_request.return_value = {
        "workflow_id": "wf_1",
        "action": "update",
        "block_count": 2,
        "edge_count": 1,
        "diagnostics": {"issues": []},
        "format_version": "workflows-plan/v1",
        "summary": {"add": 0, "change": 1, "destroy": 0, "replace": 0, "noop": 0, "total": 1, "has_changes": True},
        "resource_changes": [],
        "rendered_plan": "1 to change.",
    }

    response = Workflows(client=client).plan(INVOICE_WORKFLOW_YAML, workflow_id="wf_1")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/v1/workflows/wf_1/spec/plan"
    assert request.data == {"yaml_definition": INVOICE_WORKFLOW_YAML}
    assert isinstance(response, DeclarativePlanResponse)
    assert response.workflow_id == "wf_1"


def test_workflow_spec_apply_uses_spec_apply_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "workflow_id": "wf_1",
        "action": "noop",
        "created": False,
        "block_count": 2,
        "edge_count": 1,
        "diagnostics": {"issues": []},
        "format_version": "workflows-plan/v1",
        "summary": {"add": 0, "change": 0, "destroy": 0, "replace": 0, "noop": 1, "total": 0, "has_changes": False},
        "resource_changes": [],
        "rendered_plan": "No changes. Infrastructure is up-to-date.",
    }

    response = Workflows(client=client).apply(INVOICE_WORKFLOW_YAML)

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/v1/workflows/spec/apply"
    assert request.data == {"yaml_definition": INVOICE_WORKFLOW_YAML}
    assert isinstance(response, DeclarativeApplyResponse)
    assert response.action == "noop"
    assert response.summary is not None
    assert response.summary.noop == 1
    assert response.resource_changes == []


def test_workflow_apply_with_workflow_id_uses_nested_apply_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "workflow_id": "wf_1",
        "action": "update",
        "created": False,
        "block_count": 2,
        "edge_count": 1,
        "diagnostics": {"issues": []},
        "format_version": "workflows-plan/v1",
        "summary": {"add": 0, "change": 1, "destroy": 0, "replace": 0, "noop": 0, "total": 1, "has_changes": True},
        "resource_changes": [],
        "rendered_plan": "1 to change.",
    }

    response = Workflows(client=client).apply(INVOICE_WORKFLOW_YAML, workflow_id="wf_1")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/v1/workflows/wf_1/spec/apply"
    assert request.data == {"yaml_definition": INVOICE_WORKFLOW_YAML}
    assert isinstance(response, DeclarativeApplyResponse)
    assert response.workflow_id == "wf_1"
    assert response.created is False


def test_workflow_spec_get_uses_spec_resource_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "workflow_id": "wf_1",
        "yaml_definition": "apiVersion: workflows.retab.com/v1alpha2\n",
    }

    response = WorkflowSpec(client=client).get("wf_1")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/v1/workflows/wf_1/spec"
    assert isinstance(response, DeclarativeExportResponse)
    assert response.yaml_definition.startswith("apiVersion:")


@pytest.mark.asyncio
async def test_async_workflow_spec_validate_uses_spec_validate_route() -> None:
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

    response = await AsyncWorkflowSpec(client=client).validate("spec: {}\n")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/v1/workflows/spec/validate"
    assert response.workflow_id == "wf_1"


@pytest.mark.asyncio
async def test_async_workflow_apply_with_workflow_id_uses_nested_apply_route() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(
        return_value={
            "workflow_id": "wf_1",
            "action": "update",
            "created": False,
            "block_count": 2,
            "edge_count": 1,
            "diagnostics": {"issues": []},
            "format_version": "workflows-plan/v1",
            "summary": {"add": 0, "change": 1, "destroy": 0, "replace": 0, "noop": 0, "total": 1, "has_changes": True},
            "resource_changes": [],
            "rendered_plan": "1 to change.",
        }
    )

    response = await AsyncWorkflows(client=client).apply(INVOICE_WORKFLOW_YAML, workflow_id="wf_1")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/v1/workflows/wf_1/spec/apply"
    assert request.data == {"yaml_definition": INVOICE_WORKFLOW_YAML}
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
            "workflow_id": "workflow_123",
            "workflow_version_id": "ver_0123456789abcdef0123456789abcdef",
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
    """v2 typed fields parse and round-trip with the typed sub-models."""
    run = WorkflowRun.model_validate(
        {
            "id": "run_789",
            "organization_id": "org_1",
            "workflow_id": "wf_1",
            "workflow_version_id": "ver_abcdef0123456789abcdef0123456789",
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
    assert run.workflow_id == "wf_1"
    assert run.workflow_version_id == "ver_abcdef0123456789abcdef0123456789"
    assert run.trigger is not None
    assert run.trigger.type == "api"
    assert run.lifecycle is not None
    assert run.lifecycle.status == "completed"
    assert run.inputs is not None
    assert run.inputs.json_data == {"json-1": {"key": "value"}}
    assert run.timing is not None
    assert run.timing.completed_at is not None
    # Legacy timing fields are silently dropped by the three-timestamp model.
    dumped = run.timing.model_dump()
    assert "accumulated_review_waiting_ms" not in dumped
    assert "review_waiting_started_at" not in dumped
    assert "duration_ms" not in dumped
    assert "active_duration_ms" not in dumped

    # Defaults: omitted inputs fall back to an empty RunInputs.
    run2 = WorkflowRun.model_validate(
        {
            "id": "run_000",
            "organization_id": "org_1",
            "workflow_id": "wf_1",
            "workflow_version_id": "ver_0123456789abcdef0123456789abcdef",
            "trigger": {"type": "manual"},
            "lifecycle": {"status": "pending"},
            "timing": {"created_at": "2026-01-01T00:00:00Z"},
        }
    )
    assert run2.inputs is not None
    assert run2.inputs.documents == {}
    assert run2.inputs.json_data == {}
    assert run2.timing is not None
    assert run2.timing.created_at is not None


def test_handle_payload_rejects_removed_text_type() -> None:
    with pytest.raises(ValidationError):
        PublicHandlePayload.model_validate({"type": "text", "text": "removed"})


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

    wf = Workflows(client=client).create(project_id="proj_test", name="My Workflow", description="A test")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/v1/workflows"
    assert request.data == {"project_id": "proj_test", "name": "My Workflow", "description": "A test"}
    assert wf.id == "wf_new"


def test_workflows_update_route() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "id": "workflow_123",
        "name": "Renamed Workflow",
        "description": "Updated",
        "published": None,
        "created_at": "2026-03-12T10:00:00Z",
        "updated_at": "2026-03-12T10:00:00Z",
    }

    workflow = Workflows(client=client).update(
        "workflow_123",
        name="Renamed Workflow",
        description="Updated",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "PATCH"
    assert request.url == "/v1/workflows/workflow_123"
    assert request.data == {
        "name": "Renamed Workflow",
        "description": "Updated",
    }
    assert workflow.name == "Renamed Workflow"


def test_workflow_blocks_create_accepts_typed_request() -> None:
    client = MagicMock()
    client._prepared_request.return_value = {
        "id": "extract-1",
        "workflow_id": "wf_1",
        "organization_id": "org_1",
        "draft_version": "draft_1",
        "type": "extract",
        "label": "Extract",
        "updated_at": "2026-03-12T10:00:00Z",
    }

    block = WorkflowBlocks(client=client).create(
        workflow_id="wf_1",
        id="extract-1",
        type="extract",
        label="Extract",
        position_x=120,
        position_y=80,
        config={"json_schema": {"type": "object"}},
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/v1/workflows/blocks"
    assert request.data == {
        "workflow_id": "wf_1",
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
        "updated_at": "2026-03-12T10:00:00Z",
    }

    block = WorkflowBlocks(client=client).update(
        block_id="extract-1",
        label="Renamed",
        position_x=200,
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "PATCH"
    assert request.url == "/v1/workflows/blocks/extract-1"
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
            "updated_at": "2026-03-12T10:00:00Z",
        }
    )

    # ``draft_version`` is server-internal metadata stripped from the public
    # spec, so the SDK ignores it on parse without erroring.
    assert block.id == "extract-1"
    assert block.type == "extract"


def test_workflow_block_exposes_resolved_schema_sidecar() -> None:
    block = WorkflowBlock.model_validate(
        {
            "id": "extract-1",
            "workflow_id": "wf_1",
            "organization_id": "org_1",
            "draft_version": "draft_1",
            "type": "extract",
            "label": "Extract",
            "updated_at": "2026-03-12T10:00:00Z",
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
    assert block.resolved_schemas["output_schemas"]["output-json-0"]["properties"]["invoice_number"]["type"] == "string"


def test_step_execution_response_ignores_removed_payload_schema_fields() -> None:
    removed_payload_key = "raw" + "_" + "output"
    step_execution = StepExecutionResponse.model_validate(
        {
            "id": "sim_1",
            "workflow_id": "wf_1",
            "source_run_id": "run_1",
            "block_id": "extract-1",
            "block_type": "extract",
            "block_label": "Extract",
            "lifecycle": {"status": "completed"},
            "created_at": "2026-01-01T00:00:00+00:00",
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
    assert step_execution.handle_outputs is not None
    assert step_execution.handle_outputs["output-json-0"]["data"] == {"invoice_number": "INV-001"}


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
        "updated_at": "2026-03-12T10:00:00Z",
    }

    edge = WorkflowEdges(client=client).create(
        workflow_id="wf_1",
        id="edge-1",
        source_block="start-1",
        target_block="extract-1",
        source_handle="output-file-0",
        target_handle="input-file-0",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/v1/workflows/edges"
    assert request.data == {
        "workflow_id": "wf_1",
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
    assert request.url == "/v1/workflows/wf_1/publish"
    assert wf.published is not None
    assert wf.published.version_id == "ver_0123456789abcdef0123456789abcdef"


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
        trigger_type="api",
        from_date=date(2026, 1, 1),
        to_date=date(2026, 1, 31),
        after="run_after",
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/v1/workflows/runs"
    assert request.params["workflow_id"] == "wf_1"
    assert request.params["statuses"] == ["completed", "error"]
    assert request.params["trigger_type"] == "api"
    assert request.params["from_date"] == date(2026, 1, 1)
    assert request.params["to_date"] == date(2026, 1, 31)
    assert "fields" not in request.params
    assert request.params["after"] == "run_after"


def test_workflow_runs_create_without_inputs_sends_json_body() -> None:
    request = WorkflowRuns(client=MagicMock()).prepare_create(workflow_id="wf_1")

    assert request.method == "POST"
    assert request.url == "/v1/workflows/runs"
    assert request.data == {
        "workflow_id": "wf_1",
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


def test_workflow_runs_create_keeps_mime_data_url_for_new_documents() -> None:
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
            "url": "data:text/plain;base64,aGVsbG8=",
        }
    }


def _v2_run_payload(**overrides) -> dict:
    """Helper: build a minimal v2 WorkflowRun JSON payload for fixtures."""
    payload: dict = {
        "id": overrides.pop("id", "run_1"),
        "organization_id": "org_1",
        "workflow_id": "wf_1",
        "workflow_version_id": "ver_0123456789abcdef0123456789abcdef",
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
    assert request.url == "/v1/workflows/runs/run_1/cancel"
    assert request.data == {"command_id": "cmd_1"}
    assert result.cancellation_status == "cancelled"
    assert result.run.id == "run_1"


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
        trigger_type="api",
        preferred_columns=["invoice_number", "total_amount"],
    )

    request = client._prepared_request.call_args.args[0]
    assert request.method == "POST"
    assert request.url == "/v1/workflows/runs/export"
    assert request.data == {
        "workflow_id": "wf_1",
        "block_id": "extract-1",
        "export_source": "outputs",
        "preferred_columns": ["invoice_number", "total_amount"],
        "delimiter": ";",
        "quote": '"',
        "line_delimiter": "\n",
        "selected_run_ids": ["run_1", "run_2"],
        "trigger_type": "api",
    }
    assert result.rows == 1
    assert result.columns == 2


def test_workflow_runs_do_not_expose_wait_for_completion() -> None:
    assert not hasattr(WorkflowRuns(client=MagicMock()), "wait_for_completion")
    assert not hasattr(AsyncWorkflowRuns(client=MagicMock()), "wait_for_completion")


def test_workflow_run_step_handle_outputs_data() -> None:
    """WorkflowRunStep exposes extracted JSON through typed handle_outputs."""
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
    assert step.handle_outputs is not None
    assert step.handle_outputs["output-json-0"].data == {"total": 1234}
    dumped = step.model_dump()
    assert "status" not in dumped
    assert "terminal" not in dumped
    # handle_outputs should be typed as PublicHandlePayload

    assert step.handle_outputs is not None
    payload = step.handle_outputs["output-json-0"]
    assert isinstance(payload, PublicHandlePayload)
    assert payload.type == "json"
    assert "input_document" not in WorkflowRunStep.model_fields
    assert "output_document" not in WorkflowRunStep.model_fields
    assert "split_documents" not in WorkflowRunStep.model_fields
    assert "status" not in WorkflowRunStep.model_fields
    assert "terminal" not in WorkflowRunStep.model_fields


def test_workflow_run_step_rejects_private_json_ref_handle_payload() -> None:
    from retab.types.workflows.model import WorkflowRunStep

    with pytest.raises(ValidationError):
        WorkflowRunStep.model_validate(
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
