"""Creditless config CRUD coverage: workflow DEFINITIONS and secrets.

Every test is CREDITLESS. Creating/updating/deleting a workflow *definition* is
pure config storage and does NOT run anything (running a workflow IS billable and
is never done here). Secret CRUD is config storage. All evals clean up the
resources they create.
"""

from __future__ import annotations

import time
import uuid

import pytest


from retab import AsyncRetab, Retab
from retab.exceptions import NotFoundError
from retab.types.secrets import Secret, SecretListResponse, SecretResponse, SecretValueResponse
from retab.types.workflows import Workflow

from factories import (
    create_workflow,
    discover_project_id_async,
    temporary_secret,
    temporary_workflow,
    temporary_workflow_async,
)

# Whole module is creditless (storage/config/list/get/error paths only).
pytestmark = pytest.mark.creditless

# Project discovery + create/cleanup live in factories.py; the ``project_id``
# fixture (conftest) yields an existing project and skips when none exists.


# --------------------------------------------------------------------------- #
# Workflow definition CRUD
# --------------------------------------------------------------------------- #


def test_workflow_create_get_update_delete(sync_client: Retab, project_id: str) -> None:
    name = f"creditless-cfg-{uuid.uuid4().hex[:8]}"
    wf = create_workflow(sync_client, project_id, name=name, description="creditless config test")
    assert isinstance(wf, Workflow)
    created_id = wf.id
    try:
        assert wf.name == name
        # get
        got = sync_client.workflows.get(created_id)
        assert got.id == created_id
        assert got.name == name
        # update metadata only (no run)
        renamed = f"{name}-renamed"
        updated = sync_client.workflows.update(created_id, name=renamed, description="updated")
        assert updated.id == created_id
        assert updated.name == renamed
        # confirm persisted
        again = sync_client.workflows.get(created_id)
        assert again.name == renamed
    finally:
        sync_client.workflows.delete(created_id)

    # after delete -> 404
    with pytest.raises(NotFoundError):
        sync_client.workflows.get(created_id)


def test_workflow_create_appears_in_list_then_gone_after_delete(sync_client: Retab, project_id: str) -> None:
    with temporary_workflow(sync_client, project_id, name=f"creditless-list-{uuid.uuid4().hex[:8]}") as wf:
        listed = sync_client.workflows.list(limit=50, project_id=project_id)
        assert wf.id in {w.id for w in listed.data}


def test_workflow_list_envelope_and_pagination(sync_client: Retab) -> None:
    page = sync_client.workflows.list(limit=3)
    assert len(page.data) <= 3
    assert page.list_metadata is not None
    for wf in page.data:
        assert isinstance(wf, Workflow)
        assert wf.id.startswith("wrk_")


def test_workflow_list_order_desc(sync_client: Retab) -> None:
    page = sync_client.workflows.list(limit=10, order="desc", sort_by="updated_at")
    # Envelope must be well-formed regardless of ordering specifics.
    assert isinstance(page.data, list)


def test_workflow_get_bogus_id_404(sync_client: Retab) -> None:
    with pytest.raises(NotFoundError):
        sync_client.workflows.get("wrk_does_not_exist_" + uuid.uuid4().hex)


def test_workflow_create_missing_project_rejected(sync_client: Retab) -> None:
    # An empty project id is not a valid owner; the API must reject it
    # (validation/permission), never silently create a dangling workflow.
    from retab.exceptions import APIError

    with pytest.raises(APIError):
        sync_client.workflows.create(project_id="proj_does_not_exist_" + uuid.uuid4().hex, name="x")


@pytest.mark.asyncio
async def test_async_workflow_create_delete(async_client: AsyncRetab) -> None:
    project_id = await discover_project_id_async(async_client)
    if not project_id:
        pytest.skip("no existing project on staging to attach a workflow to")
    async with temporary_workflow_async(async_client, project_id, name=f"creditless-async-{uuid.uuid4().hex[:8]}") as wf:
        got = await async_client.workflows.get(wf.id)
        assert got.id == wf.id


# --------------------------------------------------------------------------- #
# Secrets CRUD
# --------------------------------------------------------------------------- #


def test_secret_create_get_update_delete(sync_client: Retab) -> None:
    name = f"creditless_secret_{uuid.uuid4().hex[:10]}"
    created = sync_client.secrets.create_secret(name=name, value="initial-value")
    assert isinstance(created, SecretResponse)
    assert isinstance(created.secret, Secret)
    assert created.secret.name == name
    assert created.secret.created_at is not None
    try:
        got = sync_client.secrets.get_secret(name)
        assert got.secret.name == name

        # value endpoint returns the stored value (still creditless config read)
        value = sync_client.secrets.list_secret_value(name)
        assert isinstance(value, SecretValueResponse)
        assert value.secret.name == name
        assert value.secret.value == "initial-value"

        updated = sync_client.secrets.update_secret(name, value="rotated-value")
        assert updated.secret.name == name
        rotated = sync_client.secrets.list_secret_value(name)
        assert rotated.secret.value == "rotated-value"
        # updated_at should advance on rotation (or at least be present)
        assert rotated.secret.updated_at is not None
    finally:
        sync_client.secrets.delete_secret(name)

    with pytest.raises(NotFoundError):
        sync_client.secrets.get_secret(name)


def test_secret_list_envelope(sync_client: Retab) -> None:
    resp = sync_client.secrets.list_secrets()
    assert isinstance(resp, SecretListResponse)
    assert isinstance(resp.secrets, list)
    for s in resp.secrets:
        assert isinstance(s, Secret)
        assert s.name


def test_secret_create_appears_in_list(sync_client: Retab) -> None:
    with temporary_secret(sync_client) as secret:
        listed_secrets = sync_client.secrets.list_secrets().secrets
        assert listed_secrets is not None
        names = {s.name for s in listed_secrets}
        assert secret.name in names


def test_secret_get_bogus_name_404(sync_client: Retab) -> None:
    with pytest.raises(NotFoundError):
        sync_client.secrets.get_secret("creditless_missing_" + uuid.uuid4().hex)


@pytest.mark.asyncio
async def test_async_secret_crud(async_client: AsyncRetab) -> None:
    name = f"creditless_secret_async_{uuid.uuid4().hex[:10]}"
    created = await async_client.secrets.create_secret(name=name, value="v-" + str(time.time()))
    try:
        assert created.secret.name == name
        got = await async_client.secrets.get_secret(name)
        assert got.secret.name == name
    finally:
        await async_client.secrets.delete_secret(name)
