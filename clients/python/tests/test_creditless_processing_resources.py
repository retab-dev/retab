# pyright: reportArgumentType=false, reportCallIssue=false, reportUnusedVariable=false
"""Creditless coverage for the processing list/get resources.

Every test here is CREDITLESS: it only exercises list / get / pagination /
filtering / error paths against pre-existing staging data. Nothing in this file
creates extractions, parses, classifications, or splits, runs an LLM, or
otherwise consumes credits.

Staging holds legacy rows that violate the current public contract (e.g.
``Classification.categories == null`` and ``Parse.table_parsing_format == ""``).
The model validators correctly reject those rows, so any test that needs a typed
page validates items one at a time and skips malformed legacy rows. Raw-shape
tests go through ``prepare_list`` + the low-level request so pagination and
filtering can be asserted independently of model validation.
"""

from __future__ import annotations

from typing import Any

import pytest

from pydantic import ValidationError as PydanticValidationError

from retab import AsyncRetab, Retab
from retab.exceptions import APIError, AuthenticationError, NotFoundError
from retab.types.classifications import Classification
from retab.types.extractions import Extraction
from retab.types.parses import Parse
from retab.types.splits import Split
from factories import raw_ids, raw_list, validate_tolerant

# Whole module is creditless (storage/config/list/get/error paths only).
pytestmark = pytest.mark.creditless


# (resource attribute name, model class, bogus-id prefix) for every processing list resource.
_RESOURCES: list[tuple[str, type, str]] = [
    ("extractions", Extraction, "extr_creditless_bogus_id"),
    ("parses", Parse, "prs_creditless_bogus_id"),
    ("classifications", Classification, "clss_creditless_bogus_id"),
    ("splits", Split, "splt_creditless_bogus_id"),
]

_RESOURCE_IDS = [name for name, _model, _bogus in _RESOURCES]


# --------------------------------------------------------------------------- #
# List + typed-shape assertions
# --------------------------------------------------------------------------- #


@pytest.mark.parametrize(("resource_name", "model", "_bogus"), _RESOURCES, ids=_RESOURCE_IDS)
def test_list_default_returns_typed_page(sync_client: Retab, resource_name: str, model: type, _bogus: str) -> None:
    with sync_client as client:
        envelope = raw_list(client, resource_name, limit=5)
        assert isinstance(envelope, dict)
        assert "data" in envelope
        assert isinstance(envelope["data"], list)
        assert "list_metadata" in envelope

        validated, _skipped = validate_tolerant(envelope, model)
        for item in validated:
            assert isinstance(item, model)
            assert isinstance(item.id, str) and item.id


@pytest.mark.parametrize(("resource_name", "model", "_bogus"), _RESOURCES, ids=_RESOURCE_IDS)
def test_list_required_fields_present(sync_client: Retab, resource_name: str, model: type, _bogus: str) -> None:
    """Every validated row carries the contract's required identity fields."""
    with sync_client as client:
        validated, _skipped = validate_tolerant(raw_list(client, resource_name, limit=5), model)
        if not validated:
            pytest.skip(f"no validatable {resource_name} rows on staging")
        for item in validated:
            assert getattr(item, "id", None), f"{resource_name} row missing id"
            assert getattr(item, "model", None) is not None or hasattr(item, "output"), f"{resource_name} row shape unexpected"


# --------------------------------------------------------------------------- #
# Pagination
# --------------------------------------------------------------------------- #


@pytest.mark.parametrize(("resource_name", "_model", "_bogus"), _RESOURCES, ids=_RESOURCE_IDS)
def test_pagination_limit_respected(sync_client: Retab, resource_name: str, _model: type, _bogus: str) -> None:
    with sync_client as client:
        envelope = raw_list(client, resource_name, limit=2)
        assert len(envelope.get("data") or []) <= 2


@pytest.mark.parametrize(("resource_name", "_model", "_bogus"), _RESOURCES, ids=_RESOURCE_IDS)
def test_pagination_after_cursor_round_trip(sync_client: Retab, resource_name: str, _model: type, _bogus: str) -> None:
    with sync_client as client:
        page1 = raw_list(client, resource_name, limit=2, order="desc")
        after = (page1.get("list_metadata") or {}).get("after")
        ids1 = raw_ids(page1)
        if after is None or len(ids1) < 2:
            pytest.skip(f"not enough {resource_name} rows to page")
        page2 = raw_list(client, resource_name, limit=2, order="desc", after=after)
        ids2 = raw_ids(page2)
        assert set(ids1).isdisjoint(set(ids2)), f"{resource_name} cursor returned overlapping rows"


@pytest.mark.parametrize(("resource_name", "_model", "_bogus"), _RESOURCES, ids=_RESOURCE_IDS)
def test_pagination_order_asc_desc_differ(sync_client: Retab, resource_name: str, _model: type, _bogus: str) -> None:
    with sync_client as client:
        try:
            desc_ids = raw_ids(raw_list(client, resource_name, limit=5, order="desc"))
            asc_ids = raw_ids(raw_list(client, resource_name, limit=5, order="asc"))
        except APIError as exc:
            # Some legacy staging rows can 5xx on a full-window ordered scan; that is a
            # backend/data issue, not an SDK contract issue. Skip rather than create data.
            pytest.skip(f"{resource_name} ordered list unavailable on staging: {exc.status_code}")
        if len(desc_ids) < 2 or len(asc_ids) < 2:
            pytest.skip(f"not enough {resource_name} rows to compare order")
        # asc and desc over the same collection should not start identically.
        assert asc_ids[0] != desc_ids[0] or asc_ids[:2] != desc_ids[:2]


@pytest.mark.parametrize(("resource_name", "_model", "_bogus"), _RESOURCES, ids=_RESOURCE_IDS)
def test_before_cursor_round_trip(sync_client: Retab, resource_name: str, _model: type, _bogus: str) -> None:
    with sync_client as client:
        page1 = raw_list(client, resource_name, limit=2, order="desc")
        after = (page1.get("list_metadata") or {}).get("after")
        if after is None:
            pytest.skip(f"not enough {resource_name} rows to page with before/after")
        page2 = raw_list(client, resource_name, limit=2, order="desc", after=after)
        before = (page2.get("list_metadata") or {}).get("before")
        if before is None:
            pytest.skip(f"{resource_name} did not return a before cursor")
        back = raw_list(client, resource_name, limit=2, order="desc", before=before)
        # Paging back should return rows from the first page's neighborhood, not page2's rows.
        assert set(raw_ids(back)).isdisjoint(set(raw_ids(page2)))


# --------------------------------------------------------------------------- #
# Filtering
# --------------------------------------------------------------------------- #


@pytest.mark.parametrize(("resource_name", "_model", "_bogus"), _RESOURCES, ids=_RESOURCE_IDS)
def test_from_date_yyyy_mm_dd_accepted(sync_client: Retab, resource_name: str, _model: type, _bogus: str) -> None:
    """A well-formed YYYY-MM-DD from_date is accepted and returns a typed envelope."""
    with sync_client as client:
        envelope = raw_list(client, resource_name, limit=3, from_date="2020-01-01")
        assert isinstance(envelope.get("data"), list)


@pytest.mark.parametrize(("resource_name", "_model", "_bogus"), _RESOURCES, ids=_RESOURCE_IDS)
def test_to_date_yyyy_mm_dd_accepted(sync_client: Retab, resource_name: str, _model: type, _bogus: str) -> None:
    with sync_client as client:
        envelope = raw_list(client, resource_name, limit=3, to_date="2999-01-01")
        assert isinstance(envelope.get("data"), list)


@pytest.mark.parametrize(("resource_name", "_model", "_bogus"), _RESOURCES, ids=_RESOURCE_IDS)
def test_filename_filter_shape(sync_client: Retab, resource_name: str, _model: type, _bogus: str) -> None:
    """A filename filter returns a well-shaped (possibly empty) envelope."""
    with sync_client as client:
        envelope = raw_list(
            client,
            resource_name,
            limit=5,
            filename="this_filename_should_match_nothing_creditless.pdf",
        )
        assert isinstance(envelope.get("data"), list)


def test_extractions_status_filter_shape(sync_client: Retab) -> None:
    """Extractions list accepts a status filter and returns matching-or-empty typed rows."""
    with sync_client as client:
        envelope = raw_list(client, "extractions", limit=5, status="completed")
        assert isinstance(envelope.get("data"), list)
        validated, _skipped = validate_tolerant(envelope, Extraction)
        for item in validated:
            assert item.status is None or str(item.status) == "completed" or item.status == "completed"


# --------------------------------------------------------------------------- #
# Get-by-id discovered via list
# --------------------------------------------------------------------------- #


@pytest.mark.parametrize(("resource_name", "model", "_bogus"), _RESOURCES, ids=_RESOURCE_IDS)
def test_get_by_discovered_id(sync_client: Retab, resource_name: str, model: type, _bogus: str) -> None:
    with sync_client as client:
        # Discover an id whose row also validates, so the typed get() round-trips cleanly.
        envelope = raw_list(client, resource_name, limit=20)
        validated, _skipped = validate_tolerant(envelope, model)
        if not validated:
            pytest.skip(f"no validatable {resource_name} rows to get-by-id")
        target_id = validated[0].id
        try:
            fetched = getattr(client, resource_name).get(target_id)
        except APIError as exc:
            pytest.skip(f"{resource_name} get({target_id}) unavailable on staging: {exc.status_code}")
        assert fetched.id == target_id
        assert isinstance(fetched, model)


# --------------------------------------------------------------------------- #
# Error paths
# --------------------------------------------------------------------------- #


@pytest.mark.parametrize(("resource_name", "_model", "bogus"), _RESOURCES, ids=_RESOURCE_IDS)
def test_get_bogus_id_raises_not_found(sync_client: Retab, resource_name: str, _model: type, bogus: str) -> None:
    with sync_client as client:
        with pytest.raises(NotFoundError) as excinfo:
            getattr(client, resource_name).get(bogus)
        assert excinfo.value.status_code == 404


@pytest.mark.parametrize(("resource_name", "_model", "_bogus"), _RESOURCES, ids=_RESOURCE_IDS)
def test_bad_from_date_format_raises_400(sync_client: Retab, resource_name: str, _model: type, _bogus: str) -> None:
    with sync_client as client:
        with pytest.raises(APIError) as excinfo:
            raw_list(client, resource_name, limit=1, from_date="not-a-date")
        assert excinfo.value.status_code == 400


def test_bad_api_key_raises_401(api_keys) -> None:
    """A junk API key yields a typed 401 AuthenticationError on a creditless list."""
    client = Retab(api_key="sk_junk_invalid_creditless", base_url=api_keys.retab_api_base_url, max_retries=0)
    try:
        with pytest.raises(AuthenticationError) as excinfo:
            client.extractions.list(limit=1)
        assert excinfo.value.status_code == 401
    finally:
        client.close()


# --------------------------------------------------------------------------- #
# Async + auto-pagination
# --------------------------------------------------------------------------- #


@pytest.mark.asyncio
@pytest.mark.parametrize(("resource_name", "model", "_bogus"), _RESOURCES, ids=_RESOURCE_IDS)
async def test_async_list_typed_page(async_client: AsyncRetab, resource_name: str, model: type, _bogus: str) -> None:
    async with async_client:
        resource = getattr(async_client, resource_name)
        prepared = resource.prepare_list(limit=3)
        envelope = await resource._client._prepared_request(prepared)
        assert isinstance(envelope.get("data"), list)
        validated, _skipped = validate_tolerant(envelope, model)
        for item in validated:
            assert isinstance(item, model)
            assert item.id


@pytest.mark.asyncio
async def test_async_get_bogus_id_raises_not_found(async_client: AsyncRetab) -> None:
    async with async_client:
        with pytest.raises(NotFoundError) as excinfo:
            await async_client.parses.get("prs_creditless_bogus_id")
        assert excinfo.value.status_code == 404


def test_auto_paging_iter_stops(sync_client: Retab) -> None:
    """auto_paging_iter walks pages and terminates (bounded).

    auto_paging_iter lazily re-fetches and re-validates the next page, so a
    legacy row that violates the typed contract (e.g. ``table_parsing_format``
    == "") surfaces mid-iteration. That is a backend/data issue, not an
    iterator bug: catch it, confirm we got at least the first page's rows, and
    cap the walk so a large collection never becomes an unbounded scan.
    """
    with sync_client as client:
        try:
            page = client.parses.list(limit=2, order="desc")
        except PydanticValidationError:
            pytest.skip("legacy parse rows in newest window fail typed validation")
        first_page_count = len(page.data)
        seen = 0
        try:
            for _item in page.auto_paging_iter():
                seen += 1
                if seen >= 6:  # bound the walk; just prove it crosses page boundaries
                    break
        except PydanticValidationError:
            # A legacy row in a later page failed typed validation while paging.
            pass
        assert seen >= first_page_count >= 1
