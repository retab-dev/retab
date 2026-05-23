from __future__ import annotations

from typing import Any, List, Optional

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.pagination import AsyncPaginatedList, PaginatedList
from ...types.standards import PreparedRequest
from ...types.workflows import WorkflowRunStep


class WorkflowStepsMixin:
    """Mixin providing shared prepare methods for workflow step operations."""

    def prepare_get(self, step_id: str) -> PreparedRequest:
        """Prepare a request to get one joined step row."""
        return PreparedRequest(method="GET", url=f"/v1/workflows/steps/{step_id}")

    def prepare_list(
        self,
        run_id: str,
        block_ids: Optional[List[str]] = None,
    ) -> PreparedRequest:
        """Prepare a request to list all persisted step documents for a run.

        ``block_ids`` is sent as a repeated query parameter
        (``?block_ids=a&block_ids=b``) — FastAPI's standard multi-value
        ``Query`` parsing on the server side accepts this shape directly.
        Filtering happens server-side, so the cursor closure that fetches
        page 2+ re-uses the same params and the filter stays applied.
        """
        params: dict[str, Any] = {"run_id": run_id}
        if block_ids:
            params["block_ids"] = list(block_ids)
        return PreparedRequest(
            method="GET",
            url="/v1/workflows/steps",
            params=params,
        )


class WorkflowSteps(SyncAPIResource, WorkflowStepsMixin):
    """Workflow Steps API wrapper for synchronous operations.

    Usage:
    - ``client.workflows.steps.get(step_id)`` for one persisted step document.
    - ``client.workflows.steps.list(run_id)`` for persisted step documents.
    """

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    def get(self, step_id: str) -> WorkflowRunStep:
        """Get one persisted step document by step ID.

        Args:
            step_id: The ID of the workflow step to fetch.

        Returns:
            ``WorkflowRunStep`` for the requested step.
        """
        if not isinstance(step_id, str) or not step_id:
            raise TypeError("step_id is required")
        request = self.prepare_get(step_id)
        response = self._client._prepared_request(request)
        return WorkflowRunStep.model_validate(response)

    def list(
        self,
        run_id: str,
        block_ids: Optional[List[str]] = None,
    ) -> PaginatedList[WorkflowRunStep]:
        """List step documents for a workflow run.

        Args:
            run_id: The ID of the workflow run
            block_ids: If provided, restricts the returned step documents
                to these block IDs. Filtering is applied server-side via
                a repeated ``block_ids`` query parameter, so
                ``auto_paging_iter()`` keeps the filter on every page —
                the cursor closure re-issues each request with the same
                ``params``.

        Returns:
            ``PaginatedList[WorkflowRunStep]`` — the canonical list envelope
            ``{"data": [...], "list_metadata": {"before": null, "after": null}}``.
        """
        request = self.prepare_list(run_id, block_ids=block_ids)
        return self.request_page(request, model=WorkflowRunStep)


class AsyncWorkflowSteps(AsyncAPIResource, WorkflowStepsMixin):
    """Workflow Steps API wrapper for asynchronous operations.

    Usage:
    - ``await client.workflows.steps.get(step_id)`` for one persisted step document.
    - ``await client.workflows.steps.list(run_id)`` for persisted step documents.
    """

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    async def get(self, step_id: str) -> WorkflowRunStep:
        """Get one persisted step document by step ID.

        Args:
            step_id: The ID of the workflow step to fetch.

        Returns:
            ``WorkflowRunStep`` for the requested step.
        """
        if not isinstance(step_id, str) or not step_id:
            raise TypeError("step_id is required")
        request = self.prepare_get(step_id)
        response = await self._client._prepared_request(request)
        return WorkflowRunStep.model_validate(response)

    async def list(
        self,
        run_id: str,
        block_ids: Optional[List[str]] = None,
    ) -> AsyncPaginatedList[WorkflowRunStep]:
        """List step documents for a workflow run.

        ``block_ids`` is forwarded server-side as a repeated query parameter
        so ``auto_paging_iter()`` preserves the filter across pages — the
        cursor closure re-issues each page's request with the same params.
        """
        request = self.prepare_list(run_id, block_ids=block_ids)
        return await self.request_page(request, model=WorkflowRunStep)
