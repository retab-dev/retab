from unittest.mock import AsyncMock, MagicMock

import pytest

from retab.resources.workflows.client import AsyncWorkflows, Workflows
from retab.types.workflows.model import WorkflowRun


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
