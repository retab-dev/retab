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
    assert request.url == "/v1/workflows/runs/run_123/steps"
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
    assert request.url == "/v1/workflows/runs/run_123/steps"
    assert len(steps) == 1
    assert steps[0].node_id == "extract-1"
