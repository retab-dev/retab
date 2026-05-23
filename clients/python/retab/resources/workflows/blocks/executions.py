from typing import Any, Dict

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.pagination import AsyncPaginatedList, PaginatedList
from ....types.standards import PreparedRequest
from ....types.workflows.blocks.executions import StoredBlockExecution


class WorkflowBlockExecutionsMixin:
    """Shared prepare methods for workflow block execution operations."""

    def prepare_create(
        self,
        *,
        run_id: str,
        block_id: str,
        step_id: str | None = None,
        n_consensus: int | None = None,
    ) -> PreparedRequest:
        """Prepare a request to create and run a workflow block execution."""
        data: Dict[str, Any] = {
            "run_id": run_id,
            "block_id": block_id,
        }
        if step_id is not None:
            data["step_id"] = step_id
        if n_consensus is not None:
            data["n_consensus"] = n_consensus
        return PreparedRequest(method="POST", url="/v1/workflows/blocks/executions", data=data)

    def prepare_list(
        self,
        *,
        run_id: str,
        block_id: str,
        limit: int = 20,
    ) -> PreparedRequest:
        """Prepare a request to list block executions for a run/block pair."""
        return PreparedRequest(
            method="GET",
            url="/v1/workflows/blocks/executions",
            params={
                "run_id": run_id,
                "block_id": block_id,
                "limit": limit,
            },
        )


class WorkflowBlockExecutions(SyncAPIResource, WorkflowBlockExecutionsMixin):
    """Workflow Blocks Executions API wrapper for synchronous operations."""

    def create(
        self,
        *,
        run_id: str,
        block_id: str,
        step_id: str | None = None,
        n_consensus: int | None = None,
    ) -> StoredBlockExecution:
        """Create and run a workflow block execution."""
        request = self.prepare_create(
            run_id=run_id,
            block_id=block_id,
            step_id=step_id,
            n_consensus=n_consensus,
        )
        response = self._client._prepared_request(request)
        return StoredBlockExecution.model_validate(response)

    def list(
        self,
        *,
        run_id: str,
        block_id: str,
        limit: int = 20,
    ) -> PaginatedList[StoredBlockExecution]:
        """List block executions for a run/block pair."""
        request = self.prepare_list(run_id=run_id, block_id=block_id, limit=limit)
        return self.request_page(request, model=StoredBlockExecution)


class AsyncWorkflowBlockExecutions(AsyncAPIResource, WorkflowBlockExecutionsMixin):
    """Workflow Blocks Executions API wrapper for asynchronous operations."""

    async def create(
        self,
        *,
        run_id: str,
        block_id: str,
        step_id: str | None = None,
        n_consensus: int | None = None,
    ) -> StoredBlockExecution:
        """Create and run a workflow block execution."""
        request = self.prepare_create(
            run_id=run_id,
            block_id=block_id,
            step_id=step_id,
            n_consensus=n_consensus,
        )
        response = await self._client._prepared_request(request)
        return StoredBlockExecution.model_validate(response)

    async def list(
        self,
        *,
        run_id: str,
        block_id: str,
        limit: int = 20,
    ) -> AsyncPaginatedList[StoredBlockExecution]:
        """List block executions for a run/block pair."""
        request = self.prepare_list(run_id=run_id, block_id=block_id, limit=limit)
        return await self.request_page(request, model=StoredBlockExecution)
