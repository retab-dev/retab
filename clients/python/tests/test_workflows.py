from unittest.mock import AsyncMock, MagicMock

import pytest

from retab.resources.workflows.client import AsyncWorkflows, Workflows


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
