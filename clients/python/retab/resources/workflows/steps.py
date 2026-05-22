from __future__ import annotations

from typing import List, Optional

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.pagination import PaginatedList
from ...types.standards import PreparedRequest
from ...types.workflows import WorkflowRunStep


class WorkflowStepsMixin:
    """Mixin providing shared prepare methods for workflow step operations."""

    def prepare_get(self, step_id: str) -> PreparedRequest:
        """Prepare a request to get one joined step row."""
        return PreparedRequest(method="GET", url=f"/workflows/steps/{step_id}")

    def prepare_list(self, run_id: str) -> PreparedRequest:
        """Prepare a request to list all persisted step documents for a run."""
        return PreparedRequest(method="GET", url=f"/workflows/steps?run_id={run_id}")


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
            block_ids: If provided, filters the returned step documents to these block IDs.

        Returns:
            ``PaginatedList[WorkflowRunStep]`` — the canonical list envelope
            ``{"data": [...], "list_metadata": {"before": null, "after": null}}``.
            ``block_ids`` filtering is applied client-side after the wire
            response is parsed; ``list_metadata`` is preserved verbatim from
            the unfiltered response.
        """
        request = self.prepare_list(run_id)
        response = self._client._prepared_request(request)
        result = PaginatedList[WorkflowRunStep](**response)
        steps = [WorkflowRunStep.model_validate(item) for item in result.data]
        if block_ids is not None:
            requested_block_ids = set(block_ids)
            steps = [step for step in steps if step.block_id in requested_block_ids]
        result.data = steps
        return result


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
    ) -> PaginatedList[WorkflowRunStep]:
        """List step documents for a workflow run.

        Args:
            run_id: The ID of the workflow run
            block_ids: If provided, filters the returned step documents to these block IDs.

        Returns:
            ``PaginatedList[WorkflowRunStep]`` — the canonical list envelope.
        """
        request = self.prepare_list(run_id)
        response = await self._client._prepared_request(request)
        result = PaginatedList[WorkflowRunStep](**response)
        steps = [WorkflowRunStep.model_validate(item) for item in result.data]
        if block_ids is not None:
            requested_block_ids = set(block_ids)
            steps = [step for step in steps if step.block_id in requested_block_ids]
        result.data = steps
        return result
