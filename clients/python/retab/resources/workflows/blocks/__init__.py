from typing import Any, Dict, cast

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.pagination import PaginatedList
from ....types.standards import PreparedRequest
from ....types.workflows import (
    WorkflowBlock,
    WorkflowBlockCreateRequest,
    UpdateWorkflowBlockRequest,
)
from .executions import AsyncWorkflowBlockExecutions, WorkflowBlockExecutions


class WorkflowBlocksMixin:
    """Mixin providing shared prepare methods for workflow block operations."""

    def prepare_list(
        self,
        workflow_id: str | None = None,
        before: str | None = None,
        after: str | None = None,
        limit: int | None = None,
    ) -> PreparedRequest:
        """Prepare a request to list all blocks for a workflow."""
        params: Dict[str, Any] = {}
        if workflow_id is not None:
            params["workflow_id"] = workflow_id
        if before is not None:
            params["before"] = before
        if after is not None:
            params["after"] = after
        if limit is not None:
            params["limit"] = limit
        return PreparedRequest(method="GET", url="/workflows/blocks", params=params or None)

    def prepare_get(self, block_id: str, workflow_id: str | None = None) -> PreparedRequest:
        """Prepare a request to get a single block."""
        params: Dict[str, Any] = {}
        if workflow_id is not None:
            params["workflow_id"] = workflow_id
        return PreparedRequest(method="GET", url=f"/workflows/blocks/{block_id}", params=params or None)

    def prepare_create(
        self,
        workflow_id: str,
        request: WorkflowBlockCreateRequest | None = None,
        type: str | None = None,
        id: str | None = None,
        label: str | None = None,
        position_x: float | None = None,
        position_y: float | None = None,
        width: float | None = None,
        height: float | None = None,
        config: dict | None = None,
        parent_id: str | None = None,
    ) -> PreparedRequest:
        """Prepare a request to create a new block.

        Accepts either the envelope ``request=`` model or the fanned-out
        ``type``/``id``/``label``/... kwargs the spec lists. When both are
        provided the envelope wins (callers using kwargs leave ``request=None``).
        """
        if request is None:
            if type is None:
                raise TypeError("type is required when request is not provided")
            request = WorkflowBlockCreateRequest(
                id=id,
                type=cast(Any, type),
                label=label or "",
                position_x=position_x if position_x is not None else 0,
                position_y=position_y if position_y is not None else 0,
                width=width,
                height=height,
                config=config,
                parent_id=parent_id,
            )
        data = request.model_dump(exclude_none=True)
        data["workflow_id"] = workflow_id
        return PreparedRequest(method="POST", url="/workflows/blocks", data=data)

    def prepare_update(
        self,
        block_id: str,
        request: UpdateWorkflowBlockRequest | None = None,
        label: str | None = None,
        position_x: float | None = None,
        position_y: float | None = None,
        width: float | None = None,
        height: float | None = None,
        config: dict | None = None,
        parent_id: str | None = None,
        config_mode: str | None = None,
        workflow_id: str | None = None,
    ) -> PreparedRequest:
        """Prepare a request to partially update a block.

        Accepts either the envelope ``request=`` model or the fanned-out
        per-field kwargs the spec lists. When both are provided the
        envelope wins.
        """
        if request is None:
            request = UpdateWorkflowBlockRequest(
                label=label,
                position_x=position_x,
                position_y=position_y,
                width=width,
                height=height,
                config=config,
                parent_id=parent_id,
            )
        data = request.model_dump(exclude_none=True)
        params: Dict[str, Any] = {}
        if config_mode is not None:
            params["config_mode"] = config_mode
        if workflow_id is not None:
            params["workflow_id"] = workflow_id
        return PreparedRequest(
            method="PATCH",
            url=f"/workflows/blocks/{block_id}",
            data=data,
            params=params or None,
        )

    def prepare_delete(self, block_id: str, workflow_id: str | None = None) -> PreparedRequest:
        """Prepare a request to delete a block."""
        params: Dict[str, Any] = {}
        if workflow_id is not None:
            params["workflow_id"] = workflow_id
        return PreparedRequest(method="DELETE", url=f"/workflows/blocks/{block_id}", params=params or None)


class WorkflowBlocks(SyncAPIResource, WorkflowBlocksMixin):
    """Workflow Blocks API wrapper for synchronous operations.

    Sub-clients:
        executions: Per-block execution operations (create, list) against
            ``/workflows/blocks/executions``.
    """

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.executions = WorkflowBlockExecutions(client=client)

    @staticmethod
    def _coerce_create_request(
        request: WorkflowBlockCreateRequest | None,
        id: str | None,
        type: str | None,
        label: str,
        position_x: float,
        position_y: float,
        width: float | None,
        height: float | None,
        config: dict | None,
        parent_id: str | None,
    ) -> WorkflowBlockCreateRequest:
        if request is not None:
            return request
        if id is None or type is None:
            raise TypeError("id and type are required when request is not provided")
        return WorkflowBlockCreateRequest(
            id=id,
            type=cast(Any, type),
            label=label,
            position_x=position_x,
            position_y=position_y,
            width=width,
            height=height,
            config=config,
            parent_id=parent_id,
        )

    @staticmethod
    def _coerce_update_request(
        request: UpdateWorkflowBlockRequest | None,
        label: str | None,
        position_x: float | None,
        position_y: float | None,
        width: float | None,
        height: float | None,
        config: dict | None,
        parent_id: str | None,
    ) -> UpdateWorkflowBlockRequest:
        if request is not None:
            return request
        return UpdateWorkflowBlockRequest(
            label=label,
            position_x=position_x,
            position_y=position_y,
            width=width,
            height=height,
            config=config,
            parent_id=parent_id,
        )

    def list(
        self,
        workflow_id: str | None = None,
        before: str | None = None,
        after: str | None = None,
        limit: int | None = None,
    ) -> PaginatedList[WorkflowBlock]:
        """List all blocks for a workflow."""
        request = self.prepare_list(workflow_id, before=before, after=after, limit=limit)
        response = self._client._prepared_request(request)
        result = PaginatedList[WorkflowBlock](**response)
        result.data = [WorkflowBlock.model_validate(item) for item in result.data]
        return result

    def get(self, block_id: str, workflow_id: str | None = None) -> WorkflowBlock:
        """Get a single block by ID."""
        request = self.prepare_get(block_id, workflow_id=workflow_id)
        response = self._client._prepared_request(request)
        return WorkflowBlock.model_validate(response)

    def create(
        self,
        workflow_id: str,
        id: str | None = None,
        type: str | None = None,
        label: str = "",
        position_x: float = 0,
        position_y: float = 0,
        width: float | None = None,
        height: float | None = None,
        config: dict | None = None,
        parent_id: str | None = None,
        request: WorkflowBlockCreateRequest | None = None,
    ) -> WorkflowBlock:
        """Create a new block in a workflow."""
        create_request = self._coerce_create_request(
            request=request,
            id=id,
            type=type,
            label=label,
            position_x=position_x,
            position_y=position_y,
            width=width,
            height=height,
            config=config,
            parent_id=parent_id,
        )
        prepared_request = self.prepare_create(workflow_id, create_request)
        response = self._client._prepared_request(prepared_request)
        return WorkflowBlock.model_validate(response)

    def update(
        self,
        block_id: str,
        label: str | None = None,
        position_x: float | None = None,
        position_y: float | None = None,
        width: float | None = None,
        height: float | None = None,
        config: dict | None = None,
        parent_id: str | None = None,
        request: UpdateWorkflowBlockRequest | None = None,
        config_mode: str | None = None,
        workflow_id: str | None = None,
    ) -> WorkflowBlock:
        """Update a block with partial data."""
        update_request = self._coerce_update_request(
            request=request,
            label=label,
            position_x=position_x,
            position_y=position_y,
            width=width,
            height=height,
            config=config,
            parent_id=parent_id,
        )
        prepared_request = self.prepare_update(
            block_id,
            update_request,
            config_mode=config_mode,
            workflow_id=workflow_id,
        )
        response = self._client._prepared_request(prepared_request)
        return WorkflowBlock.model_validate(response)

    def delete(self, block_id: str, workflow_id: str | None = None) -> None:
        """Delete a block."""
        request = self.prepare_delete(block_id, workflow_id=workflow_id)
        self._client._prepared_request(request)


class AsyncWorkflowBlocks(AsyncAPIResource, WorkflowBlocksMixin):
    """Workflow Blocks API wrapper for asynchronous operations.

    Sub-clients:
        executions: Per-block execution operations (create, list) against
            ``/workflows/blocks/executions``.
    """

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.executions = AsyncWorkflowBlockExecutions(client=client)

    async def list(
        self,
        workflow_id: str | None = None,
        before: str | None = None,
        after: str | None = None,
        limit: int | None = None,
    ) -> PaginatedList[WorkflowBlock]:
        """List all blocks for a workflow."""
        request = self.prepare_list(workflow_id, before=before, after=after, limit=limit)
        response = await self._client._prepared_request(request)
        result = PaginatedList[WorkflowBlock](**response)
        result.data = [WorkflowBlock.model_validate(item) for item in result.data]
        return result

    async def get(self, block_id: str, workflow_id: str | None = None) -> WorkflowBlock:
        """Get a single block by ID."""
        request = self.prepare_get(block_id, workflow_id=workflow_id)
        response = await self._client._prepared_request(request)
        return WorkflowBlock.model_validate(response)

    async def create(
        self,
        workflow_id: str,
        id: str | None = None,
        type: str | None = None,
        label: str = "",
        position_x: float = 0,
        position_y: float = 0,
        width: float | None = None,
        height: float | None = None,
        config: dict | None = None,
        parent_id: str | None = None,
        request: WorkflowBlockCreateRequest | None = None,
    ) -> WorkflowBlock:
        """Create a new block in a workflow."""
        create_request = self._coerce_create_request(
            request=request,
            id=id,
            type=type,
            label=label,
            position_x=position_x,
            position_y=position_y,
            width=width,
            height=height,
            config=config,
            parent_id=parent_id,
        )
        prepared_request = self.prepare_create(workflow_id, create_request)
        response = await self._client._prepared_request(prepared_request)
        return WorkflowBlock.model_validate(response)

    async def update(
        self,
        block_id: str,
        label: str | None = None,
        position_x: float | None = None,
        position_y: float | None = None,
        width: float | None = None,
        height: float | None = None,
        config: dict | None = None,
        parent_id: str | None = None,
        request: UpdateWorkflowBlockRequest | None = None,
        config_mode: str | None = None,
        workflow_id: str | None = None,
    ) -> WorkflowBlock:
        """Update a block with partial data."""
        update_request = self._coerce_update_request(
            request=request,
            label=label,
            position_x=position_x,
            position_y=position_y,
            width=width,
            height=height,
            config=config,
            parent_id=parent_id,
        )
        prepared_request = self.prepare_update(
            block_id,
            update_request,
            config_mode=config_mode,
            workflow_id=workflow_id,
        )
        response = await self._client._prepared_request(prepared_request)
        return WorkflowBlock.model_validate(response)

    async def delete(self, block_id: str, workflow_id: str | None = None) -> None:
        """Delete a block."""
        request = self.prepare_delete(block_id, workflow_id=workflow_id)
        await self._client._prepared_request(request)

    _coerce_create_request = WorkflowBlocks._coerce_create_request
    _coerce_update_request = WorkflowBlocks._coerce_update_request


__all__ = [
    "WorkflowBlocks",
    "AsyncWorkflowBlocks",
    "WorkflowBlocksMixin",
    # NOTE: `WorkflowBlockExecutions` / `AsyncWorkflowBlockExecutions`
    # are imported above to wire `self.executions = …` on the resource
    # class, but they are NOT in `__all__`. The single canonical access
    # path is `client.workflows.blocks.executions` at runtime, and
    # `from retab.resources.workflows.blocks.executions import
    # WorkflowBlockExecutions` for direct imports.
]
