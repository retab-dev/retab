"""Regression tests for the canonical paginated envelope on workflow GRAPH list endpoints.

Pins the wire-shape contract for the seven endpoints converted from bare
arrays to ``PaginatedList[T]`` envelopes:

    GET /v1/workflows/{wf}/blocks                                       -> PaginatedList[WorkflowBlock]
    GET /v1/workflows/{wf}/blocks/{id}/config-history                   -> PaginatedList[BlockConfigVersion]
    GET /v1/workflows/{wf}/edges                                        -> PaginatedList[WorkflowEdgeDoc]
    GET /v1/workflows/artifacts                                         -> PaginatedList[WorkflowArtifact]
    GET /v1/workflows/{wf}/snapshots                                    -> PaginatedList[WorkflowSnapshot]
    GET /v1/workflows/runs/{run_id}/steps                               -> PaginatedList[WorkflowRunStep]
    GET /v1/workflows/runs/{run_id}/steps/{block_id}/simulations        -> PaginatedList[BlockSimulation]
"""

from __future__ import annotations

import datetime
from unittest.mock import AsyncMock, MagicMock

import pytest

from retab.resources.workflows.artifacts.client import WorkflowArtifacts
from retab.resources.workflows.blocks.client import (
    AsyncWorkflowBlocks,
    WorkflowBlocks,
)
from retab.resources.workflows.client import AsyncWorkflows, Workflows
from retab.resources.workflows.edges.client import WorkflowEdges
from retab.types.pagination import PaginatedList


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
        }
    )

    result = WorkflowBlocks(client=client).list("wf_aaa")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/wf_aaa/blocks"
    assert isinstance(result, PaginatedList)
    assert len(result) == 1
    assert result[0].id == "extract-1"
    assert result.list_metadata.before is None
    assert result.list_metadata.after is None


def test_workflow_blocks_config_history_returns_paginated_envelope() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _envelope(
        {
            "config_fingerprint": "fp_1",
            "block_type": "extract",
            "block_label": "Extract",
            "config_snapshot": {"model": "gpt-5"},
            "first_seen_at": "2026-01-01T00:00:00Z",
            "last_seen_at": "2026-01-02T00:00:00Z",
            "snapshot_versions": [1, 2],
            "run_count": 7,
            "is_current": True,
        }
    )

    result = WorkflowBlocks(client=client).config_history("wf_aaa", "extract-1")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/wf_aaa/blocks/extract-1/config-history"
    assert isinstance(result, PaginatedList)
    assert len(result) == 1
    assert result[0].config_fingerprint == "fp_1"
    assert result[0].run_count == 7
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
        }
    )

    result = WorkflowEdges(client=client).list("wf_aaa")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/wf_aaa/edges"
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
    assert request.url == "/workflows/artifacts"
    assert isinstance(result, PaginatedList)
    assert len(result) == 2
    assert result[0].operation == "extraction"
    assert result.list_metadata.before is None
    assert result.list_metadata.after is None


def test_workflow_snapshots_list_returns_paginated_envelope() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _envelope(
        {
            "id": "wfsn_1",
            "snapshot_id": "snap_1",
            "workflow_id": "wf_aaa",
            "version": 3,
            "description": "third release",
            "block_count": 4,
            "edge_count": 3,
            "published_at": "2026-04-01T12:00:00Z",
        }
    )

    result = Workflows(client=client).list_snapshots("wf_aaa")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/wf_aaa/snapshots"
    assert isinstance(result, PaginatedList)
    assert len(result) == 1
    assert result[0].snapshot_id == "snap_1"
    assert result[0].version == 3
    assert result.list_metadata.before is None
    assert result.list_metadata.after is None


def test_workflow_simulations_list_returns_paginated_envelope() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _envelope(
        {
            "id": "sim_1",
            "workflow_id": "wf_aaa",
            "run_id": "run_aaa",
            "block_id": "extract-1",
            "block_type": "extract",
            "success": True,
        }
    )

    result = WorkflowBlocks(client=client).list_simulations("run_aaa", "extract-1")

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/workflows/runs/run_aaa/steps/extract-1/simulations"
    assert isinstance(result, PaginatedList)
    assert len(result) == 1
    assert result[0].id == "sim_1"
    assert result[0].success is True
    assert result.list_metadata.before is None
    assert result.list_metadata.after is None


def test_workflow_simulations_list_propagates_limit_param() -> None:
    client = MagicMock()
    client._prepared_request.return_value = _envelope()

    WorkflowBlocks(client=client).list_simulations("run_aaa", "extract-1", limit=42)

    request = client._prepared_request.call_args.args[0]
    assert request.params == {"limit": 42}


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
            }
        )
    )

    result = await AsyncWorkflowBlocks(client=client).list("wf_aaa")

    assert isinstance(result, PaginatedList)
    assert len(result) == 1
    assert result[0].id == "extract-1"


@pytest.mark.asyncio
async def test_async_workflow_snapshots_list_returns_paginated_envelope() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(
        return_value=_envelope(
            {
                "id": "wfsn_1",
                "snapshot_id": "snap_1",
                "workflow_id": "wf_aaa",
                "version": 1,
                "published_at": "2026-04-01T12:00:00Z",
            }
        )
    )

    result = await AsyncWorkflows(client=client).list_snapshots("wf_aaa", limit=10)

    request = client._prepared_request.call_args.args[0]
    assert request.params == {"limit": 10}
    assert isinstance(result, PaginatedList)
    assert result[0].version == 1


# Quiet linter on imports kept around for clarity / future use.
_ = datetime
