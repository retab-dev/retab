# pyright: reportUnusedVariable=false
"""Creditless coverage for workflow RUNTIME read resources.

Every test here is CREDITLESS: it only lists / gets pre-existing runtime
records produced by historical workflow runs. Nothing here creates a run,
resumes a run, approves/rejects a review, or otherwise drives execution — all of
which would be billable. We only read.

Covered (not covered by the other ``test_creditless_*`` files):
  - ``client.workflows.runs``      — list / get / pagination / filters / 404
  - ``client.workflows.steps``     — list / get-by (step_id, run_id) / 404
  - ``client.workflows.artifacts`` — list (run-scoped) / get / 404 / 400-without-scope
  - ``client.workflows.reviews``   — list / get / 404 (read-only; NEVER approve/reject)
"""

from __future__ import annotations

import uuid

import pytest


from retab import AsyncRetab, Retab
from retab.exceptions import APIError, AuthenticationError, NotFoundError
from retab.types.classifications import WorkflowRunsStatus
from retab.types.workflows.artifacts import WorkflowArtifact
from retab.types.workflows.reviews import Review
from retab.types.workflows.runs import WorkflowRun
from retab.types.workflows.steps import WorkflowRunStep

# Whole module is creditless (storage/config/list/get/error paths only).
pytestmark = pytest.mark.creditless


# --------------------------------------------------------------------------- #
# Workflow runs
# --------------------------------------------------------------------------- #


def test_runs_list_typed_page_and_envelope(sync_client: Retab) -> None:
    page = sync_client.workflows.runs.list(limit=3)
    assert len(page.data) <= 3
    assert page.list_metadata is not None
    for run in page.data:
        assert isinstance(run, WorkflowRun)
        assert isinstance(run.id, str) and run.id.startswith("run_")


def test_runs_pagination_after_cursor_disjoint(sync_client: Retab) -> None:
    page1 = sync_client.workflows.runs.list(limit=2, order="desc")
    after = page1.list_metadata.after
    if after is None or len(page1.data) < 2:
        pytest.skip("not enough workflow runs on staging to page")
    page2 = sync_client.workflows.runs.list(limit=2, order="desc", after=after)
    ids1 = {r.id for r in page1.data}
    ids2 = {r.id for r in page2.data}
    assert ids1.isdisjoint(ids2), "runs cursor returned overlapping rows"


def test_runs_status_filter_shape(sync_client: Retab) -> None:
    page = sync_client.workflows.runs.list(limit=5, status=WorkflowRunsStatus.COMPLETED)
    assert isinstance(page.data, list)
    for run in page.data:
        assert isinstance(run, WorkflowRun)


def test_runs_from_date_yyyy_mm_dd_accepted(sync_client: Retab) -> None:
    page = sync_client.workflows.runs.list(limit=3, from_date="2020-01-01")
    assert isinstance(page.data, list)


def test_runs_get_by_discovered_id(sync_client: Retab) -> None:
    page = sync_client.workflows.runs.list(limit=10)
    if not page.data:
        pytest.skip("no workflow runs on staging to get-by-id")
    target_id = page.data[0].id
    fetched = sync_client.workflows.runs.get(target_id)
    assert isinstance(fetched, WorkflowRun)
    assert fetched.id == target_id


def test_runs_get_bogus_id_404(sync_client: Retab) -> None:
    with pytest.raises(NotFoundError) as excinfo:
        sync_client.workflows.runs.get("run_creditless_bogus_" + uuid.uuid4().hex)
    assert excinfo.value.status_code == 404


# --------------------------------------------------------------------------- #
# Workflow run steps
# --------------------------------------------------------------------------- #


def test_steps_list_typed_page(sync_client: Retab) -> None:
    page = sync_client.workflows.steps.list(limit=3)
    assert len(page.data) <= 3
    assert page.list_metadata is not None
    for step in page.data:
        assert isinstance(step, WorkflowRunStep)
        # WorkflowRunStep is keyed by (step_id, run_id), not a single ``id``.
        assert isinstance(step.step_id, str) and step.step_id
        assert isinstance(step.run_id, str) and step.run_id


def test_steps_get_by_discovered_step_and_run(sync_client: Retab) -> None:
    page = sync_client.workflows.steps.list(limit=10)
    if not page.data:
        pytest.skip("no workflow run steps on staging to get-by-id")
    step = page.data[0]
    fetched = sync_client.workflows.steps.get(step_id=step.step_id, run_id=step.run_id)
    assert isinstance(fetched, WorkflowRunStep)
    assert fetched.step_id == step.step_id
    assert fetched.run_id == step.run_id


def test_steps_scoped_to_run_belong_to_that_run(sync_client: Retab) -> None:
    runs = sync_client.workflows.runs.list(limit=10)
    if not runs.data:
        pytest.skip("no workflow runs on staging to scope steps")
    run_id = runs.data[0].id
    steps = sync_client.workflows.steps.list(run_id=run_id, limit=10)
    for step in steps.data:
        assert step.run_id == run_id, "run-scoped step list returned a foreign run_id"


def test_steps_get_bogus_404(sync_client: Retab) -> None:
    with pytest.raises(APIError) as excinfo:
        sync_client.workflows.steps.get(
            step_id="block_creditless_bogus_" + uuid.uuid4().hex,
            run_id="run_creditless_bogus_" + uuid.uuid4().hex,
        )
    assert excinfo.value.status_code in (404, 400)


# --------------------------------------------------------------------------- #
# Workflow artifacts
# --------------------------------------------------------------------------- #


def test_artifacts_list_requires_scope_400(sync_client: Retab) -> None:
    """Listing artifacts without ``run_id``/``step_id`` is a clean 400, not a scan."""
    with pytest.raises(APIError) as excinfo:
        sync_client.workflows.artifacts.list(limit=3)
    assert excinfo.value.status_code == 400


def test_artifacts_list_run_scoped_shape(sync_client: Retab) -> None:
    runs = sync_client.workflows.runs.list(limit=10)
    if not runs.data:
        pytest.skip("no workflow runs on staging to scope artifacts")
    run_id = runs.data[0].id
    page = sync_client.workflows.artifacts.list(run_id=run_id, limit=5)
    assert isinstance(page.data, list)
    for artifact in page.data:
        assert isinstance(artifact, WorkflowArtifact)


def test_artifacts_get_run_scoped_discovered_id(sync_client: Retab) -> None:
    runs = sync_client.workflows.runs.list(limit=15)
    target_artifact_id: str | None = None
    for run in runs.data:
        page = sync_client.workflows.artifacts.list(run_id=run.id, limit=5)
        for artifact in page.data:
            aid = getattr(artifact, "id", None)
            if isinstance(aid, str) and aid:
                target_artifact_id = aid
                break
        if target_artifact_id:
            break
    if not target_artifact_id:
        pytest.skip("no workflow artifacts discoverable across recent staging runs")
    # ``artifacts.get`` returns a concrete union member (e.g. ExtractionWorkflowArtifact)
    # keyed off the id prefix — NOT the base ``WorkflowArtifact`` model that ``list``
    # returns. Assert on the shared identity fields the concrete members carry.
    fetched = sync_client.workflows.artifacts.get(target_artifact_id)
    assert getattr(fetched, "id", None) == target_artifact_id
    assert getattr(fetched, "operation", None)


def test_artifacts_get_bogus_id_404(sync_client: Retab) -> None:
    with pytest.raises(NotFoundError) as excinfo:
        sync_client.workflows.artifacts.get("extr_creditless_bogus_" + uuid.uuid4().hex)
    assert excinfo.value.status_code == 404


# --------------------------------------------------------------------------- #
# Workflow reviews (READ-ONLY — never approve/reject, those drive execution)
# --------------------------------------------------------------------------- #


def test_reviews_list_typed_page(sync_client: Retab) -> None:
    page = sync_client.workflows.reviews.list(limit=3)
    assert len(page.data) <= 3
    assert page.list_metadata is not None
    for review in page.data:
        assert isinstance(review, Review)
        assert isinstance(review.id, str) and review.id


def test_reviews_get_by_discovered_id(sync_client: Retab) -> None:
    page = sync_client.workflows.reviews.list(limit=10)
    if not page.data:
        pytest.skip("no reviews on staging to get-by-id")
    target_id = page.data[0].id
    fetched = sync_client.workflows.reviews.get(target_id)
    assert isinstance(fetched, Review)
    assert fetched.id == target_id


def test_reviews_get_bogus_id_404(sync_client: Retab) -> None:
    with pytest.raises(NotFoundError) as excinfo:
        sync_client.workflows.reviews.get("rev_creditless_bogus_" + uuid.uuid4().hex)
    assert excinfo.value.status_code == 404


# --------------------------------------------------------------------------- #
# Auth + async
# --------------------------------------------------------------------------- #


def test_runs_list_bad_api_key_401(bad_key_client: Retab) -> None:
    with pytest.raises(AuthenticationError) as excinfo:
        bad_key_client.workflows.runs.list(limit=1)
    assert excinfo.value.status_code == 401


@pytest.mark.asyncio
async def test_async_runs_list_and_get(async_client: AsyncRetab) -> None:
    page = await async_client.workflows.runs.list(limit=3)
    assert isinstance(page.data, list)
    for run in page.data:
        assert isinstance(run, WorkflowRun)
    if page.data:
        fetched = await async_client.workflows.runs.get(page.data[0].id)
        assert fetched.id == page.data[0].id


@pytest.mark.asyncio
async def test_async_runs_get_bogus_id_404(async_client: AsyncRetab) -> None:
    with pytest.raises(NotFoundError) as excinfo:
        await async_client.workflows.runs.get("run_creditless_bogus_" + uuid.uuid4().hex)
    assert excinfo.value.status_code == 404
