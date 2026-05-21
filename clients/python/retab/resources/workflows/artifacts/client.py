from __future__ import annotations


from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.pagination import PaginatedList
from ....types.standards import PreparedRequest
from ....types.workflows import (
    WorkflowArtifact,
    WorkflowArtifactOperation,
)


class WorkflowArtifactsMixin:
    """Mixin providing shared prepare methods for workflow artifacts."""

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

    def prepare_get(self, artifact_id: str) -> PreparedRequest:
        """Prepare a request to fetch one workflow artifact by flat id."""
        return PreparedRequest(
            method="GET",
            url=f"/workflows/artifacts/{artifact_id}",
        )


class WorkflowArtifacts(SyncAPIResource, WorkflowArtifactsMixin):
    """Workflow artifact API wrapper for synchronous operations."""

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

    def get(self, artifact_id: str) -> WorkflowArtifact:
        """Fetch one dereferenced workflow artifact by flat artifact id."""
        request = self.prepare_get(artifact_id)
        response = self._client._prepared_request(request)
        return WorkflowArtifact.model_validate(response)


class AsyncWorkflowArtifacts(AsyncAPIResource, WorkflowArtifactsMixin):
    """Workflow artifact API wrapper for asynchronous operations."""

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

    async def get(self, artifact_id: str) -> WorkflowArtifact:
        """Fetch one dereferenced workflow artifact by flat artifact id."""
        request = self.prepare_get(artifact_id)
        response = await self._client._prepared_request(request)
        return WorkflowArtifact.model_validate(response)
