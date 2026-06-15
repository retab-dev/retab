# pyright: reportUnusedVariable=false
"""Creditless coverage for the EDITS resource family: edits + edit templates.

Every test here is CREDITLESS. It exercises only list / get / pagination /
error paths against pre-existing staging data. Nothing here creates an edit
(``edits.create`` runs the form-fill pipeline and IS billable) or creates a
template (``templates.create`` uploads + preprocesses a document). We never run
an LLM, never process a document, and never consume credits.

Covered (not covered by the other ``test_creditless_*`` files):
  - ``client.edits``           — list / get / pagination / 404
  - ``client.edits.templates`` — list / get / pagination / 404
"""

from __future__ import annotations

import uuid

import pytest


from retab import AsyncRetab, Retab
from retab.exceptions import AuthenticationError, NotFoundError
from retab.types.edits import Edit
from retab.types.edits.templates import EditTemplate

# Whole module is creditless (storage/config/list/get/error paths only).
pytestmark = pytest.mark.creditless


# --------------------------------------------------------------------------- #
# Edits — list / get / pagination / errors
# --------------------------------------------------------------------------- #


def test_edits_list_typed_page_and_envelope(sync_client: Retab) -> None:
    page = sync_client.edits.list(limit=3)
    assert len(page.data) <= 3
    assert page.list_metadata is not None
    for edit in page.data:
        assert isinstance(edit, Edit)
        assert isinstance(edit.id, str) and edit.id


def test_edits_pagination_limit_respected(sync_client: Retab) -> None:
    page = sync_client.edits.list(limit=2)
    assert len(page.data) <= 2


def test_edits_pagination_after_cursor_disjoint(sync_client: Retab) -> None:
    page1 = sync_client.edits.list(limit=2, order="desc")
    after = page1.list_metadata.after
    if after is None or len(page1.data) < 2:
        pytest.skip("not enough edits on staging to page")
    page2 = sync_client.edits.list(limit=2, order="desc", after=after)
    ids1 = {e.id for e in page1.data}
    ids2 = {e.id for e in page2.data}
    assert ids1.isdisjoint(ids2), "edits cursor returned overlapping rows"


def test_edits_from_date_yyyy_mm_dd_accepted(sync_client: Retab) -> None:
    page = sync_client.edits.list(limit=3, from_date="2020-01-01")
    assert isinstance(page.data, list)


def test_edits_filename_filter_shape(sync_client: Retab) -> None:
    page = sync_client.edits.list(limit=5, filename="creditless_matches_nothing_xyz.pdf")
    assert isinstance(page.data, list)


def test_edits_get_by_discovered_id(sync_client: Retab) -> None:
    page = sync_client.edits.list(limit=10)
    if not page.data:
        pytest.skip("no edits on staging to get-by-id")
    target_id = page.data[0].id
    fetched = sync_client.edits.get(target_id)
    assert isinstance(fetched, Edit)
    assert fetched.id == target_id


def test_edits_get_bogus_id_404(sync_client: Retab) -> None:
    with pytest.raises(NotFoundError) as excinfo:
        sync_client.edits.get("edt_creditless_bogus_" + uuid.uuid4().hex)
    assert excinfo.value.status_code == 404


# --------------------------------------------------------------------------- #
# Edit templates — list / get / pagination / errors
# --------------------------------------------------------------------------- #


def test_templates_list_typed_page_and_envelope(sync_client: Retab) -> None:
    page = sync_client.edits.templates.list(limit=3)
    assert len(page.data) <= 3
    assert page.list_metadata is not None
    for tpl in page.data:
        assert isinstance(tpl, EditTemplate)
        assert isinstance(tpl.id, str) and tpl.id


def test_templates_pagination_limit_respected(sync_client: Retab) -> None:
    page = sync_client.edits.templates.list(limit=1)
    assert len(page.data) <= 1


def test_templates_get_by_discovered_id(sync_client: Retab) -> None:
    page = sync_client.edits.templates.list(limit=10)
    if not page.data:
        pytest.skip("no edit templates on staging to get-by-id")
    target_id = page.data[0].id
    fetched = sync_client.edits.templates.get(target_id)
    assert isinstance(fetched, EditTemplate)
    assert fetched.id == target_id


def test_templates_get_bogus_id_404(sync_client: Retab) -> None:
    with pytest.raises(NotFoundError) as excinfo:
        sync_client.edits.templates.get("etpl_creditless_bogus_" + uuid.uuid4().hex)
    assert excinfo.value.status_code == 404


# --------------------------------------------------------------------------- #
# Auth + async
# --------------------------------------------------------------------------- #


def test_edits_list_bad_api_key_401(api_keys) -> None:
    client = Retab(api_key="sk_junk_invalid_creditless", base_url=api_keys.retab_api_base_url, max_retries=0)
    try:
        with pytest.raises(AuthenticationError) as excinfo:
            client.edits.list(limit=1)
        assert excinfo.value.status_code == 401
    finally:
        client.close()


@pytest.mark.asyncio
async def test_async_edits_list_and_get(async_client: AsyncRetab) -> None:
    page = await async_client.edits.list(limit=3)
    assert isinstance(page.data, list)
    for edit in page.data:
        assert isinstance(edit, Edit)
    if page.data:
        fetched = await async_client.edits.get(page.data[0].id)
        assert fetched.id == page.data[0].id


@pytest.mark.asyncio
async def test_async_templates_list(async_client: AsyncRetab) -> None:
    page = await async_client.edits.templates.list(limit=3)
    assert isinstance(page.data, list)
    for tpl in page.data:
        assert isinstance(tpl, EditTemplate)
