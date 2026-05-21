from __future__ import annotations

from typing import List, Optional

from ....._resource import AsyncAPIResource, SyncAPIResource
from .....types.pagination import PaginatedList
from .....types.standards import PreparedRequest
from .....types.workflows import (
    StepExecutionResponse,
    WorkflowRunStep,
)


class WorkflowStepsMixin:
    """Mixin providing shared prepare methods for workflow step operations."""

    def prepare_get(self, run_id: str, block_id: str) -> PreparedRequest:
        """Prepare a request to get lifecycle and handles for a specific step."""
        return PreparedRequest(method="GET", url=f"/workflows/steps/{block_id}?run_id={run_id}")

    def prepare_list(self, run_id: str) -> PreparedRequest:
        """Prepare a request to list all persisted step documents for a run."""
        return PreparedRequest(method="GET", url=f"/workflows/steps?run_id={run_id}")


class WorkflowSteps(SyncAPIResource, WorkflowStepsMixin):
    """Workflow Run Steps API wrapper for synchronous operations.

    Usage:
    - ``client.workflows.runs.steps.get(run_id, block_id)`` for one step.
    - ``client.workflows.runs.steps.list(run_id)`` for step summaries.
    """

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    def get(self, run_id: str, block_id: str) -> StepExecutionResponse:
        """Get the full execution record for one step.

        Args:
            run_id: The ID of the workflow run.
            block_id: The ID of the block/step to fetch.

        Returns:
            ``StepExecutionResponse`` for the requested block.
        """
        if not isinstance(block_id, str) or not block_id:
            raise TypeError("block_id is required")
        request = self.prepare_get(run_id, block_id)
        response = self._client._prepared_request(request)
        return StepExecutionResponse.model_validate(response)

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
    """Workflow Run Steps API wrapper for asynchronous operations.

    Usage:
    - ``await client.workflows.runs.steps.get(run_id, block_id)`` for one step.
    - ``await client.workflows.runs.steps.list(run_id)`` for step summaries.
    """

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    async def get(self, run_id: str, block_id: str) -> StepExecutionResponse:
        """Get the full execution record for one step.

        Args:
            run_id: The ID of the workflow run.
            block_id: The ID of the block/step to fetch.

        Returns:
            ``StepExecutionResponse`` for the requested block.
        """
        if not isinstance(block_id, str) or not block_id:
            raise TypeError("block_id is required")
        request = self.prepare_get(run_id, block_id)
        response = await self._client._prepared_request(request)
        return StepExecutionResponse.model_validate(response)

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
