from __future__ import annotations

from typing import List

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.pagination import PaginatedList
from ....types.standards import PreparedRequest
from ....types.workflows import (
    StepArtifactRef,
    WorkflowArtifact,
    WorkflowArtifactOperation,
)


class WorkflowArtifactsMixin:
    """Mixin providing shared prepare methods for workflow artifacts."""

    def prepare_get(
        self,
        operation: WorkflowArtifactOperation | StepArtifactRef,
        artifact_id: str | None = None,
    ) -> PreparedRequest:
        """Prepare a request to dereference a workflow artifact ref."""
        if isinstance(operation, StepArtifactRef):
            artifact_ref = operation
            operation = artifact_ref.operation
            artifact_id = artifact_ref.id
        if not artifact_id:
            raise TypeError("artifact_id is required")
        return PreparedRequest(
            method="GET",
            url=f"/workflows/artifacts/{operation}/{artifact_id}",
        )

    def prepare_list(
        self,
        run_id: str,
        operation: WorkflowArtifactOperation | None = None,
        block_id: str | None = None,
    ) -> PreparedRequest:
        """Prepare a request to list dereferenced artifacts for a run."""
        params = {
            "run_id": run_id,
            "operation": operation,
            "block_id": block_id,
        }
        return PreparedRequest(
            method="GET",
            url="/workflows/artifacts",
            params={key: value for key, value in params.items() if value is not None},
        )


class WorkflowArtifacts(SyncAPIResource, WorkflowArtifactsMixin):
    """Workflow artifact API wrapper for synchronous operations."""

    def get(
        self,
        operation: WorkflowArtifactOperation | StepArtifactRef,
        artifact_id: str | None = None,
    ) -> WorkflowArtifact:
        """Dereference a workflow step artifact ref into its persisted record."""
        request = self.prepare_get(operation, artifact_id)
        response = self._client._prepared_request(request)
        return WorkflowArtifact.model_validate(response)

    def list(
        self,
        run_id: str,
        operation: WorkflowArtifactOperation | None = None,
        block_id: str | None = None,
    ) -> PaginatedList[WorkflowArtifact]:
        """List dereferenced artifacts produced by one workflow run.

        Returns the canonical
        ``{"data": [...], "list_metadata": {"before": null, "after": null}}``
        pagination envelope. ID pagination is not yet implemented.
        """
        request = self.prepare_list(run_id, operation=operation, block_id=block_id)
        response = self._client._prepared_request(request)
        result = PaginatedList[WorkflowArtifact](**response)
        result.data = [WorkflowArtifact.model_validate(item) for item in result.data]
        return result


class AsyncWorkflowArtifacts(AsyncAPIResource, WorkflowArtifactsMixin):
    """Workflow artifact API wrapper for asynchronous operations."""

    async def get(
        self,
        operation: WorkflowArtifactOperation | StepArtifactRef,
        artifact_id: str | None = None,
    ) -> WorkflowArtifact:
        """Dereference a workflow step artifact ref into its persisted record."""
        request = self.prepare_get(operation, artifact_id)
        response = await self._client._prepared_request(request)
        return WorkflowArtifact.model_validate(response)

    async def list(
        self,
        run_id: str,
        operation: WorkflowArtifactOperation | None = None,
        block_id: str | None = None,
    ) -> PaginatedList[WorkflowArtifact]:
        """List dereferenced artifacts produced by one workflow run.

        Returns the canonical
        ``{"data": [...], "list_metadata": {"before": null, "after": null}}``
        pagination envelope.
        """
        request = self.prepare_list(run_id, operation=operation, block_id=block_id)
        response = await self._client._prepared_request(request)
        result = PaginatedList[WorkflowArtifact](**response)
        result.data = [WorkflowArtifact.model_validate(item) for item in result.data]
        return result
