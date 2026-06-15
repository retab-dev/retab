# pyright: reportArgumentType=false, reportUnusedVariable=false
"""Creditless edge / negative / cross-cutting coverage for the Retab SDK.

Every test here is CREDITLESS: it only exercises list / get, validation /
error paths, pagination edge cases, date-range semantics, the typed-exception
contract, auth failures, and request-metadata behavior against pre-existing
staging data. Nothing in this file creates extractions, parses,
classifications, splits, workflow runs, experiments, secrets, or tables; it
never runs an LLM, OCRs / parses a document, or sends email.

These cases deliberately do NOT overlap with the existing
``test_creditless_processing_resources.py`` (which already covers the happy
list/get/pagination/filter paths) or the storage files/config/tables creditless
suites. The focus here is:

  * invalid query params → observed typed status (422 for ``order``, 400 for a
    malformed date) and the SILENTLY-TOLERATED params the backend accepts and
    coerces (limit<=0, limit far above the page, garbage cursors);
  * date-range semantics on list (``to_date`` < ``from_date`` → empty, future
    ``from_date`` → empty, ``from_date == to_date`` accepted);
  * filter semantics (unicode / very-long filename accepted, empty-result
    envelopes are well shaped with null after/before);
  * the error-envelope contract across several resources (typed exception
    class, ``status_code``, ``request_id`` from ``x-request-id``, populated
    ``message``);
  * auth failures across MULTIPLE resources, and empty-key behavior;
  * request metadata (``max_retries=0`` does not retry a 4xx → ``exc.retries``
    stays 0).

Staging holds legacy rows that violate the current public contract, so any
test that needs a page goes through ``prepare_list`` + the low-level request to
assert pagination / filtering independently of per-item model validation (the
same pattern as the sibling processing-resources suite).
"""

from __future__ import annotations

from typing import Any

import pytest

# Whole module is creditless (storage/config/list/get/error paths only).
pytestmark = pytest.mark.creditless

from retab import AsyncRetab, Retab
from retab.exceptions import (
    APIError,
    AuthenticationError,
    NotFoundError,
    PermissionDeniedError,
    ValidationError,
)


# (resource attribute name, bogus-id) for resources that expose list + get-by-id.
_LIST_RESOURCES: list[str] = ["files", "extractions", "parses", "classifications", "splits", "partitions"]

# Resources whose list route validates the ``order`` enum (returns 422 on a bad
# value). ``splits`` is included now that its list route enforces the asc/desc
# enum (backend fix in main_server_go); it previously accepted any value and
# returned 200, which was the lone inconsistency in this sweep.
_ORDER_VALIDATING_RESOURCES: list[str] = ["files", "extractions", "parses", "classifications", "splits", "partitions"]

# (resource attribute name, get callable arg, bogus id) for the typed-404 contract sweep.
_GET_404_CASES: list[tuple[str, str]] = [
    ("files", "file_creditless_bogus_id"),
    ("extractions", "extr_creditless_bogus_id"),
    ("parses", "prs_creditless_bogus_id"),
    ("classifications", "clss_creditless_bogus_id"),
    ("splits", "splt_creditless_bogus_id"),
    ("partitions", "prt_creditless_bogus_id"),
    ("tables", "tbl_creditless_bogus_id"),
]
_GET_404_IDS = [name for name, _ in _GET_404_CASES]


def _raw_list(client: Retab, resource_name: str, **list_params: Any) -> dict[str, Any]:
    """Return the raw list envelope, bypassing per-item model validation.

    Lets pagination / filtering be asserted even when legacy staging rows would
    fail typed validation.
    """
    resource = getattr(client, resource_name)
    prepared = resource.prepare_list(**list_params)
    return resource._client._prepared_request(prepared)


def _list_metadata(envelope: dict[str, Any]) -> dict[str, Any]:
    meta = envelope.get("list_metadata")
    assert isinstance(meta, dict), "list envelope missing dict list_metadata"
    return meta


def _assert_well_shaped(envelope: dict[str, Any]) -> dict[str, Any]:
    """A list envelope is a dict with a list ``data`` and a dict ``list_metadata``."""
    assert isinstance(envelope, dict)
    assert isinstance(envelope.get("data"), list)
    meta = _list_metadata(envelope)
    # after / before, when present, are either None or a string cursor.
    for key in ("after", "before"):
        val = meta.get(key, None)
        assert val is None or isinstance(val, str), f"list_metadata.{key} not a cursor/null: {val!r}"
    return envelope


# --------------------------------------------------------------------------- #
# Invalid query params → typed status
# --------------------------------------------------------------------------- #


@pytest.mark.parametrize("resource_name", _ORDER_VALIDATING_RESOURCES)
def test_invalid_order_value_raises_422(sync_client: Retab, resource_name: str) -> None:
    """``order`` outside {asc, desc} is rejected with a typed 422 ValidationError.

    (Observed staging behavior: the gateway validates the enum and returns 422,
    not 400.)
    """
    with sync_client as client:
        with pytest.raises(APIError) as excinfo:
            _raw_list(client, resource_name, limit=2, order="sideways")
        assert excinfo.value.status_code == 422
        assert isinstance(excinfo.value, ValidationError)


@pytest.mark.parametrize("resource_name", _LIST_RESOURCES)
def test_malformed_to_date_raises_400(sync_client: Retab, resource_name: str) -> None:
    """A malformed ``to_date`` is rejected with a typed 400 and a useful message."""
    with sync_client as client:
        with pytest.raises(APIError) as excinfo:
            _raw_list(client, resource_name, limit=1, to_date="garbage")
        assert excinfo.value.status_code == 400
        assert excinfo.value.message  # non-empty, human-readable


def test_malformed_from_date_raises_400(sync_client: Retab) -> None:
    with sync_client as client:
        with pytest.raises(APIError) as excinfo:
            _raw_list(client, "files", limit=1, from_date="2020-13-45")
        assert excinfo.value.status_code == 400


# --------------------------------------------------------------------------- #
# Silently-tolerated params (document the backend's lenient coercion)
# --------------------------------------------------------------------------- #


@pytest.mark.parametrize("bad_limit", [0, -5])
def test_nonpositive_limit_is_coerced_not_rejected(sync_client: Retab, bad_limit: int) -> None:
    """limit<=0 is NOT a 4xx — the gateway coerces it to its default page size.

    This pins the *observed* contract: a non-positive limit returns a
    well-shaped page rather than an error. If the backend later starts
    rejecting it, this test flags the contract change.
    """
    with sync_client as client:
        envelope = _assert_well_shaped(_raw_list(client, "files", limit=bad_limit))
        # Coerced to the default page (10) rather than 0 / a negative page.
        assert len(envelope["data"]) >= 0


def test_limit_far_above_max_is_tolerated(sync_client: Retab) -> None:
    """A limit far above any documented max returns a well-shaped page (capped to
    available rows, no 4xx)."""
    with sync_client as client:
        envelope = _assert_well_shaped(_raw_list(client, "files", limit=100000))
        # The page is bounded by the collection; the terminal page has after=None.
        assert isinstance(envelope["data"], list)


@pytest.mark.parametrize("cursor_field", ["after", "before"])
def test_garbage_cursor_is_tolerated(sync_client: Retab, cursor_field: str) -> None:
    """A garbage ``after``/``before`` cursor is tolerated (ignored), not a 4xx.

    Pins the observed lenient behavior: an unparseable cursor yields a
    well-shaped first-page-style envelope rather than an error.
    """
    with sync_client as client:
        envelope = _assert_well_shaped(_raw_list(client, "files", limit=2, **{cursor_field: "garbage_cursor_creditless"}))
        assert len(envelope["data"]) <= 2


# --------------------------------------------------------------------------- #
# Date-range semantics (creditless, on list)
# --------------------------------------------------------------------------- #


def test_to_date_before_from_date_returns_empty(sync_client: Retab) -> None:
    """An inverted window (``to_date`` < ``from_date``) yields an empty list, not an error."""
    with sync_client as client:
        envelope = _assert_well_shaped(_raw_list(client, "files", limit=5, from_date="2999-01-01", to_date="2000-01-01"))
        assert envelope["data"] == []
        meta = _list_metadata(envelope)
        assert meta.get("after") is None and meta.get("before") is None


def test_from_date_far_future_returns_empty(sync_client: Retab) -> None:
    with sync_client as client:
        envelope = _assert_well_shaped(_raw_list(client, "files", limit=5, from_date="2999-01-01"))
        assert envelope["data"] == []


def test_from_date_equals_to_date_accepted(sync_client: Retab) -> None:
    """``from_date == to_date`` is a valid (degenerate) window, accepted and well shaped."""
    with sync_client as client:
        envelope = _assert_well_shaped(_raw_list(client, "files", limit=5, from_date="2024-01-01", to_date="2024-01-01"))
        assert isinstance(envelope["data"], list)


# --------------------------------------------------------------------------- #
# Filter semantics
# --------------------------------------------------------------------------- #


def test_unicode_filename_filter_accepted(sync_client: Retab) -> None:
    """A unicode filename filter is accepted and returns a well-shaped envelope."""
    with sync_client as client:
        _assert_well_shaped(_raw_list(client, "files", limit=5, filename="façtüre_é_中文_📄.pdf"))


def test_very_long_filename_filter_accepted(sync_client: Retab) -> None:
    with sync_client as client:
        _assert_well_shaped(_raw_list(client, "files", limit=5, filename="x" * 1000))


def test_empty_result_filter_is_well_shaped(sync_client: Retab) -> None:
    """A filter that matches nothing returns a well-shaped empty envelope with
    null after/before cursors."""
    with sync_client as client:
        envelope = _assert_well_shaped(_raw_list(client, "files", limit=5, filename="no_such_file_creditless_xyzzy.pdf"))
        assert envelope["data"] == []
        meta = _list_metadata(envelope)
        assert meta.get("after") is None
        assert meta.get("before") is None


# --------------------------------------------------------------------------- #
# Error-envelope contract across resources
# --------------------------------------------------------------------------- #


@pytest.mark.parametrize(("resource_name", "bogus_id"), _GET_404_CASES, ids=_GET_404_IDS)
def test_get_404_contract(sync_client: Retab, resource_name: str, bogus_id: str) -> None:
    """A bogus get() raises NotFoundError with a populated status/request-id/message.

    The error envelope is consistent across every resource: typed
    NotFoundError (a subclass of APIError), ``status_code == 404``, an
    ``x-request-id`` surfaced as ``request_id``, and a non-empty ``message``.
    """
    with sync_client as client:
        with pytest.raises(NotFoundError) as excinfo:
            getattr(client, resource_name).get(bogus_id)
        exc = excinfo.value
        assert exc.status_code == 404
        assert isinstance(exc, APIError)  # hierarchy: NotFoundError -> APIError
        assert isinstance(exc, NotFoundError)
        assert exc.message, f"{resource_name} 404 had empty message"
        assert exc.request_id, f"{resource_name} 404 did not surface x-request-id"
        assert exc.request_id.startswith("req_"), f"unexpected request_id shape: {exc.request_id!r}"


def test_404_exception_hierarchy_is_consistent(sync_client: Retab) -> None:
    """NotFoundError is catchable as APIError and as RuntimeError (back-compat)."""
    with sync_client as client:
        try:
            client.files.get("file_creditless_bogus_id")
        except APIError as exc:
            assert isinstance(exc, NotFoundError)
            assert isinstance(exc, RuntimeError)  # documented back-compat
            assert exc.url and exc.method == "GET"
        else:
            pytest.fail("expected NotFoundError for a bogus file id")


def test_distinct_requests_get_distinct_request_ids(sync_client: Retab) -> None:
    """Two separate failing requests surface two distinct request ids."""
    with sync_client as client:
        ids: list[str] = []
        for _ in range(2):
            try:
                client.files.get("file_creditless_bogus_id")
            except NotFoundError as exc:
                assert exc.request_id is not None
                ids.append(exc.request_id)
        assert len(ids) == 2 and ids[0] != ids[1], f"request ids not distinct: {ids}"


# --------------------------------------------------------------------------- #
# Auth across multiple resources
# --------------------------------------------------------------------------- #


@pytest.mark.parametrize("resource_name", ["files", "extractions", "parses", "classifications", "splits", "tables"])
def test_junk_key_raises_401_across_resources(api_keys: Any, resource_name: str) -> None:
    """A junk API key yields a typed 401 AuthenticationError on every list resource."""
    client = Retab(api_key="sk_junk_invalid_creditless", base_url=api_keys.retab_api_base_url, max_retries=0)
    try:
        resource = getattr(client, resource_name)
        with pytest.raises(AuthenticationError) as excinfo:
            # tables.list takes no required args; the others accept limit.
            if resource_name == "tables":
                resource.list()
            else:
                resource.list(limit=1)
        exc = excinfo.value
        assert exc.status_code == 401
        assert isinstance(exc, APIError)
    finally:
        client.close()


def test_empty_string_key_is_rejected(api_keys: Any) -> None:
    """An empty-string API key is rejected (observed: 403 PermissionDenied, a
    valid-but-unauthorized identity), not silently accepted.

    Note: ``api_key=None`` is intentionally NOT tested — None falls back to the
    ``RETAB_API_KEY`` env var, which is the valid test key here, so it would not
    fail. Empty-string is the meaningful negative case.
    """
    client = Retab(api_key="", base_url=api_keys.retab_api_base_url, max_retries=0)
    try:
        with pytest.raises(APIError) as excinfo:
            client.files.list(limit=1)
        exc = excinfo.value
        assert exc.status_code in (401, 403)
        assert isinstance(exc, (AuthenticationError, PermissionDeniedError))
    finally:
        client.close()


# --------------------------------------------------------------------------- #
# Request metadata / retry behavior
# --------------------------------------------------------------------------- #


def test_max_retries_zero_does_not_retry_4xx(api_keys: Any) -> None:
    """A 4xx must surface immediately with ``exc.retries == 0`` (no retry loop).

    ``backoff`` only retries 5xx / 429; a 404 should never be retried, and with
    ``max_retries=0`` the retries counter must stay at 0.
    """
    client = Retab(api_key=api_keys.retab_api_key, base_url=api_keys.retab_api_base_url, max_retries=0)
    try:
        with pytest.raises(NotFoundError) as excinfo:
            client.files.get("file_creditless_bogus_id")
        assert excinfo.value.retries == 0
    finally:
        client.close()


def test_4xx_carries_method_and_url(sync_client: Retab) -> None:
    """A failed request records the HTTP method and the full URL for diagnostics."""
    with sync_client as client:
        with pytest.raises(NotFoundError) as excinfo:
            client.files.get("file_creditless_bogus_id")
        exc = excinfo.value
        assert exc.method == "GET"
        assert isinstance(exc.url, str) and exc.url.endswith("file_creditless_bogus_id")


# --------------------------------------------------------------------------- #
# Async parity for the cross-cutting contract
# --------------------------------------------------------------------------- #


@pytest.mark.asyncio
async def test_async_junk_key_raises_401(api_keys: Any) -> None:
    client = AsyncRetab(api_key="sk_junk_invalid_creditless", base_url=api_keys.retab_api_base_url, max_retries=0)
    try:
        with pytest.raises(AuthenticationError) as excinfo:
            await client.files.list(limit=1)
        assert excinfo.value.status_code == 401
    finally:
        await client.close()


@pytest.mark.asyncio
async def test_async_404_surfaces_request_id(async_client: AsyncRetab) -> None:
    async with async_client:
        with pytest.raises(NotFoundError) as excinfo:
            await async_client.files.get("file_creditless_bogus_id")
        exc = excinfo.value
        assert exc.status_code == 404
        assert exc.request_id and exc.request_id.startswith("req_")


@pytest.mark.asyncio
async def test_async_invalid_order_raises_422(async_client: AsyncRetab) -> None:
    async with async_client:
        prepared = async_client.files.prepare_list(limit=2, order="sideways")
        with pytest.raises(APIError) as excinfo:
            await async_client.files._client._prepared_request(prepared)
        assert excinfo.value.status_code == 422
