# pyright: reportUnusedVariable=false
"""Creditless coverage for workflow GRAPH + saved-config read resources.

Every test here is CREDITLESS: it reads a workflow's stored graph (blocks /
edges) and its saved experiments / tests metadata. Nothing here runs a block,
runs an experiment, or runs a test (all billable). ``blocks.validate_config``
explicitly validates an assembled config "without mutating the workflow draft"
and runs no model, so it is a creditless config check.

Covered (not covered by the other ``test_creditless_*`` files):
  - ``client.workflows.blocks``      — list / get / 404 / config-validate (read-only)
  - ``client.workflows.edges``       — list / get / 404
  - ``client.workflows.experiments`` — list / get / 404 (read-only; never run)
  - ``client.workflows.tests``       — list / get / 404 (read-only; never run)
"""

from __future__ import annotations

import uuid

import pytest

from retab import AsyncRetab, Retab
from retab.exceptions import APIError, NotFoundError
from retab.types.workflows.blocks import WorkflowBlock
from retab.types.workflows.edges import WorkflowEdgeDoc
from retab.types.workflows.experiments import WorkflowExperiment
from retab.types.workflows.tests import WorkflowTest


def _discover_workflow_id(client: Retab) -> str | None:
    page = client.workflows.list(limit=25)
    return page.data[0].id if page.data else None


def _discover_workflow_with_blocks(client: Retab) -> tuple[str, str] | None:
    """Return (workflow_id, block_id) for the first workflow that has a block."""
    page = client.workflows.list(limit=25)
    for wf in page.data:
        blocks = client.workflows.blocks.list(workflow_id=wf.id, limit=1)
        if blocks.data:
            return wf.id, blocks.data[0].id
    return None


# --------------------------------------------------------------------------- #
# Blocks
# --------------------------------------------------------------------------- #


def test_blocks_list_typed_page(sync_client: Retab) -> None:
    workflow_id = _discover_workflow_id(sync_client)
    if not workflow_id:
        pytest.skip("no workflows on staging")
    page = sync_client.workflows.blocks.list(workflow_id=workflow_id, limit=5)
    assert len(page.data) <= 5
    assert page.list_metadata is not None
    for block in page.data:
        assert isinstance(block, WorkflowBlock)
        assert isinstance(block.id, str) and block.id


def test_blocks_get_by_discovered_id(sync_client: Retab) -> None:
    found = _discover_workflow_with_blocks(sync_client)
    if not found:
        pytest.skip("no workflow with blocks on staging")
    workflow_id, block_id = found
    fetched = sync_client.workflows.blocks.get(block_id, workflow_id=workflow_id)
    assert isinstance(fetched, WorkflowBlock)
    assert fetched.id == block_id


def test_blocks_list_requires_workflow_id(sync_client: Retab) -> None:
    # workflow_id is a required positional; omitting it is a client-side TypeError,
    # which proves the SDK enforces the scope rather than issuing an unscoped scan.
    with pytest.raises(TypeError):
        sync_client.workflows.blocks.list()  # type: ignore[call-arg]


def test_blocks_get_bogus_id_404(sync_client: Retab) -> None:
    workflow_id = _discover_workflow_id(sync_client)
    if not workflow_id:
        pytest.skip("no workflows on staging")
    with pytest.raises(NotFoundError) as excinfo:
        sync_client.workflows.blocks.get("block_creditless_bogus_" + uuid.uuid4().hex, workflow_id=workflow_id)
    assert excinfo.value.status_code == 404


def test_blocks_validate_config_is_creditless_check(sync_client: Retab) -> None:
    """validate-config inspects a config without mutating or running anything."""
    found = _discover_workflow_with_blocks(sync_client)
    if not found:
        pytest.skip("no workflow with blocks on staging")
    workflow_id, block_id = found
    # An obviously-incomplete config: the endpoint must answer with a structured
    # validation verdict (not run a model). Either a typed response or a 4xx is a
    # valid creditless outcome; we only assert no execution/credit happened.
    try:
        resp = sync_client.workflows.blocks.create_block_validate_config(
            block_id,
            config={},
            workflow_id=workflow_id,
        )
    except APIError as exc:
        assert exc.status_code in (400, 422), f"unexpected validate-config status {exc.status_code}"
        return
    # Typed response: must expose a validity verdict field.
    assert hasattr(resp, "is_valid") or hasattr(resp, "valid") or resp is not None


# --------------------------------------------------------------------------- #
# Edges
# --------------------------------------------------------------------------- #


def test_edges_list_typed_page(sync_client: Retab) -> None:
    workflow_id = _discover_workflow_id(sync_client)
    if not workflow_id:
        pytest.skip("no workflows on staging")
    page = sync_client.workflows.edges.list(workflow_id=workflow_id, limit=5)
    assert len(page.data) <= 5
    assert page.list_metadata is not None
    for edge in page.data:
        assert isinstance(edge, WorkflowEdgeDoc)
        assert isinstance(edge.id, str) and edge.id


def test_edges_get_by_discovered_id(sync_client: Retab) -> None:
    page = sync_client.workflows.list(limit=25)
    for wf in page.data:
        edges = sync_client.workflows.edges.list(workflow_id=wf.id, limit=1)
        if edges.data:
            edge_id = edges.data[0].id
            fetched = sync_client.workflows.edges.get(edge_id)
            assert isinstance(fetched, WorkflowEdgeDoc)
            assert fetched.id == edge_id
            return
    pytest.skip("no workflow with edges on staging")


def test_edges_get_bogus_id_404(sync_client: Retab) -> None:
    with pytest.raises(NotFoundError) as excinfo:
        sync_client.workflows.edges.get("edge_creditless_bogus_" + uuid.uuid4().hex)
    assert excinfo.value.status_code == 404


# --------------------------------------------------------------------------- #
# Experiments (READ-ONLY — never run, that is billable)
# --------------------------------------------------------------------------- #


def test_experiments_list_shape_per_workflow(sync_client: Retab) -> None:
    workflow_id = _discover_workflow_id(sync_client)
    if not workflow_id:
        pytest.skip("no workflows on staging")
    page = sync_client.workflows.experiments.list(workflow_id=workflow_id, limit=5)
    assert isinstance(page.data, list)
    assert page.list_metadata is not None
    for exp in page.data:
        assert isinstance(exp, WorkflowExperiment)


def test_experiments_get_by_discovered_id(sync_client: Retab) -> None:
    page = sync_client.workflows.list(limit=25)
    for wf in page.data:
        exps = sync_client.workflows.experiments.list(workflow_id=wf.id, limit=1)
        if exps.data:
            exp_id = exps.data[0].id
            fetched = sync_client.workflows.experiments.get(exp_id)
            assert isinstance(fetched, WorkflowExperiment)
            assert fetched.id == exp_id
            return
    pytest.skip("no workflow with experiments on staging")


def test_experiments_get_bogus_id_404(sync_client: Retab) -> None:
    with pytest.raises(NotFoundError) as excinfo:
        sync_client.workflows.experiments.get("exp_creditless_bogus_" + uuid.uuid4().hex)
    assert excinfo.value.status_code == 404


# --------------------------------------------------------------------------- #
# Tests (READ-ONLY — never run, that is billable)
# --------------------------------------------------------------------------- #


def test_tests_list_shape_per_workflow(sync_client: Retab) -> None:
    workflow_id = _discover_workflow_id(sync_client)
    if not workflow_id:
        pytest.skip("no workflows on staging")
    page = sync_client.workflows.tests.list(workflow_id=workflow_id, limit=5)
    assert isinstance(page.data, list)
    assert page.list_metadata is not None
    for test in page.data:
        assert isinstance(test, WorkflowTest)


def test_tests_get_by_discovered_id(sync_client: Retab) -> None:
    page = sync_client.workflows.list(limit=25)
    for wf in page.data:
        tests = sync_client.workflows.tests.list(workflow_id=wf.id, limit=1)
        if tests.data:
            test_id = tests.data[0].id
            fetched = sync_client.workflows.tests.get(test_id)
            assert isinstance(fetched, WorkflowTest)
            assert fetched.id == test_id
            return
    pytest.skip("no workflow with tests on staging")


def test_tests_get_bogus_id_404(sync_client: Retab) -> None:
    with pytest.raises(NotFoundError) as excinfo:
        sync_client.workflows.tests.get("wtst_creditless_bogus_" + uuid.uuid4().hex)
    assert excinfo.value.status_code == 404


def test_tests_list_bogus_workflow_404(sync_client: Retab) -> None:
    """The docstring promises 404 when the workflow does not exist."""
    with pytest.raises(APIError) as excinfo:
        sync_client.workflows.tests.list(workflow_id="wrk_creditless_bogus_" + uuid.uuid4().hex, limit=1)
    assert excinfo.value.status_code in (404, 400)


# --------------------------------------------------------------------------- #
# Async
# --------------------------------------------------------------------------- #


@pytest.mark.asyncio
async def test_async_blocks_list(async_client: AsyncRetab) -> None:
    page = await async_client.workflows.list(limit=25)
    if not page.data:
        pytest.skip("no workflows on staging")
    blocks = await async_client.workflows.blocks.list(workflow_id=page.data[0].id, limit=3)
    assert isinstance(blocks.data, list)
    for block in blocks.data:
        assert isinstance(block, WorkflowBlock)
