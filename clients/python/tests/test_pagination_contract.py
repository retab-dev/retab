"""Runtime regression test for the SDK pagination closure contract.

Per [.notes/blueprints/sdk-pagination-contract.md](../../../../../.notes/blueprints/sdk-pagination-contract.md),
every paginated ``.list()`` method on a public resource must delegate to the
central ``SyncAPIResource.request_page`` / ``AsyncAPIResource.request_page``
helper. The helper is what wires the ``_fetch_next_page`` closure so callers
get transparent auto-pagination via ``for item in page:`` and
``async for item in page:``.

If a future "cleanup" rewrites a ``.list()`` to call ``client._prepared_request``
directly and constructs ``PaginatedList(...)`` by hand, the wire response still
looks correct on the first page — but ``auto_paging_iter()`` silently stops
after the first page because ``_fetch_next_page`` is ``None``. This is the
exact silent failure mode the central helper exists to prevent, and the bug
WILL NOT surface in any of the per-resource list tests (which only check the
first page).

This test catches that drift at CI time by walking every public resource via
``inspect``, calling its ``.list()`` against a mock that returns an envelope
with ``list_metadata.after = "cursor-2"`` (i.e. "the server says there's more"),
and asserting the returned page has its closure wired.

Why the resource enumeration is hardcoded
-----------------------------------------
The discovery walk could in principle introspect ``Retab.__init__`` to find
every attribute that's a ``SyncAPIResource`` instance. But three sub-resources
live on resource objects, not directly on the client (``workflows.runs``,
``workflows.experiments.results``, etc.), so the walk would have to recurse
into resource attributes anyway. Hardcoding the top-level + sub-resource
paths keeps the discovery deterministic and makes drift impossible to miss
when someone adds a new resource: they have to update this list, which is
exactly the prompt to think about the closure contract.

The contract is "every ``PaginatedList[T]``-returning ``.list()`` method goes
through ``request_page``". Methods that return a different envelope type
*not* paginated lists in the contract's sense — they don't claim to support
``auto_paging_iter()``. We filter those out by inspecting the return annotation.
If a future resource needs to bypass ``request_page`` for a genuinely paginated
route, add it to ``KNOWN_BYPASS`` with a comment pointing at the "Acceptable
exceptions" section of the contract doc.
"""

from __future__ import annotations

import asyncio
import inspect
import typing
from typing import Any, Callable
from unittest.mock import AsyncMock, MagicMock

import pytest

from retab import AsyncRetab, Retab
from retab._resource import AsyncAPIResource, SyncAPIResource
from retab.types.pagination import AsyncPaginatedList, PaginatedList


# ---------------------------------------------------------------------------
# Allowlist for legitimate bypasses
# ---------------------------------------------------------------------------
# Per the "Acceptable exceptions" section of the pagination contract blueprint,
# nothing in the Python SDK currently legitimately bypasses ``request_page``
# for a ``PaginatedList[T]``-shaped response. (Go has WorkflowArtifacts/Blocks
# with dual-shape decoders; those are Go-specific.) Keep this set empty; if
# a new exception is unavoidable, add the full dotted path (e.g.
# ``"workflows.foo.list"``) with a one-line comment explaining why the
# central helper can't be used.
KNOWN_BYPASS: set[str] = set()


# ---------------------------------------------------------------------------
# Resource enumeration — hardcoded so adding a new resource forces a test edit
# ---------------------------------------------------------------------------
# Each entry is a dotted attribute path off the root client (Retab /
# AsyncRetab). The walk follows the dots and resolves the ``.list`` method
# at the leaf. New resources MUST be appended here (alongside their
# ``request_page`` rewire) so the contract test covers them.
RESOURCE_PATHS: tuple[str, ...] = (
    # Top-level
    "extractions",
    "parses",
    "partitions",
    "splits",
    "classifications",
    "edits",
    "files",
    "jobs",
    "workflows",
    # Sub-resources on `edits`
    "edits.templates",
    # Sub-resources on `workflows`
    "workflows.runs",
    "workflows.steps",
    "workflows.reviews",
    "workflows.artifacts",
    "workflows.blocks",
    "workflows.edges",
    "workflows.tests",
    "workflows.experiments",
    # Two-deep sub-resources
    "workflows.blocks.executions",
    "workflows.tests.runs",
    "workflows.tests.results",
    "workflows.experiments.runs",
    "workflows.experiments.results",
)


# ---------------------------------------------------------------------------
# Discovery helpers
# ---------------------------------------------------------------------------


def _resolve(root: Any, dotted_path: str) -> Any:
    """Walk a dotted attribute path on ``root`` and return the leaf object."""
    obj = root
    for part in dotted_path.split("."):
        obj = getattr(obj, part)
    return obj


def _returns_paginated_list(method: Callable[..., Any]) -> bool:
    """True iff the method's return annotation is ``PaginatedList[T]`` or
    ``AsyncPaginatedList[T]``.

    The closure invariant only applies to paginated envelopes — list methods
    that return a different envelope type don't claim to support
    ``auto_paging_iter()`` and are out of scope.

    Pydantic generics like ``PaginatedList[Extraction]`` are real classes
    (subclasses of ``PaginatedList``) rather than ``typing.GenericAlias``
    instances, so ``typing.get_origin`` returns ``None``. We use
    ``issubclass`` on the resolved annotation instead — it covers both the
    parameterised form ``PaginatedList[Foo]`` and the bare ``PaginatedList``.
    """
    try:
        hints = typing.get_type_hints(method)
    except Exception:
        # Methods built up from forward refs we can't resolve at runtime
        # default to ``False`` — they'll be silently skipped. The
        # alternative (fallthrough True) risks false positives.
        return False
    ret = hints.get("return")
    if ret is None or not isinstance(ret, type):
        return False
    return issubclass(ret, (PaginatedList, AsyncPaginatedList))


SENTINEL = "sentinel_for_contract_test"


def _build_kwargs_for_required_params(method: Callable[..., Any]) -> dict[str, Any]:
    """Return a kwargs dict supplying ``SENTINEL`` for every required param.

    Most ``.list()`` methods accept all-optional kwargs, so ``method()`` works.
    A few (e.g. ``workflow_steps.list(run_id)``,
    ``experiment_results.list(run_id)``) take a required positional that
    becomes part of the URL path or query. The mock doesn't care what the
    value is — it returns the canned envelope no matter what the request
    body looks like — so a string sentinel is harmless and keeps the
    discovery walk hands-off.
    """
    sig = inspect.signature(method)
    kwargs: dict[str, Any] = {}
    for name, param in sig.parameters.items():
        if name == "self":
            continue
        if param.default is not inspect.Parameter.empty:
            continue
        if param.kind in (
            inspect.Parameter.VAR_POSITIONAL,
            inspect.Parameter.VAR_KEYWORD,
        ):
            continue
        # POSITIONAL_OR_KEYWORD or KEYWORD_ONLY with no default → must supply
        kwargs[name] = SENTINEL
    return kwargs


def _envelope() -> dict[str, Any]:
    """Canonical wire envelope with ``after`` set so the closure must wire up.

    A page coming back with ``after = None`` (i.e. "no more pages") is allowed
    to skip the closure check — ``auto_paging_iter`` stops naturally. By
    forcing ``after = "cursor-2"`` we make the closure presence the only thing
    that determines whether a caller could continue paging.
    """
    return {"data": [], "list_metadata": {"before": None, "after": "cursor-2"}}


def _discover_methods(client: Any) -> list[tuple[str, Callable[..., Any]]]:
    """Walk ``RESOURCE_PATHS`` and collect ``(full_path, bound_method)`` for
    every ``.list`` method whose return annotation is a paginated envelope.

    Methods listed in ``KNOWN_BYPASS`` are skipped explicitly.
    """
    discovered: list[tuple[str, Callable[..., Any]]] = []
    for resource_path in RESOURCE_PATHS:
        resource = _resolve(client, resource_path)
        # Sanity: every resource we list must be an APIResource. If a name in
        # ``RESOURCE_PATHS`` resolves to something else (typo, refactor) the
        # whole test should fail loudly rather than silently skip.
        assert isinstance(resource, (SyncAPIResource, AsyncAPIResource)), (
            f"RESOURCE_PATHS entry {resource_path!r} resolved to {type(resource).__name__}, not a SyncAPIResource/AsyncAPIResource. Update the enumeration."
        )
        list_method = getattr(resource, "list", None)
        if list_method is None or not callable(list_method):
            continue
        full_path = f"{resource_path}.list"
        if full_path in KNOWN_BYPASS:
            continue
        if not _returns_paginated_list(list_method):
            continue
        discovered.append((full_path, list_method))
    return discovered


# Resolve the discovery once at module import — the parametrize IDs come from
# this list, so every contract violation surfaces as a named test failure.
_SYNC_CLIENT_FOR_DISCOVERY = Retab(api_key="contract-test", base_url="http://localhost")
_ASYNC_CLIENT_FOR_DISCOVERY = AsyncRetab(api_key="contract-test", base_url="http://localhost")

_SYNC_METHODS = _discover_methods(_SYNC_CLIENT_FOR_DISCOVERY)
_ASYNC_METHODS = _discover_methods(_ASYNC_CLIENT_FOR_DISCOVERY)
_SYNC_CLIENT_FOR_DISCOVERY.close()
asyncio.run(_ASYNC_CLIENT_FOR_DISCOVERY.close())


# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------


def test_discovery_found_methods() -> None:
    """If the introspection ever returns nothing, the parametrized cases below
    would silently pass with zero rows. Pin a sanity floor: there's at least
    one paginated list on every top-level resource that exposes one.
    """
    assert len(_SYNC_METHODS) >= 10, f"Sync discovery returned only {len(_SYNC_METHODS)} method(s); the resource enumeration or return-annotation filter is likely broken."
    assert len(_ASYNC_METHODS) >= 10, f"Async discovery returned only {len(_ASYNC_METHODS)} method(s); the resource enumeration or return-annotation filter is likely broken."


def test_discovery_clients_are_closed_after_param_generation() -> None:
    assert _SYNC_CLIENT_FOR_DISCOVERY.client.is_closed
    assert _ASYNC_CLIENT_FOR_DISCOVERY.client.is_closed


@pytest.mark.parametrize(
    "full_path, list_method",
    _SYNC_METHODS,
    ids=[name for name, _ in _SYNC_METHODS],
)
def test_sync_list_method_wires_closure(
    full_path: str,
    list_method: Callable[..., Any],
) -> None:
    """Every sync ``.list()`` must return a page with ``_fetch_next_page`` set
    when ``list_metadata.after`` is non-null. That's the runtime signature of
    going through ``request_page`` — a hand-rolled ``PaginatedList(...)``
    construction wouldn't set the private attribute.
    """
    # Swap the client's wire primitive in-place. ``request_page`` (running on
    # the resource itself) calls ``self._client._prepared_request(request)``;
    # mocking that one method is enough to short-circuit the whole HTTP path.
    resource = list_method.__self__  # type: ignore[attr-defined]
    resource._client._prepared_request = MagicMock(return_value=_envelope())

    kwargs = _build_kwargs_for_required_params(list_method)
    page = list_method(**kwargs)

    assert isinstance(page, PaginatedList), f"{full_path} returned {type(page).__name__}, expected PaginatedList. Check the return annotation and the helper delegation."
    assert page._fetch_next_page is not None, (
        f"{full_path} did not wire `_fetch_next_page` — it bypassed "
        "SyncAPIResource.request_page. Auto-pagination via `for item in page:` "
        "will silently stop after the first page. Rewrite this list method to "
        "delegate via `self.request_page(request, model=...)`. See "
        ".notes/blueprints/sdk-pagination-contract.md for the contract."
    )


@pytest.mark.parametrize(
    "full_path, list_method",
    _ASYNC_METHODS,
    ids=[name for name, _ in _ASYNC_METHODS],
)
def test_async_list_method_wires_closure(
    full_path: str,
    list_method: Callable[..., Any],
) -> None:
    """Async sibling of the sync test. The async ``request_page`` is the only
    code path that wires ``_fetch_next_page`` on the async page; bypassing it
    breaks ``async for item in page:`` the same way.
    """
    resource = list_method.__self__  # type: ignore[attr-defined]
    resource._client._prepared_request = AsyncMock(return_value=_envelope())

    kwargs = _build_kwargs_for_required_params(list_method)
    page = asyncio.run(list_method(**kwargs))

    assert isinstance(page, AsyncPaginatedList), f"{full_path} returned {type(page).__name__}, expected AsyncPaginatedList."
    assert page._fetch_next_page is not None, (
        f"{full_path} did not wire `_fetch_next_page` — it bypassed "
        "AsyncAPIResource.request_page. `async for item in page:` will silently "
        "stop after the first page. Rewrite this list method to delegate via "
        "`await self.request_page(request, model=...)`. See "
        ".notes/blueprints/sdk-pagination-contract.md for the contract."
    )
