from typing import Any, Dict, List, Sequence

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.standards import PreparedRequest
from ....types.workflows import WorkflowEdgeDoc, WorkflowEdgeCreateRequest


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
        request: WorkflowEdgeCreateRequest,
    ) -> PreparedRequest:
        """Prepare a request to create a new edge."""
        data = request.model_dump(exclude_none=True)
        return PreparedRequest(method="POST", url=f"/workflows/{workflow_id}/edges", data=data)

    def prepare_create_batch(
        self,
        workflow_id: str,
        edges: Sequence[WorkflowEdgeCreateRequest],
    ) -> PreparedRequest:
        """Prepare a request to create multiple edges at once."""
        return PreparedRequest(
            method="POST",
            url=f"/workflows/{workflow_id}/edges/batch",
            data=[edge.model_dump(exclude_none=True) for edge in edges],
        )

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

    @staticmethod
    def _coerce_create_request(
        request: WorkflowEdgeCreateRequest | None,
        id: str | None,
        source_block: str | None,
        target_block: str | None,
        source_handle: str | None,
        target_handle: str | None,
    ) -> WorkflowEdgeCreateRequest:
        if request is not None:
            return request
        if id is None or source_block is None or target_block is None:
            raise TypeError("id, source_block, and target_block are required when request is not provided")
        return WorkflowEdgeCreateRequest(
            id=id,
            source_block=source_block,
            target_block=target_block,
            source_handle=source_handle,
            target_handle=target_handle,
        )

    @staticmethod
    def _coerce_batch_requests(
        edges: Sequence[WorkflowEdgeCreateRequest | Dict[str, Any]],
    ) -> List[WorkflowEdgeCreateRequest]:
        return [
            edge if isinstance(edge, WorkflowEdgeCreateRequest) else WorkflowEdgeCreateRequest.model_validate(edge)
            for edge in edges
        ]

    def list(
        self,
        workflow_id: str,

        source_block: str | None = None,
        target_block: str | None = None,
    ) -> List[WorkflowEdgeDoc]:
        """List all edges for a workflow.

        Args:
            workflow_id: The workflow ID
            source_block: Filter by source block ID (optional)
            target_block: Filter by target block ID (optional)

        Returns:
            List of workflow edge documents
        """
        request = self.prepare_list(workflow_id, source_block=source_block, target_block=target_block)
        response = self._client._prepared_request(request)
        return [WorkflowEdgeDoc.model_validate(item) for item in response]

    def get(self, workflow_id: str, edge_id: str) -> WorkflowEdgeDoc:
        """Get a single edge by ID."""
        request = self.prepare_get(workflow_id, edge_id)
        response = self._client._prepared_request(request)
        return WorkflowEdgeDoc.model_validate(response)

    def create(
        self,
        workflow_id: str,
        id: str | None = None,
        source_block: str | None = None,
        target_block: str | None = None,
        source_handle: str | None = None,
        target_handle: str | None = None,
        request: WorkflowEdgeCreateRequest | None = None,
    ) -> WorkflowEdgeDoc:
        """Create a new edge connecting two blocks.

        Args:
            workflow_id: The workflow ID
            id: Client-provided edge ID (e.g., "edge-1")
            source_block: Source block ID
            target_block: Target block ID
            source_handle: Output handle on source block (e.g., "output-file-0")
            target_handle: Input handle on target block (e.g., "input-file-0")
            request: Optional typed request model. When provided, its values are used directly.

        Returns:
            WorkflowEdgeDoc: The created edge document
        """
        create_request = self._coerce_create_request(
            request=request,
            id=id,
            source_block=source_block,
            target_block=target_block,
            source_handle=source_handle,
            target_handle=target_handle,
        )
        prepared_request = self.prepare_create(workflow_id, create_request)
        response = self._client._prepared_request(prepared_request)
        return WorkflowEdgeDoc.model_validate(response)

    def create_batch(
        self,
        workflow_id: str,
        edges: Sequence[WorkflowEdgeCreateRequest | Dict[str, Any]],
    ) -> List[WorkflowEdgeDoc]:
        """Create multiple edges in a single request.

        Args:
            workflow_id: The workflow ID
            edges: Typed edge request models or dicts with ``id``, ``source_block``, and ``target_block``

        Returns:
            List of created edges
        """
        batch_requests = self._coerce_batch_requests(edges)
        prepared_request = self.prepare_create_batch(workflow_id, batch_requests)
        response = self._client._prepared_request(prepared_request)
        return [WorkflowEdgeDoc.model_validate(item) for item in response]

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
    ) -> List[WorkflowEdgeDoc]:
        """List all edges for a workflow."""
        request = self.prepare_list(workflow_id, source_block=source_block, target_block=target_block)
        response = await self._client._prepared_request(request)
        return [WorkflowEdgeDoc.model_validate(item) for item in response]

    async def get(self, workflow_id: str, edge_id: str) -> WorkflowEdgeDoc:
        """Get a single edge by ID."""
        request = self.prepare_get(workflow_id, edge_id)
        response = await self._client._prepared_request(request)
        return WorkflowEdgeDoc.model_validate(response)

    async def create(
        self,
        workflow_id: str,
        id: str | None = None,
        source_block: str | None = None,
        target_block: str | None = None,
        source_handle: str | None = None,
        target_handle: str | None = None,
        request: WorkflowEdgeCreateRequest | None = None,
    ) -> WorkflowEdgeDoc:
        """Create a new edge connecting two blocks."""
        create_request = self._coerce_create_request(
            request=request,
            id=id,
            source_block=source_block,
            target_block=target_block,
            source_handle=source_handle,
            target_handle=target_handle,
        )
        prepared_request = self.prepare_create(workflow_id, create_request)
        response = await self._client._prepared_request(prepared_request)
        return WorkflowEdgeDoc.model_validate(response)

    async def create_batch(
        self,
        workflow_id: str,
        edges: Sequence[WorkflowEdgeCreateRequest | Dict[str, Any]],
    ) -> List[WorkflowEdgeDoc]:
        """Create multiple edges in a single request."""
        batch_requests = self._coerce_batch_requests(edges)
        prepared_request = self.prepare_create_batch(workflow_id, batch_requests)
        response = await self._client._prepared_request(prepared_request)
        return [WorkflowEdgeDoc.model_validate(item) for item in response]

    async def delete(self, workflow_id: str, edge_id: str) -> None:
        """Delete an edge."""
        request = self.prepare_delete(workflow_id, edge_id)
        await self._client._prepared_request(request)

    async def delete_all(self, workflow_id: str) -> None:
        """Delete all edges for a workflow."""
        request = self.prepare_delete_all(workflow_id)
        await self._client._prepared_request(request)

    _coerce_create_request = WorkflowEdges._coerce_create_request
    _coerce_batch_requests = WorkflowEdges._coerce_batch_requests
