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

# Whole module is creditless (storage/config/list/get/error paths only).
pytestmark = pytest.mark.creditless

from retab import Retab
from retab.exceptions import APIError, AuthenticationError, NotFoundError, PermissionDeniedError
from retab.resources.tables import AsyncTables, Tables
from retab.types.tables import WorkflowTableListResponse, WorkflowTableResponse

from factories import TINY_CSV, temporary_table

# Project discovery, the tiny CSV, and table create/cleanup live in factories.py;
# the ``project_id`` fixture (conftest) supplies an existing project or skips.


# --------------------------------------------------------------------------- #
# Tables: required-param + error paths (creditless)
# --------------------------------------------------------------------------- #


def test_tables_list_requires_project_id(sync_client: Retab) -> None:
    # The route rejects a project-less list with 400 "Project ID is required".
    with pytest.raises(APIError) as exc_info:
        sync_client.tables.list()
    assert exc_info.value.status_code == 400


def test_tables_list_empty_response_sdk_contract_bug(sync_client: Retab, project_id: str) -> None:
    """KNOWN SDK BUG (reported): staging returns ``{}`` for an empty table list,
    but ``WorkflowTableListResponse.tables`` is a *required* field with no
    default, so ``tables.list(project_id=...)`` raises a pydantic ValidationError
    instead of an empty ``tables=[]`` envelope.

    This test pins the current broken behavior so a fix (default ``tables=[]`` in
    the model, or the API always emitting the key) flips it to xpass.
    """
    from pydantic import ValidationError as PydanticValidationError

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


def test_tables_create_get_delete_roundtrip(sync_client: Retab, project_id: str) -> None:
    name = f"creditless-table-{uuid.uuid4().hex[:8]}"
    with temporary_table(sync_client, project_id, name=name, file=TINY_CSV) as table:
        table_id = table.id
        assert table.name == name
        assert table.row_count == 2

        got = sync_client.tables.get(table_id)
        assert isinstance(got, WorkflowTableResponse)
        assert got.table.id == table_id
        assert got.table.row_count == 2

    # The context manager deleted the table on exit -> now gone.
    with pytest.raises(NotFoundError):
        sync_client.tables.get(table_id)


def test_tables_download_returns_bytes_without_text_decoding() -> None:
    class FakeClient:
        def __init__(self) -> None:
            self.seen_url: str | None = None

        def _prepared_request_bytes(self, request: object) -> bytes:
            self.seen_url = getattr(request, "url")
            return b"name,amount\nalice,10\n"

        def _prepared_request(self, request: object) -> str:
            raise AssertionError("download must use the raw bytes request path")

    fake = FakeClient()
    resource = Tables(client=fake)  # type: ignore[arg-type]

    downloaded = resource.download("tbl_test")

    assert downloaded == b"name,amount\nalice,10\n"
    assert fake.seen_url == "/v1/tables/tbl_test/download"


@pytest.mark.asyncio
async def test_async_tables_download_returns_bytes_without_text_decoding() -> None:
    class FakeClient:
        def __init__(self) -> None:
            self.seen_url: str | None = None

        async def _prepared_request_bytes(self, request: object) -> bytes:
            self.seen_url = getattr(request, "url")
            return b"name,amount\nalice,10\n"

        async def _prepared_request(self, request: object) -> str:
            raise AssertionError("download must use the raw bytes request path")

    fake = FakeClient()
    resource = AsyncTables(client=fake)  # type: ignore[arg-type]

    downloaded = await resource.download("tbl_test")

    assert downloaded == b"name,amount\nalice,10\n"
    assert fake.seen_url == "/v1/tables/tbl_test/download"


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
