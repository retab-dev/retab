from typing import Any, Dict, List, Optional

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.standards import PreparedRequest
from ....types.workflows import WorkflowEdge


class WorkflowEdgesMixin:
    """Mixin providing shared prepare methods for workflow edge operations."""

    def prepare_list(
        self,
        workflow_id: str,
        source_block: str | None = None,
        target_block: str | None = None,
    ) -> PreparedRequest:
        """Prepare a request to list all edges for a workflow."""
        params: Dict[str, Any] = {}
        if source_block is not None:
            params["source_block"] = source_block
        if target_block is not None:
            params["target_block"] = target_block
        return PreparedRequest(method="GET", url=f"/workflows/{workflow_id}/edges", params=params or None)

    def prepare_get(self, workflow_id: str, edge_id: str) -> PreparedRequest:
        """Prepare a request to get a single edge."""
        return PreparedRequest(method="GET", url=f"/workflows/{workflow_id}/edges/{edge_id}")

    def prepare_create(
        self,
        workflow_id: str,
        id: str,
        source_block: str,
        target_block: str,
        source_handle: str | None = None,
        target_handle: str | None = None,
    ) -> PreparedRequest:
        """Prepare a request to create a new edge."""
        data: Dict[str, Any] = {
            "id": id,
            "source_block": source_block,
            "target_block": target_block,
        }
        if source_handle is not None:
            data["source_handle"] = source_handle
        if target_handle is not None:
            data["target_handle"] = target_handle
        return PreparedRequest(method="POST", url=f"/workflows/{workflow_id}/edges", data=data)

    def prepare_create_batch(self, workflow_id: str, edges: List[Dict[str, Any]]) -> PreparedRequest:
        """Prepare a request to create multiple edges at once."""
        return PreparedRequest(method="POST", url=f"/workflows/{workflow_id}/edges/batch", data=edges)

    def prepare_delete(self, workflow_id: str, edge_id: str) -> PreparedRequest:
        """Prepare a request to delete an edge."""
        return PreparedRequest(method="DELETE", url=f"/workflows/{workflow_id}/edges/{edge_id}")

    def prepare_delete_all(self, workflow_id: str) -> PreparedRequest:
        """Prepare a request to delete all edges for a workflow."""
        return PreparedRequest(method="DELETE", url=f"/workflows/{workflow_id}/edges")


class WorkflowEdges(SyncAPIResource, WorkflowEdgesMixin):
    """Workflow Edges API wrapper for synchronous operations.

    Usage: ``client.workflows.edges.list(workflow_id)``
    """

    def list(
        self,
        workflow_id: str,

        source_block: str | None = None,
        target_block: str | None = None,
    ) -> List[WorkflowEdge]:
        """List all edges for a workflow.

        Args:
            workflow_id: The workflow ID
            source_block: Filter by source block ID (optional)
            target_block: Filter by target block ID (optional)

        Returns:
            List of workflow edges
        """
        request = self.prepare_list(workflow_id, source_block=source_block, target_block=target_block)
        response = self._client._prepared_request(request)
        return [WorkflowEdge.model_validate(item) for item in response]

    def get(self, workflow_id: str, edge_id: str) -> WorkflowEdge:
        """Get a single edge by ID."""
        request = self.prepare_get(workflow_id, edge_id)
        response = self._client._prepared_request(request)
        return WorkflowEdge.model_validate(response)

    def create(
        self,
        workflow_id: str,

        id: str,
        source_block: str,
        target_block: str,
        source_handle: str | None = None,
        target_handle: str | None = None,
    ) -> WorkflowEdge:
        """Create a new edge connecting two blocks.

        Args:
            workflow_id: The workflow ID
            id: Client-provided edge ID (e.g., "edge-1")
            source_block: Source block ID
            target_block: Target block ID
            source_handle: Output handle on source block (e.g., "output-file-0")
            target_handle: Input handle on target block (e.g., "input-file-0")

        Returns:
            WorkflowEdge: The created edge
        """
        request = self.prepare_create(
            workflow_id, id=id, source_block=source_block, target_block=target_block,
            source_handle=source_handle, target_handle=target_handle,
        )
        response = self._client._prepared_request(request)
        return WorkflowEdge.model_validate(response)

    def create_batch(self, workflow_id: str, edges: List[Dict[str, Any]]) -> List[WorkflowEdge]:
        """Create multiple edges in a single request.

        Args:
            workflow_id: The workflow ID
            edges: List of edge dicts, each with ``id``, ``source_block``, ``target_block``

        Returns:
            List of created edges
        """
        request = self.prepare_create_batch(workflow_id, edges)
        response = self._client._prepared_request(request)
        return [WorkflowEdge.model_validate(item) for item in response]

    def delete(self, workflow_id: str, edge_id: str) -> None:
        """Delete an edge."""
        request = self.prepare_delete(workflow_id, edge_id)
        self._client._prepared_request(request)

    def delete_all(self, workflow_id: str) -> None:
        """Delete all edges for a workflow."""
        request = self.prepare_delete_all(workflow_id)
        self._client._prepared_request(request)


class AsyncWorkflowEdges(AsyncAPIResource, WorkflowEdgesMixin):
    """Workflow Edges API wrapper for asynchronous operations.

    Usage: ``await client.workflows.edges.list(workflow_id)``
    """

    async def list(
        self,
        workflow_id: str,

        source_block: str | None = None,
        target_block: str | None = None,
    ) -> List[WorkflowEdge]:
        """List all edges for a workflow."""
        request = self.prepare_list(workflow_id, source_block=source_block, target_block=target_block)
        response = await self._client._prepared_request(request)
        return [WorkflowEdge.model_validate(item) for item in response]

    async def get(self, workflow_id: str, edge_id: str) -> WorkflowEdge:
        """Get a single edge by ID."""
        request = self.prepare_get(workflow_id, edge_id)
        response = await self._client._prepared_request(request)
        return WorkflowEdge.model_validate(response)

    async def create(
        self,
        workflow_id: str,

        id: str,
        source_block: str,
        target_block: str,
        source_handle: str | None = None,
        target_handle: str | None = None,
    ) -> WorkflowEdge:
        """Create a new edge connecting two blocks."""
        request = self.prepare_create(
            workflow_id, id=id, source_block=source_block, target_block=target_block,
            source_handle=source_handle, target_handle=target_handle,
        )
        response = await self._client._prepared_request(request)
        return WorkflowEdge.model_validate(response)

    async def create_batch(self, workflow_id: str, edges: List[Dict[str, Any]]) -> List[WorkflowEdge]:
        """Create multiple edges in a single request."""
        request = self.prepare_create_batch(workflow_id, edges)
        response = await self._client._prepared_request(request)
        return [WorkflowEdge.model_validate(item) for item in response]

    async def delete(self, workflow_id: str, edge_id: str) -> None:
        """Delete an edge."""
        request = self.prepare_delete(workflow_id, edge_id)
        await self._client._prepared_request(request)

    async def delete_all(self, workflow_id: str) -> None:
        """Delete all edges for a workflow."""
        request = self.prepare_delete_all(workflow_id)
        await self._client._prepared_request(request)
