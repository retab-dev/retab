from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.pagination import PaginatedList
from ....types.standards import PreparedRequest
from ....types.workflows import (
    WorkflowBlock,
    WorkflowBlockCreateRequest,
    UpdateWorkflowBlockRequest,
)


class WorkflowBlocksMixin:
    """Mixin providing shared prepare methods for workflow block operations."""

    def prepare_list(self, workflow_id: str) -> PreparedRequest:
        """Prepare a request to list all blocks for a workflow."""
        return PreparedRequest(method="GET", url=f"/workflows/blocks?workflow_id={workflow_id}")

    def prepare_get(self, block_id: str) -> PreparedRequest:
        """Prepare a request to get a single block."""
        return PreparedRequest(method="GET", url=f"/workflows/blocks/{block_id}")

    def prepare_create(
        self,
        workflow_id: str,
        request: WorkflowBlockCreateRequest,
    ) -> PreparedRequest:
        """Prepare a request to create a new block."""
        data = request.model_dump(exclude_none=True)
        data["workflow_id"] = workflow_id
        return PreparedRequest(method="POST", url="/workflows/blocks", data=data)

    def prepare_update(
        self,
        block_id: str,
        request: UpdateWorkflowBlockRequest,
    ) -> PreparedRequest:
        """Prepare a request to partially update a block."""
        data = request.model_dump(exclude_none=True)
        return PreparedRequest(method="PATCH", url=f"/workflows/blocks/{block_id}", data=data)

    def prepare_delete(self, block_id: str) -> PreparedRequest:
        """Prepare a request to delete a block."""
        return PreparedRequest(method="DELETE", url=f"/workflows/blocks/{block_id}")


class WorkflowBlocks(SyncAPIResource, WorkflowBlocksMixin):
    """Workflow Blocks API wrapper for synchronous operations."""

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
            type=type,
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

    def list(self, workflow_id: str) -> PaginatedList[WorkflowBlock]:
        """List all blocks for a workflow."""
        request = self.prepare_list(workflow_id)
        response = self._client._prepared_request(request)
        result = PaginatedList[WorkflowBlock](**response)
        result.data = [WorkflowBlock.model_validate(item) for item in result.data]
        return result

    def get(self, block_id: str) -> WorkflowBlock:
        """Get a single block by ID."""
        request = self.prepare_get(block_id)
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
        prepared_request = self.prepare_update(block_id, update_request)
        response = self._client._prepared_request(prepared_request)
        return WorkflowBlock.model_validate(response)

    def delete(self, block_id: str) -> None:
        """Delete a block."""
        request = self.prepare_delete(block_id)
        self._client._prepared_request(request)


class AsyncWorkflowBlocks(AsyncAPIResource, WorkflowBlocksMixin):
    """Workflow Blocks API wrapper for asynchronous operations."""

    async def list(self, workflow_id: str) -> PaginatedList[WorkflowBlock]:
        """List all blocks for a workflow."""
        request = self.prepare_list(workflow_id)
        response = await self._client._prepared_request(request)
        result = PaginatedList[WorkflowBlock](**response)
        result.data = [WorkflowBlock.model_validate(item) for item in result.data]
        return result

    async def get(self, block_id: str) -> WorkflowBlock:
        """Get a single block by ID."""
        request = self.prepare_get(block_id)
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
        prepared_request = self.prepare_update(block_id, update_request)
        response = await self._client._prepared_request(prepared_request)
        return WorkflowBlock.model_validate(response)

    async def delete(self, block_id: str) -> None:
        """Delete a block."""
        request = self.prepare_delete(block_id)
        await self._client._prepared_request(request)

    _coerce_create_request = WorkflowBlocks._coerce_create_request
    _coerce_update_request = WorkflowBlocks._coerce_update_request
