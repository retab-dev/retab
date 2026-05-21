from typing import Any, Dict

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.pagination import PaginatedList
from ....types.standards import PreparedRequest
from ....types.workflows import BlockSimulation


class WorkflowSimulationsMixin:
    """Shared prepare methods for workflow simulation operations."""

    def prepare_create(
        self,
        *,
        run_id: str,
        block_id: str,
        step_id: str | None = None,
        n_consensus: int | None = None,
    ) -> PreparedRequest:
        """Prepare a request to create and run a workflow block simulation."""
        data: Dict[str, Any] = {
            "run_id": run_id,
            "block_id": block_id,
        }
        if step_id is not None:
            data["step_id"] = step_id
        if n_consensus is not None:
            data["n_consensus"] = n_consensus
        return PreparedRequest(method="POST", url="/workflows/simulations", data=data)

    def prepare_list(
        self,
        *,
        run_id: str,
        block_id: str,
        limit: int = 20,
    ) -> PreparedRequest:
        """Prepare a request to list simulations for a run/block pair."""
        return PreparedRequest(
            method="GET",
            url="/workflows/simulations",
            params={
                "run_id": run_id,
                "block_id": block_id,
                "limit": limit,
            },
        )


class WorkflowSimulations(SyncAPIResource, WorkflowSimulationsMixin):
    """Workflow Simulations API wrapper for synchronous operations."""

    def create(
        self,
        *,
        run_id: str,
        block_id: str,
        step_id: str | None = None,
        n_consensus: int | None = None,
    ) -> BlockSimulation:
        """Create and run a workflow block simulation."""
        request = self.prepare_create(
            run_id=run_id,
            block_id=block_id,
            step_id=step_id,
            n_consensus=n_consensus,
        )
        response = self._client._prepared_request(request)
        return BlockSimulation.model_validate(response)

    def list(
        self,
        *,
        run_id: str,
        block_id: str,
        limit: int = 20,
    ) -> PaginatedList[BlockSimulation]:
        """List simulations for a run/block pair."""
        request = self.prepare_list(run_id=run_id, block_id=block_id, limit=limit)
        response = self._client._prepared_request(request)
        result = PaginatedList[BlockSimulation](**response)
        result.data = [BlockSimulation.model_validate(item) for item in result.data]
        return result


class AsyncWorkflowSimulations(AsyncAPIResource, WorkflowSimulationsMixin):
    """Workflow Simulations API wrapper for asynchronous operations."""

    async def create(
        self,
        *,
        run_id: str,
        block_id: str,
        step_id: str | None = None,
        n_consensus: int | None = None,
    ) -> BlockSimulation:
        """Create and run a workflow block simulation."""
        request = self.prepare_create(
            run_id=run_id,
            block_id=block_id,
            step_id=step_id,
            n_consensus=n_consensus,
        )
        response = await self._client._prepared_request(request)
        return BlockSimulation.model_validate(response)

    async def list(
        self,
        *,
        run_id: str,
        block_id: str,
        limit: int = 20,
    ) -> PaginatedList[BlockSimulation]:
        """List simulations for a run/block pair."""
        request = self.prepare_list(run_id=run_id, block_id=block_id, limit=limit)
        response = await self._client._prepared_request(request)
        result = PaginatedList[BlockSimulation](**response)
        result.data = [BlockSimulation.model_validate(item) for item in result.data]
        return result
