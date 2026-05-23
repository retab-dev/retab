"""Regression tests for the default `sort_by` value on workflow run listing.

The backend rejects `sort_by="created_at"` for `/v1/workflows/runs` with HTTP
400 — it only accepts the dotted fields `"timing.created_at"` and
`"timing.started_at"`. The SDK previously defaulted to the bare
`"created_at"`, which made `client.workflows.runs.list()` (no args) always
fail against the real server.

These tests pin the default at the wire-correct value so the bug cannot
silently regress.
"""

from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock

import pytest

from retab.resources.workflows import AsyncWorkflows, Workflows


_EMPTY_PAGE = {"data": [], "list_metadata": {"before": None, "after": None}}


def _sync_client() -> MagicMock:
    client = MagicMock()
    client._prepared_request.return_value = _EMPTY_PAGE
    return client


def _async_client() -> MagicMock:
    client = MagicMock()
    client._prepared_request = AsyncMock(return_value=_EMPTY_PAGE)
    return client


def test_runs_list_default_sort_by_uses_dotted_timing_path() -> None:
    """`client.workflows.runs.list()` with no `sort_by` MUST send
    `sort_by=timing.created_at`. The backend rejects the bare
    `created_at` with HTTP 400, so this default is load-bearing.
    """
    client = _sync_client()

    Workflows(client=client).runs.list()

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/v1/workflows/runs"
    assert request.params["sort_by"] == "timing.created_at"


def test_runs_list_caller_supplied_sort_by_is_passed_through() -> None:
    """An explicit `sort_by` from the caller must still reach the wire
    unmodified — the default-fix must not clobber explicit overrides.
    """
    client = _sync_client()

    Workflows(client=client).runs.list(sort_by="timing.started_at")

    request = client._prepared_request.call_args.args[0]
    assert request.params["sort_by"] == "timing.started_at"


@pytest.mark.asyncio
async def test_async_runs_list_default_sort_by_uses_dotted_timing_path() -> None:
    """Async parity for the default-sort regression."""
    client = _async_client()

    await AsyncWorkflows(client=client).runs.list()

    request = client._prepared_request.call_args.args[0]
    assert request.method == "GET"
    assert request.url == "/v1/workflows/runs"
    assert request.params["sort_by"] == "timing.created_at"
