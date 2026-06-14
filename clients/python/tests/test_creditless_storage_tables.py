"""Creditless coverage for the Tables resource + cross-cutting error paths.

Creating a table uploads a CSV (pure storage, creditless) but it also requires a
project context and leaves a row we would need to clean up; we focus on the
creditless read/error paths (list requires project_id, validation, 404, auth) and
exercise the full create->get->delete round-trip only when a project is available,
always cleaning up.
"""

from __future__ import annotations

import uuid

import pytest

from retab import Retab
from retab.exceptions import APIError, AuthenticationError, NotFoundError, PermissionDeniedError
from retab.types.tables import WorkflowTableListResponse, WorkflowTableResponse


def _discover_project_id(client: Retab) -> str | None:
    page = client.workflows.list(limit=25)
    for wf in page.data:
        pid = getattr(wf, "project_id", None)
        if pid:
            return pid
    return None


_TINY_CSV = b"name,amount\nalice,10\nbob,20\n"


# --------------------------------------------------------------------------- #
# Tables: required-param + error paths (creditless)
# --------------------------------------------------------------------------- #


def test_tables_list_requires_project_id(sync_client: Retab) -> None:
    # The route rejects a project-less list with 400 "Project ID is required".
    with pytest.raises(APIError) as exc_info:
        sync_client.tables.list()
    assert exc_info.value.status_code == 400


def test_tables_list_empty_response_sdk_contract_bug(sync_client: Retab) -> None:
    """KNOWN SDK BUG (reported): staging returns ``{}`` for an empty table list,
    but ``WorkflowTableListResponse.tables`` is a *required* field with no
    default, so ``tables.list(project_id=...)`` raises a pydantic ValidationError
    instead of an empty ``tables=[]`` envelope.

    This test pins the current broken behavior so a fix (default ``tables=[]`` in
    the model, or the API always emitting the key) flips it to xpass.
    """
    from pydantic import ValidationError as PydanticValidationError

    project_id = _discover_project_id(sync_client)
    if not project_id:
        pytest.skip("no existing project on staging")

    try:
        resp = sync_client.tables.list(project_id=project_id)
    except PydanticValidationError:
        pytest.xfail("SDK bug: WorkflowTableListResponse.tables required but API omits it for empty lists")
    else:
        # If the project happened to have tables, the envelope parses fine.
        assert isinstance(resp, WorkflowTableListResponse)
        assert isinstance(resp.tables, list)


def test_tables_get_bogus_id_404(sync_client: Retab) -> None:
    with pytest.raises(NotFoundError):
        sync_client.tables.get("tbl_does_not_exist_" + uuid.uuid4().hex)


def test_tables_create_get_delete_roundtrip(sync_client: Retab) -> None:
    project_id = _discover_project_id(sync_client)
    if not project_id:
        pytest.skip("no existing project on staging")

    name = f"creditless-table-{uuid.uuid4().hex[:8]}"
    created = sync_client.tables.create(name=name, file=_TINY_CSV, project_id=project_id)
    assert isinstance(created, WorkflowTableListResponse)
    assert created.tables, "create should return at least the new table"
    table = created.tables[0]
    table_id = table.id
    try:
        assert table.name == name
        assert table.row_count == 2

        got = sync_client.tables.get(table_id)
        assert isinstance(got, WorkflowTableResponse)
        assert got.table.id == table_id
        assert got.table.row_count == 2
    finally:
        sync_client.tables.delete(table_id)

    with pytest.raises(NotFoundError):
        sync_client.tables.get(table_id)


# --------------------------------------------------------------------------- #
# Cross-cutting: auth, malformed params, response contracts
# --------------------------------------------------------------------------- #


def test_junk_api_key_rejected_across_resources() -> None:
    bad = Retab(api_key="sk_definitely_not_valid", base_url="https://staging-api.retab.com", max_retries=0)
    try:
        with pytest.raises((AuthenticationError, PermissionDeniedError)):
            bad.workflows.list(limit=1)
        with pytest.raises((AuthenticationError, PermissionDeniedError)):
            bad.secrets.list_secrets()
    finally:
        bad.close()


def test_files_list_malformed_order_param(sync_client: Retab) -> None:
    # An invalid enum value for ``order`` must be rejected, not silently ignored.
    with pytest.raises(APIError):
        sync_client.files.list(limit=5, order="sideways")  # type: ignore[arg-type]


def test_workflows_list_negative_limit_rejected(sync_client: Retab) -> None:
    with pytest.raises(APIError):
        sync_client.workflows.list(limit=-5)


def test_pagination_metadata_contract(sync_client: Retab) -> None:
    page = sync_client.files.list(limit=2)
    # list_metadata always carries before/after (possibly None) — has_more is
    # derived from ``after``.
    assert hasattr(page.list_metadata, "before")
    assert hasattr(page.list_metadata, "after")
    assert page.has_more == (page.list_metadata.after is not None)
