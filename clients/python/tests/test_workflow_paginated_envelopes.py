"""Regression tests for the canonical paginated envelope on workflow list endpoints.

Pins the wire-shape contract for the current endpoints converted from bare
arrays to ``PaginatedList[T]`` envelopes:

    GET /v1/workflows/blocks?workflow_id={wf}                                       -> PaginatedList[WorkflowBlock]
    GET /v1/workflows/edges?workflow_id={wf}                                        -> PaginatedList[WorkflowEdgeDoc]
    GET /v1/workflows/artifacts                                         -> PaginatedList[WorkflowArtifact]
    GET /v1/workflows/steps?run_id={run_id}                               -> PaginatedList[WorkflowRunStep]
"""

from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock

import pytest

from retab.resources.workflows.artifacts import WorkflowArtifacts
from retab.resources.workflows.blocks import (
    AsyncWorkflowBlocks,
    WorkflowBlocks,
)
from retab.resources.workflows.edges import WorkflowEdges
from retab.types.pagination import AsyncPaginatedList, PaginatedList

# Whole module is unit (pure offline; no server/credentials needed).
pytestmark = pytest.mark.unit


def _envelope(*items: dict) -> dict:
    return {"data": list(items), "list_metadata": {"before": None, "after": None}}


def test_workflow_blocks_list_returns_paginated_envelope() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _envelope(
        {
            "id": "extract-1",
            "workflow_id": "wf_aaa",
            "type": "extract",
            "label": "Extract",
            "position_x": 0,
            "position_y": 0,
            "updated_at": "2026-01-01T00:00:00Z",
        }
    )

    result = WorkflowBlocks(client=client).list("wf_aaa")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/v1/workflows/blocks"
    assert request.params == {"workflow_id": "wf_aaa", "limit": 100}
    assert isinstance(result, PaginatedList)
    assert len(result) == 1
    assert result[0].id == "extract-1"
    assert result.list_metadata.before is None
    assert result.list_metadata.after is None


def test_workflow_edges_list_returns_paginated_envelope() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _envelope(
        {
            "id": "edge-1",
            "workflow_id": "wf_aaa",
            "organization_id": "org_1",
            "source_block": "start-1",
            "target_block": "extract-1",
            "source_handle": "output-file-0",
            "target_handle": "input-file-0",
            "updated_at": "2026-01-01T00:00:00Z",
        }
    )

    result = WorkflowEdges(client=client).list("wf_aaa")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/v1/workflows/edges"
    assert request.params == {"workflow_id": "wf_aaa", "limit": 100}
    assert isinstance(result, PaginatedList)
    assert len(result) == 1
    assert result[0].id == "edge-1"
    assert result.list_metadata.before is None
    assert result.list_metadata.after is None


def test_workflow_artifacts_list_returns_paginated_envelope() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _envelope(
        {"operation": "extraction", "id": "ext_123"},
        {"operation": "parse", "id": "parse_456"},
    )

    result = WorkflowArtifacts(client=client).list("run_aaa")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/v1/workflows/artifacts"
    assert isinstance(result, PaginatedList)
    assert len(result) == 2
    assert result[0].operation == "extraction"
    assert result.list_metadata.before is None
    assert result.list_metadata.after is None


@pytest.mark.asyncio
async def test_async_workflow_blocks_list_returns_paginated_envelope() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(
        return_value=_envelope(
            {
                "id": "extract-1",
                "workflow_id": "wf_aaa",
                "type": "extract",
                "label": "Extract",
                "position_x": 0,
                "position_y": 0,
                "updated_at": "2026-01-01T00:00:00Z",
            }
        )
    )

    result = await AsyncWorkflowBlocks(client=client).list("wf_aaa")

    assert isinstance(result, AsyncPaginatedList)
    assert len(result) == 1
    assert result[0].id == "extract-1"
