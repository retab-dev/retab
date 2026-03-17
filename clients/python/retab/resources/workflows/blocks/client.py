from typing import Any, Dict, List, Optional

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.standards import PreparedRequest
from ....types.workflows import WorkflowBlock


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
        id: str,
        type: str,
        label: str = "",
        position_x: float = 0,
        position_y: float = 0,
        width: float | None = None,
        height: float | None = None,
        config: dict | None = None,
        subflow_id: str | None = None,
        parent_id: str | None = None,
    ) -> PreparedRequest:
        """Prepare a request to create a new block."""
        data: Dict[str, Any] = {
            "id": id,
            "type": type,
            "label": label,
            "position_x": position_x,
            "position_y": position_y,
        }
        if width is not None:
            data["width"] = width
        if height is not None:
            data["height"] = height
        if config is not None:
            data["config"] = config
        if subflow_id is not None:
            data["subflow_id"] = subflow_id
        if parent_id is not None:
            data["parent_id"] = parent_id
        return PreparedRequest(method="POST", url=f"/workflows/{workflow_id}/blocks", data=data)

    def prepare_create_batch(self, workflow_id: str, blocks: List[Dict[str, Any]]) -> PreparedRequest:
        """Prepare a request to create multiple blocks at once."""
        return PreparedRequest(method="POST", url=f"/workflows/{workflow_id}/blocks/batch", data=blocks)

    def prepare_update(
        self,
        workflow_id: str,
        block_id: str,
        label: str | None = None,
        position_x: float | None = None,
        position_y: float | None = None,
        width: float | None = None,
        height: float | None = None,
        config: dict | None = None,
        subflow_id: str | None = None,
        parent_id: str | None = None,
    ) -> PreparedRequest:
        """Prepare a request to partially update a block."""
        data: Dict[str, Any] = {}
        if label is not None:
            data["label"] = label
        if position_x is not None:
            data["position_x"] = position_x
        if position_y is not None:
            data["position_y"] = position_y
        if width is not None:
            data["width"] = width
        if height is not None:
            data["height"] = height
        if config is not None:
            data["config"] = config
        if subflow_id is not None:
            data["subflow_id"] = subflow_id
        if parent_id is not None:
            data["parent_id"] = parent_id
        return PreparedRequest(method="PATCH", url=f"/workflows/{workflow_id}/blocks/{block_id}", data=data)

    def prepare_delete(self, workflow_id: str, block_id: str) -> PreparedRequest:
        """Prepare a request to delete a block (also deletes connected edges)."""
        return PreparedRequest(method="DELETE", url=f"/workflows/{workflow_id}/blocks/{block_id}")


class WorkflowBlocks(SyncAPIResource, WorkflowBlocksMixin):
    """Workflow Blocks API wrapper for synchronous operations.

    Usage: ``client.workflows.blocks.list(workflow_id)``
    """

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

        id: str,
        type: str,
        label: str = "",
        position_x: float = 0,
        position_y: float = 0,
        width: float | None = None,
        height: float | None = None,
        config: dict | None = None,
        subflow_id: str | None = None,
        parent_id: str | None = None,
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

        Returns:
            WorkflowBlock: The created block
        """
        request = self.prepare_create(
            workflow_id, id=id, type=type, label=label,
            position_x=position_x, position_y=position_y,
            width=width, height=height, config=config,
            subflow_id=subflow_id, parent_id=parent_id,
        )
        response = self._client._prepared_request(request)
        return WorkflowBlock.model_validate(response)

    def create_batch(self, workflow_id: str, blocks: List[Dict[str, Any]]) -> List[WorkflowBlock]:
        """Create multiple blocks in a single request.

        Args:
            workflow_id: The workflow ID
            blocks: List of block dicts, each with at least ``id`` and ``type``

        Returns:
            List of created blocks
        """
        request = self.prepare_create_batch(workflow_id, blocks)
        response = self._client._prepared_request(request)
        return [WorkflowBlock.model_validate(item) for item in response]

    def update(
        self,
        workflow_id: str,
        block_id: str,

        label: str | None = None,
        position_x: float | None = None,
        position_y: float | None = None,
        width: float | None = None,
        height: float | None = None,
        config: dict | None = None,
        subflow_id: str | None = None,
        parent_id: str | None = None,
    ) -> WorkflowBlock:
        """Update a block with partial data. Only provided fields are updated.

        Returns:
            WorkflowBlock: The updated block
        """
        request = self.prepare_update(
            workflow_id, block_id, label=label,
            position_x=position_x, position_y=position_y,
            width=width, height=height, config=config,
            subflow_id=subflow_id, parent_id=parent_id,
        )
        response = self._client._prepared_request(request)
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

        id: str,
        type: str,
        label: str = "",
        position_x: float = 0,
        position_y: float = 0,
        width: float | None = None,
        height: float | None = None,
        config: dict | None = None,
        subflow_id: str | None = None,
        parent_id: str | None = None,
    ) -> WorkflowBlock:
        """Create a new block in a workflow."""
        request = self.prepare_create(
            workflow_id, id=id, type=type, label=label,
            position_x=position_x, position_y=position_y,
            width=width, height=height, config=config,
            subflow_id=subflow_id, parent_id=parent_id,
        )
        response = await self._client._prepared_request(request)
        return WorkflowBlock.model_validate(response)

    async def create_batch(self, workflow_id: str, blocks: List[Dict[str, Any]]) -> List[WorkflowBlock]:
        """Create multiple blocks in a single request."""
        request = self.prepare_create_batch(workflow_id, blocks)
        response = await self._client._prepared_request(request)
        return [WorkflowBlock.model_validate(item) for item in response]

    async def update(
        self,
        workflow_id: str,
        block_id: str,

        label: str | None = None,
        position_x: float | None = None,
        position_y: float | None = None,
        width: float | None = None,
        height: float | None = None,
        config: dict | None = None,
        subflow_id: str | None = None,
        parent_id: str | None = None,
    ) -> WorkflowBlock:
        """Update a block with partial data."""
        request = self.prepare_update(
            workflow_id, block_id, label=label,
            position_x=position_x, position_y=position_y,
            width=width, height=height, config=config,
            subflow_id=subflow_id, parent_id=parent_id,
        )
        response = await self._client._prepared_request(request)
        return WorkflowBlock.model_validate(response)

    async def delete(self, workflow_id: str, block_id: str) -> None:
        """Delete a block and any edges connected to it."""
        request = self.prepare_delete(workflow_id, block_id)
        await self._client._prepared_request(request)
