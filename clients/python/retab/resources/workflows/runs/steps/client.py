from __future__ import annotations

from typing import TYPE_CHECKING, List, Optional

from ....._resource import AsyncAPIResource, SyncAPIResource
from .....types.standards import PreparedRequest
from .....types.workflows import (
    StepOutputResponse,
    StepOutputsBatchResponse,
    WorkflowRunStep,
)

if TYPE_CHECKING:
    from .....types.workflows import WorkflowRun


class WorkflowStepsMixin:
    """Mixin providing shared prepare methods for workflow step operations."""

    def prepare_get(self, run_id: str, block_id: str) -> PreparedRequest:
        """Prepare a request to get status and handles for a specific step."""
        return PreparedRequest(method="GET", url=f"/workflows/runs/{run_id}/steps/{block_id}")

    def prepare_list(self, run_id: str) -> PreparedRequest:
        """Prepare a request to list all persisted step documents for a run."""
        return PreparedRequest(method="GET", url=f"/workflows/runs/{run_id}/steps")

    def prepare_get_many(self, run_id: str, block_ids: List[str]) -> PreparedRequest:
        """Prepare a request to fetch normalized outputs for a subset of blocks."""
        return PreparedRequest(
            method="POST",
            url=f"/workflows/runs/{run_id}/steps/batch",
            data={"block_ids": block_ids},
        )


class WorkflowSteps(SyncAPIResource, WorkflowStepsMixin):
    """Workflow Run Steps API wrapper for synchronous operations.

    Usage: ``client.workflows.runs.steps.get(run_id, block_id)``
    """

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    def get(self, run_id: str, block_id: str) -> StepOutputResponse:
        """Get step status and handle data for a specific step in a workflow run.

        Args:
            run_id: The ID of the workflow run
            block_id: The ID of the node/step to get output for

        Returns:
            StepOutputResponse with handle_outputs and handle_inputs.
                Use ``.extracted_data`` for quick access to JSON output data.
        """
        request = self.prepare_get(run_id, block_id)
        response = self._client._prepared_request(request)
        return StepOutputResponse.model_validate(response)

    def list(self, run_id: str, block_ids: Optional[List[str]] = None) -> List[WorkflowRunStep]:
        """List step documents for a workflow run.

        Args:
            run_id: The ID of the workflow run
            block_ids: If provided, filters the returned step documents to these block IDs.

        Returns:
            List[WorkflowRunStep] for the requested steps.
        """
        request = self.prepare_list(run_id)
        response = self._client._prepared_request(request)
        steps = [WorkflowRunStep.model_validate(item) for item in response]
        if block_ids is None:
            return steps
        requested_block_ids = set(block_ids)
        return [step for step in steps if step.block_id in requested_block_ids]

    def get_many(self, run_id: str, block_ids: List[str]) -> StepOutputsBatchResponse:
        """Fetch normalized outputs for a subset of blocks in a workflow run."""
        request = self.prepare_get_many(run_id, block_ids)
        response = self._client._prepared_request(request)
        return StepOutputsBatchResponse.model_validate(response)

    def get_all(self, run: WorkflowRun | str) -> StepOutputsBatchResponse:
        """Fetch outputs for all steps in a workflow run in one call.

        Args:
            run: A ``WorkflowRun`` object or a run ID string.
                If a string is passed, the run is fetched first to discover step node IDs.

        Returns:
            StepOutputsBatchResponse with all step outputs keyed by node ID.
        """
        if isinstance(run, str):
            from .....types.workflows import WorkflowRun as WR
            req = PreparedRequest(method="GET", url=f"/workflows/runs/{run}")
            resp = self._client._prepared_request(req)
            run = WR.model_validate(resp)
        block_ids = [s.block_id for s in run.steps]
        if not block_ids:
            return StepOutputsBatchResponse(outputs={})
        return self.get_many(run.id, block_ids=block_ids)


class AsyncWorkflowSteps(AsyncAPIResource, WorkflowStepsMixin):
    """Workflow Run Steps API wrapper for asynchronous operations.

    Usage: ``await client.workflows.runs.steps.get(run_id, block_id)``
    """

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    async def get(self, run_id: str, block_id: str) -> StepOutputResponse:
        """Get step status and handle data for a specific step in a workflow run.

        Args:
            run_id: The ID of the workflow run
            block_id: The ID of the node/step to get output for

        Returns:
            StepOutputResponse with handle_outputs and handle_inputs.
                Use ``.extracted_data`` for quick access to JSON output data.
        """
        request = self.prepare_get(run_id, block_id)
        response = await self._client._prepared_request(request)
        return StepOutputResponse.model_validate(response)

    async def list(self, run_id: str, block_ids: Optional[List[str]] = None) -> List[WorkflowRunStep]:
        """List step documents for a workflow run.

        Args:
            run_id: The ID of the workflow run
            block_ids: If provided, filters the returned step documents to these block IDs.

        Returns:
            List[WorkflowRunStep] for the requested steps.
        """
        request = self.prepare_list(run_id)
        response = await self._client._prepared_request(request)
        steps = [WorkflowRunStep.model_validate(item) for item in response]
        if block_ids is None:
            return steps
        requested_block_ids = set(block_ids)
        return [step for step in steps if step.block_id in requested_block_ids]

    async def get_many(self, run_id: str, block_ids: List[str]) -> StepOutputsBatchResponse:
        """Fetch normalized outputs for a subset of blocks in a workflow run."""
        request = self.prepare_get_many(run_id, block_ids)
        response = await self._client._prepared_request(request)
        return StepOutputsBatchResponse.model_validate(response)

    async def get_all(self, run: WorkflowRun | str) -> StepOutputsBatchResponse:
        """Fetch outputs for all steps in a workflow run in one call.

        Args:
            run: A ``WorkflowRun`` object or a run ID string.

        Returns:
            StepOutputsBatchResponse with all step outputs keyed by node ID.
        """
        if isinstance(run, str):
            from .....types.workflows import WorkflowRun as WR
            req = PreparedRequest(method="GET", url=f"/workflows/runs/{run}")
            resp = await self._client._prepared_request(req)
            run = WR.model_validate(resp)
        block_ids = [s.block_id for s in run.steps]
        if not block_ids:
            return StepOutputsBatchResponse(outputs={})
        return await self.get_many(run.id, block_ids=block_ids)
