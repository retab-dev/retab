"""Regression: ``WorkflowSteps.list(block_ids=...)`` pushes the filter to the
server and the cursor closure re-sends it on every subsequent page.

Before the fix, the SDK applied ``block_ids`` client-side AFTER
``request_page`` returned. The closure that fetched page 2 never carried
``block_ids`` in the wire request, so the server returned every row for
the run on subsequent pages. Pushing filtering server-side means each
page's ``_prepared_request`` call carries ``block_ids`` in ``params``.
"""

from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock

import pytest

from retab.resources.workflows.steps import AsyncWorkflowSteps, WorkflowSteps


def _step(*, step_id: str, block_id: str) -> dict[str, object]:
    return {
        "step_id": step_id,
        "run_id": "run_1",
        "block_id": block_id,
        "block_type": "extract",
        "block_label": "Extract block",
        "lifecycle": {"status": "completed"},
    }


def _envelope(*items: dict[str, object], after: str | None = None) -> dict[str, object]:
    return {"data": list(items), "list_metadata": {"before": None, "after": after}}


def test_steps_list_block_ids_sends_param_on_first_page() -> None:
    """`block_ids` is sent in `params` on the very first wire request."""
    client = MagicMock()
    client._prepared_request.return_value = _envelope(
        _step(step_id="s_a1", block_id="block_keep"),
    )

    page = WorkflowSteps(client=client).list("run_1", block_ids=["block_keep"])

    assert [step.step_id for step in page.data] == ["s_a1"]
    first_request = client._prepared_request.call_args_list[0].args[0]
    assert first_request.params == {"run_id": "run_1", "block_ids": ["block_keep"], "limit": 200}


def test_steps_list_block_ids_filter_survives_auto_paging() -> None:
    """Every page (including pages fetched by the closure) must carry block_ids.

    With server-side filtering the response data is already trusted; the
    test asserts on the wire requests (the bug was that page-2's request
    silently dropped the filter).
    """
    client = MagicMock()
    client._prepared_request.side_effect = [
        _envelope(_step(step_id="s_a1", block_id="block_keep"), after="cursor_p2"),
        _envelope(_step(step_id="s_a2", block_id="block_keep"), after=None),
    ]

    page = WorkflowSteps(client=client).list("run_1", block_ids=["block_keep"])

    collected = [step.step_id for step in page.auto_paging_iter()]
    assert collected == ["s_a1", "s_a2"]
    assert client._prepared_request.call_count == 2

    page1_request = client._prepared_request.call_args_list[0].args[0]
    page2_request = client._prepared_request.call_args_list[1].args[0]
    assert page1_request.params == {"run_id": "run_1", "block_ids": ["block_keep"], "limit": 200}
    # Page 2 must keep `block_ids` AND add the cursor.
    assert page2_request.params == {
        "run_id": "run_1",
        "block_ids": ["block_keep"],
        "limit": 200,
        "after": "cursor_p2",
    }


def test_steps_list_without_block_ids_omits_param() -> None:
    """No filter supplied → `block_ids` does not appear in the request params."""
    client = MagicMock()
    client._prepared_request.return_value = _envelope(
        _step(step_id="s_a1", block_id="block_x"),
        _step(step_id="s_b1", block_id="block_y"),
    )

    page = WorkflowSteps(client=client).list("run_1")

    assert [step.step_id for step in page.data] == ["s_a1", "s_b1"]
    request = client._prepared_request.call_args_list[0].args[0]
    assert request.params == {"run_id": "run_1", "limit": 200}


# ---------------------------------------------------------------------------
# Async parity
# ---------------------------------------------------------------------------


@pytest.mark.asyncio
async def test_async_steps_list_block_ids_filter_survives_auto_paging() -> None:
    client = MagicMock()
    client._prepared_request = AsyncMock(
        side_effect=[
            _envelope(_step(step_id="s_a1", block_id="block_keep"), after="cursor_p2"),
            _envelope(_step(step_id="s_a2", block_id="block_keep"), after=None),
        ]
    )

    page = await AsyncWorkflowSteps(client=client).list("run_1", block_ids=["block_keep"])

    collected = [step.step_id async for step in page.auto_paging_iter()]
    assert collected == ["s_a1", "s_a2"]
    assert client._prepared_request.call_count == 2

    page1_request = client._prepared_request.call_args_list[0].args[0]
    page2_request = client._prepared_request.call_args_list[1].args[0]
    assert page1_request.params == {"run_id": "run_1", "block_ids": ["block_keep"], "limit": 200}
    assert page2_request.params == {
        "run_id": "run_1",
        "block_ids": ["block_keep"],
        "limit": 200,
        "after": "cursor_p2",
    }
