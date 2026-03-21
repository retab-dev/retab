from typing import Any, Dict, List, Sequence

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.standards import PreparedRequest
from ....types.workflows import (
    WorkflowBlock,
    WorkflowBlockCreateRequest,
    WorkflowBlockUpdateRequest,
)


class WorkflowBlocksMixin:
    """Mixin providing shared prepare methods for workflow block operations."""

    def prepare_list(self, workflow_id: str, subflow_id: str | None = None) -> PreparedRequest:
        """Prepare a request to list all blocks for a workflow."""
        params: Dict[str, Any] = {}
        if subflow_id is not None:
            params["subflow_id"] = subflow_id
        return PreparedRequest(method="GET", url=f"/workflows/{workflow_id}/blocks", params=params or None)

    def prepare_get(self, workflow_id: str, block_id: str) -> PreparedRequest:
        """Prepare a request to get a single block."""
        return PreparedRequest(method="GET", url=f"/workflows/{workflow_id}/blocks/{block_id}")

    def prepare_create(
        self,
        workflow_id: str,
        request: WorkflowBlockCreateRequest,
    ) -> PreparedRequest:
        """Prepare a request to create a new block."""
        data = request.model_dump(exclude_none=True)
        return PreparedRequest(method="POST", url=f"/workflows/{workflow_id}/blocks", data=data)

    def prepare_create_batch(
        self,
        workflow_id: str,
        blocks: Sequence[WorkflowBlockCreateRequest],
    ) -> PreparedRequest:
        """Prepare a request to create multiple blocks at once."""
        return PreparedRequest(
            method="POST",
            url=f"/workflows/{workflow_id}/blocks/batch",
            data=[block.model_dump(exclude_none=True) for block in blocks],
        )

    def prepare_update(
        self,
        workflow_id: str,
        request: WorkflowBlockUpdateRequest,
    ) -> PreparedRequest:
        """Prepare a request to partially update a block."""
        data = request.model_dump(exclude_none=True, exclude={"block_id"})
        return PreparedRequest(method="PATCH", url=f"/workflows/{workflow_id}/blocks/{request.block_id}", data=data)

    def prepare_delete(self, workflow_id: str, block_id: str) -> PreparedRequest:
        """Prepare a request to delete a block (also deletes connected edges)."""
        return PreparedRequest(method="DELETE", url=f"/workflows/{workflow_id}/blocks/{block_id}")


class WorkflowBlocks(SyncAPIResource, WorkflowBlocksMixin):
    """Workflow Blocks API wrapper for synchronous operations.

    Usage: ``client.workflows.blocks.list(workflow_id)``
    """

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
        subflow_id: str | None,
        parent_id: str | None,
    ) -> WorkflowBlockCreateRequest:
        if request is not None:
            return request
        if id is None or type is None:
            raise TypeError("id and type are required when request is not provided")
        return WorkflowBlockCreateRequest(
            id=id,
            type=type,
            label=label,
            position_x=position_x,
            position_y=position_y,
            width=width,
            height=height,
            config=config,
            subflow_id=subflow_id,
            parent_id=parent_id,
        )

    @staticmethod
    def _coerce_update_request(
        request: WorkflowBlockUpdateRequest | None,
        block_id: str | None,
        label: str | None,
        position_x: float | None,
        position_y: float | None,
        width: float | None,
        height: float | None,
        config: dict | None,
        subflow_id: str | None,
        parent_id: str | None,
    ) -> WorkflowBlockUpdateRequest:
        if request is not None:
            return request
        if block_id is None:
            raise TypeError("block_id is required when request is not provided")
        return WorkflowBlockUpdateRequest(
            block_id=block_id,
            label=label,
            position_x=position_x,
            position_y=position_y,
            width=width,
            height=height,
            config=config,
            subflow_id=subflow_id,
            parent_id=parent_id,
        )

    @staticmethod
    def _coerce_batch_requests(
        blocks: Sequence[WorkflowBlockCreateRequest | Dict[str, Any]],
    ) -> List[WorkflowBlockCreateRequest]:
        return [
            block if isinstance(block, WorkflowBlockCreateRequest) else WorkflowBlockCreateRequest.model_validate(block)
            for block in blocks
        ]

    def list(self, workflow_id: str, subflow_id: str | None = None) -> List[WorkflowBlock]:
        """List all blocks for a workflow.

        Args:
            workflow_id: The workflow ID
            subflow_id: Filter by subflow ID (blocks inside a specific subflow)

        Returns:
            List of workflow blocks
        """
        request = self.prepare_list(workflow_id, subflow_id=subflow_id)
        response = self._client._prepared_request(request)
        return [WorkflowBlock.model_validate(item) for item in response]

    def get(self, workflow_id: str, block_id: str) -> WorkflowBlock:
        """Get a single block by ID."""
        request = self.prepare_get(workflow_id, block_id)
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
        subflow_id: str | None = None,
        parent_id: str | None = None,
        request: WorkflowBlockCreateRequest | None = None,
    ) -> WorkflowBlock:
        """Create a new block in a workflow.

        Args:
            workflow_id: The workflow ID
            id: Client-provided block ID (e.g., "extract-1")
            type: Block type (start, extract, parse, classifier, hil, conditional, api_call, end, etc.)
            label: Display label
            position_x: X position on canvas
            position_y: Y position on canvas
            width: Block width (optional)
            height: Block height (optional)
            config: Block-specific configuration dict (optional)
            subflow_id: Parent subflow ID if inside a subflow (optional)
            parent_id: Parent container block ID (optional)
            request: Optional typed request model. When provided, its values are used directly.

        Returns:
            WorkflowBlock: The created block
        """
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
            subflow_id=subflow_id,
            parent_id=parent_id,
        )
        prepared_request = self.prepare_create(workflow_id, create_request)
        response = self._client._prepared_request(prepared_request)
        return WorkflowBlock.model_validate(response)

    def create_batch(
        self,
        workflow_id: str,
        blocks: Sequence[WorkflowBlockCreateRequest | Dict[str, Any]],
    ) -> List[WorkflowBlock]:
        """Create multiple blocks in a single request.

        Args:
            workflow_id: The workflow ID
            blocks: Typed block request models or dicts, each with at least ``id`` and ``type``

        Returns:
            List of created blocks
        """
        batch_requests = self._coerce_batch_requests(blocks)
        prepared_request = self.prepare_create_batch(workflow_id, batch_requests)
        response = self._client._prepared_request(prepared_request)
        return [WorkflowBlock.model_validate(item) for item in response]

    def update(
        self,
        workflow_id: str,
        block_id: str | None = None,
        label: str | None = None,
        position_x: float | None = None,
        position_y: float | None = None,
        width: float | None = None,
        height: float | None = None,
        config: dict | None = None,
        subflow_id: str | None = None,
        parent_id: str | None = None,
        request: WorkflowBlockUpdateRequest | None = None,
    ) -> WorkflowBlock:
        """Update a block with partial data. Only provided fields are updated.

        Returns:
            WorkflowBlock: The updated block
        """
        update_request = self._coerce_update_request(
            request=request,
            block_id=block_id,
            label=label,
            position_x=position_x,
            position_y=position_y,
            width=width,
            height=height,
            config=config,
            subflow_id=subflow_id,
            parent_id=parent_id,
        )
        prepared_request = self.prepare_update(workflow_id, update_request)
        response = self._client._prepared_request(prepared_request)
        return WorkflowBlock.model_validate(response)

    def delete(self, workflow_id: str, block_id: str) -> None:
        """Delete a block and any edges connected to it."""
        request = self.prepare_delete(workflow_id, block_id)
        self._client._prepared_request(request)


class AsyncWorkflowBlocks(AsyncAPIResource, WorkflowBlocksMixin):
    """Workflow Blocks API wrapper for asynchronous operations.

    Usage: ``await client.workflows.blocks.list(workflow_id)``
    """

    async def list(self, workflow_id: str, subflow_id: str | None = None) -> List[WorkflowBlock]:
        """List all blocks for a workflow."""
        request = self.prepare_list(workflow_id, subflow_id=subflow_id)
        response = await self._client._prepared_request(request)
        return [WorkflowBlock.model_validate(item) for item in response]

    async def get(self, workflow_id: str, block_id: str) -> WorkflowBlock:
        """Get a single block by ID."""
        request = self.prepare_get(workflow_id, block_id)
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
        subflow_id: str | None = None,
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
            subflow_id=subflow_id,
            parent_id=parent_id,
        )
        prepared_request = self.prepare_create(workflow_id, create_request)
        response = await self._client._prepared_request(prepared_request)
        return WorkflowBlock.model_validate(response)

    async def create_batch(
        self,
        workflow_id: str,
        blocks: Sequence[WorkflowBlockCreateRequest | Dict[str, Any]],
    ) -> List[WorkflowBlock]:
        """Create multiple blocks in a single request."""
        batch_requests = self._coerce_batch_requests(blocks)
        prepared_request = self.prepare_create_batch(workflow_id, batch_requests)
        response = await self._client._prepared_request(prepared_request)
        return [WorkflowBlock.model_validate(item) for item in response]

    async def update(
        self,
        workflow_id: str,
        block_id: str | None = None,
        label: str | None = None,
        position_x: float | None = None,
        position_y: float | None = None,
        width: float | None = None,
        height: float | None = None,
        config: dict | None = None,
        subflow_id: str | None = None,
        parent_id: str | None = None,
        request: WorkflowBlockUpdateRequest | None = None,
    ) -> WorkflowBlock:
        """Update a block with partial data."""
        update_request = self._coerce_update_request(
            request=request,
            block_id=block_id,
            label=label,
            position_x=position_x,
            position_y=position_y,
            width=width,
            height=height,
            config=config,
            subflow_id=subflow_id,
            parent_id=parent_id,
        )
        prepared_request = self.prepare_update(workflow_id, update_request)
        response = await self._client._prepared_request(prepared_request)
        return WorkflowBlock.model_validate(response)

    async def delete(self, workflow_id: str, block_id: str) -> None:
        """Delete a block and any edges connected to it."""
        request = self.prepare_delete(workflow_id, block_id)
        await self._client._prepared_request(request)
    _coerce_create_request = WorkflowBlocks._coerce_create_request
    _coerce_update_request = WorkflowBlocks._coerce_update_request
    _coerce_batch_requests = WorkflowBlocks._coerce_batch_requests
