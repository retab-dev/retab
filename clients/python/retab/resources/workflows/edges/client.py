from typing import Any, Dict

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.pagination import PaginatedList
from ....types.standards import PreparedRequest
from ....types.workflows import WorkflowEdgeDoc, WorkflowEdgeCreateRequest


class WorkflowEdgesMixin:
    """Mixin providing shared prepare methods for workflow edge operations."""

    def prepare_list(
        self,
        workflow_id: str | None = None,
        source_block: str | None = None,
        target_block: str | None = None,
        before: str | None = None,
        after: str | None = None,
        limit: int | None = None,
    ) -> PreparedRequest:
        """Prepare a request to list all edges for a workflow."""
        params: Dict[str, Any] = {}
        if workflow_id is not None:
            params["workflow_id"] = workflow_id
        if source_block is not None:
            params["source_block"] = source_block
        if target_block is not None:
            params["target_block"] = target_block
        if before is not None:
            params["before"] = before
        if after is not None:
            params["after"] = after
        if limit is not None:
            params["limit"] = limit
        return PreparedRequest(method="GET", url="/workflows/edges", params=params or None)

    def prepare_get(self, edge_id: str) -> PreparedRequest:
        """Prepare a request to get a single edge."""
        return PreparedRequest(method="GET", url=f"/workflows/edges/{edge_id}")

    def prepare_create(
        self,
        workflow_id: str,
        request: WorkflowEdgeCreateRequest | None = None,
        source_block: str | None = None,
        target_block: str | None = None,
        id: str | None = None,
        source_handle: str | None = None,
        target_handle: str | None = None,
    ) -> PreparedRequest:
        """Prepare a request to create a new edge.

        Accepts either the envelope ``request=`` model or the fanned-out
        per-field kwargs the spec lists.
        """
        if request is None:
            if source_block is None or target_block is None:
                raise TypeError("source_block and target_block are required when request is not provided")
            request = WorkflowEdgeCreateRequest(
                id=id,
                source_block=source_block,
                target_block=target_block,
                source_handle=source_handle,
                target_handle=target_handle,
            )
        data = request.model_dump(exclude_none=True)
        data["workflow_id"] = workflow_id
        return PreparedRequest(method="POST", url="/workflows/edges", data=data)

    def prepare_delete(self, edge_id: str) -> PreparedRequest:
        """Prepare a request to delete an edge."""
        return PreparedRequest(method="DELETE", url=f"/workflows/edges/{edge_id}")


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

    def list(
        self,
        workflow_id: str | None = None,
        source_block: str | None = None,
        target_block: str | None = None,
        before: str | None = None,
        after: str | None = None,
        limit: int | None = None,
    ) -> PaginatedList[WorkflowEdgeDoc]:
        """List all edges for a workflow.

        Args:
            workflow_id: The workflow ID
            source_block: Filter by source block ID (optional)
            target_block: Filter by target block ID (optional)
            before: ID-based pagination cursor (optional)
            after: ID-based pagination cursor (optional)
            limit: Items per page (optional)

        Returns:
            ``PaginatedList[WorkflowEdgeDoc]`` — the canonical list envelope
            ``{"data": [...], "list_metadata": {"before": null, "after": null}}``.
            ID pagination is not yet implemented for this endpoint.
        """
        request = self.prepare_list(
            workflow_id,
            source_block=source_block,
            target_block=target_block,
            before=before,
            after=after,
            limit=limit,
        )
        response = self._client._prepared_request(request)
        result = PaginatedList[WorkflowEdgeDoc](**response)
        result.data = [WorkflowEdgeDoc.model_validate(item) for item in result.data]
        return result

    def get(self, edge_id: str) -> WorkflowEdgeDoc:
        """Get a single edge by ID."""
        request = self.prepare_get(edge_id)
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

    def delete(self, edge_id: str) -> None:
        """Delete an edge."""
        request = self.prepare_delete(edge_id)
        self._client._prepared_request(request)


class AsyncWorkflowEdges(AsyncAPIResource, WorkflowEdgesMixin):
    """Workflow Edges API wrapper for asynchronous operations.

    Usage: ``await client.workflows.edges.list(workflow_id)``
    """

    async def list(
        self,
        workflow_id: str | None = None,
        source_block: str | None = None,
        target_block: str | None = None,
        before: str | None = None,
        after: str | None = None,
        limit: int | None = None,
    ) -> PaginatedList[WorkflowEdgeDoc]:
        """List all edges for a workflow.

        Returns the canonical
        ``{"data": [...], "list_metadata": {"before": null, "after": null}}``
        pagination envelope.
        """
        request = self.prepare_list(
            workflow_id,
            source_block=source_block,
            target_block=target_block,
            before=before,
            after=after,
            limit=limit,
        )
        response = await self._client._prepared_request(request)
        result = PaginatedList[WorkflowEdgeDoc](**response)
        result.data = [WorkflowEdgeDoc.model_validate(item) for item in result.data]
        return result

    async def get(self, edge_id: str) -> WorkflowEdgeDoc:
        """Get a single edge by ID."""
        request = self.prepare_get(edge_id)
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

    async def delete(self, edge_id: str) -> None:
        """Delete an edge."""
        request = self.prepare_delete(edge_id)
        await self._client._prepared_request(request)

    _coerce_create_request = WorkflowEdges._coerce_create_request
