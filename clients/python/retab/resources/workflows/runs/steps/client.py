from typing import List

from ....._resource import AsyncAPIResource, SyncAPIResource
from .....types.standards import PreparedRequest
from .....types.workflows import (
    StepOutputResponse,
    WorkflowRunStep,
)


class WorkflowStepsMixin:
    """Mixin providing shared prepare methods for workflow step operations."""

    def prepare_get(self, run_id: str, node_id: str) -> PreparedRequest:
        """Prepare a request to get the output of a specific step."""
        return PreparedRequest(method="GET", url=f"/v1/workflows/runs/{run_id}/steps/{node_id}")

    def prepare_list(self, run_id: str) -> PreparedRequest:
        """Prepare a request to list all step documents for a workflow run."""
        return PreparedRequest(method="GET", url=f"/v1/workflows/runs/{run_id}/steps")


class WorkflowSteps(SyncAPIResource, WorkflowStepsMixin):
    """Workflow Run Steps API wrapper for synchronous operations.

    Usage: ``client.workflows.runs.steps.get(run_id, node_id)``

    Example:
        >>> from retab import Retab
        >>> client = Retab(api_key="your-api-key")
        >>> step = client.workflows.runs.steps.get(run.id, "extract-node-1")
        >>> print(step.handle_outputs)
    """

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    def get(self, run_id: str, node_id: str) -> StepOutputResponse:
        """Get the full output data for a specific step in a workflow run.

        Args:
            run_id: The ID of the workflow run
            node_id: The ID of the node/step to get output for

        Returns:
            StepOutputResponse: The step's output data including handle_outputs and handle_inputs

        Example:
            >>> step = client.workflows.runs.steps.get(run.id, "extract-node-1")
            >>> if step.handle_outputs:
            ...     for handle_id, payload in step.handle_outputs.items():
            ...         print(f"{handle_id}: {payload}")
        """
        request = self.prepare_get(run_id, node_id)
        response = self._client._prepared_request(request)
        return StepOutputResponse.model_validate(response)

    def list(self, run_id: str) -> List[WorkflowRunStep]:
        """List all persisted step documents for a workflow run.

        Args:
            run_id: The ID of the workflow run

        Returns:
            List[WorkflowRunStep]: The stored step documents for the run
        """
        request = self.prepare_list(run_id)
        response = self._client._prepared_request(request)
        return [WorkflowRunStep.model_validate(item) for item in response]


class AsyncWorkflowSteps(AsyncAPIResource, WorkflowStepsMixin):
    """Workflow Run Steps API wrapper for asynchronous operations.

    Usage: ``await client.workflows.runs.steps.get(run_id, node_id)``

    Example:
        >>> from retab import AsyncRetab
        >>> client = AsyncRetab(api_key="your-api-key")
        >>> step = await client.workflows.runs.steps.get(run.id, "extract-node-1")
        >>> print(step.handle_outputs)
    """

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    async def get(self, run_id: str, node_id: str) -> StepOutputResponse:
        """Get the full output data for a specific step in a workflow run.

        Args:
            run_id: The ID of the workflow run
            node_id: The ID of the node/step to get output for

        Returns:
            StepOutputResponse: The step's output data including handle_outputs and handle_inputs

        Example:
            >>> step = await client.workflows.runs.steps.get(run.id, "extract-node-1")
            >>> if step.handle_outputs:
            ...     for handle_id, payload in step.handle_outputs.items():
            ...         print(f"{handle_id}: {payload}")
        """
        request = self.prepare_get(run_id, node_id)
        response = await self._client._prepared_request(request)
        return StepOutputResponse.model_validate(response)

    async def list(self, run_id: str) -> List[WorkflowRunStep]:
        """List all persisted step documents for a workflow run.

        Args:
            run_id: The ID of the workflow run

        Returns:
            List[WorkflowRunStep]: The stored step documents for the run
        """
        request = self.prepare_list(run_id)
        response = await self._client._prepared_request(request)
        return [WorkflowRunStep.model_validate(item) for item in response]
