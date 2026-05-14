from typing import Any, Dict, List, Sequence

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.pagination import PaginatedList
from ....types.standards import PreparedRequest
from ....types.workflows import (
    BlockConfigVersion,
    BlockResolvedSchemasResponse,
    BlockSimulation,
    WorkflowBlock,
    WorkflowBlockCreateRequest,
    WorkflowBlockUpdateRequest,
)


class WorkflowBlocksMixin:
    """Mixin providing shared prepare methods for workflow block operations."""

    def prepare_list(self, workflow_id: str) -> PreparedRequest:
        """Prepare a request to list all blocks for a workflow."""
        return PreparedRequest(method="GET", url=f"/workflows/{workflow_id}/blocks")

    def prepare_get(self, workflow_id: str, block_id: str) -> PreparedRequest:
        """Prepare a request to get a single block."""
        return PreparedRequest(method="GET", url=f"/workflows/{workflow_id}/blocks/{block_id}")

    def prepare_get_resolved_schemas(self, workflow_id: str, block_id: str) -> PreparedRequest:
        """Prepare a request to get graph-derived schemas for one block."""
        return PreparedRequest(method="GET", url=f"/workflows/{workflow_id}/blocks/{block_id}/resolved-schemas")

    def prepare_config_history(self, workflow_id: str, block_id: str) -> PreparedRequest:
        """Prepare a request to fetch the config-version timeline for a block."""
        return PreparedRequest(
            method="GET",
            url=f"/workflows/{workflow_id}/blocks/{block_id}/config-history",
        )

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

    def prepare_list_simulations(
        self,
        run_id: str,
        block_id: str,
        limit: int | None = None,
    ) -> PreparedRequest:
        """Prepare a request to list recent simulation results for a block in a run."""
        params: Dict[str, Any] = {}
        if limit is not None:
            params["limit"] = limit
        return PreparedRequest(
            method="GET",
            url=f"/workflows/runs/{run_id}/steps/{block_id}/simulations",
            params=params or None,
        )

    def prepare_simulate(
        self,
        run_id: str,
        block_id: str,
        n_consensus: int | None = None,
        step_id: str | None = None,
        check_eligibility: bool = True,
    ) -> PreparedRequest:
        """Prepare a request to replay one block with the current draft config.

        Note: this is keyed by ``run_id`` (the workflow run whose inputs are
        replayed), NOT by ``workflow_id`` — the backend route lives under
        ``/v1/workflows/runs/{run_id}/steps/{block_id}/simulate``.
        """
        params: Dict[str, Any] = {}
        if n_consensus is not None:
            params["n_consensus"] = n_consensus
        if step_id is not None:
            params["step_id"] = step_id
        # Only send when overriding the default — the backend defaults to
        # ``True``, and sending ``check_eligibility=true`` redundantly would
        # be noise.
        if not check_eligibility:
            params["check_eligibility"] = False
        return PreparedRequest(
            method="POST",
            url=f"/workflows/runs/{run_id}/steps/{block_id}/simulate",
            params=params or None,
        )


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
        request: WorkflowBlockUpdateRequest | None,
        block_id: str | None,
        label: str | None,
        position_x: float | None,
        position_y: float | None,
        width: float | None,
        height: float | None,
        config: dict | None,
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

    def list(self, workflow_id: str) -> PaginatedList[WorkflowBlock]:
        """List all blocks for a workflow.

        Args:
            workflow_id: The workflow ID

        Returns:
            ``PaginatedList[WorkflowBlock]`` — the canonical list envelope
            ``{"data": [...], "list_metadata": {"before": null, "after": null}}``.
            Cursor pagination is not yet implemented; ``list_metadata`` is
            always ``{before: None, after: None}``.
        """
        request = self.prepare_list(workflow_id)
        response = self._client._prepared_request(request)
        result = PaginatedList[WorkflowBlock](**response)
        result.data = [WorkflowBlock.model_validate(item) for item in result.data]
        return result

    def config_history(self, workflow_id: str, block_id: str) -> PaginatedList[BlockConfigVersion]:
        """Return the config-version timeline for a block.

        Each ``data`` entry groups consecutive workflow snapshots in which the
        block's config didn't change, with the snapshot version range and the
        captured config snapshot. Wrapped in the canonical pagination envelope;
        cursor pagination is not yet implemented for this endpoint.
        """
        request = self.prepare_config_history(workflow_id, block_id)
        response = self._client._prepared_request(request)
        result = PaginatedList[BlockConfigVersion](**response)
        result.data = [BlockConfigVersion.model_validate(item) for item in result.data]
        return result

    def get(self, workflow_id: str, block_id: str) -> WorkflowBlock:
        """Get a single block by ID."""
        request = self.prepare_get(workflow_id, block_id)
        response = self._client._prepared_request(request)
        return WorkflowBlock.model_validate(response)

    def get_resolved_schemas(self, workflow_id: str, block_id: str) -> BlockResolvedSchemasResponse:
        """Get graph-derived schemas for one current-draft block."""
        request = self.prepare_get_resolved_schemas(workflow_id, block_id)
        response = self._client._prepared_request(request)
        return BlockResolvedSchemasResponse.model_validate(response)

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
        """Create a new block in a workflow.

        Args:
            workflow_id: The workflow ID
            id: Client-provided block ID (e.g., "extract-1")
            type: Block type (start, extract, parse, classifier, hil, conditional, api_call, etc.)
            label: Display label
            position_x: X position on canvas
            position_y: Y position on canvas
            width: Block width (optional)
            height: Block height (optional)
            config: Block-specific configuration dict (optional)
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
            parent_id=parent_id,
        )
        prepared_request = self.prepare_update(workflow_id, update_request)
        response = self._client._prepared_request(prepared_request)
        return WorkflowBlock.model_validate(response)

    def delete(self, workflow_id: str, block_id: str) -> None:
        """Delete a block and any edges connected to it."""
        request = self.prepare_delete(workflow_id, block_id)
        self._client._prepared_request(request)

    def list_simulations(
        self,
        run_id: str,
        block_id: str,
        limit: int | None = None,
    ) -> PaginatedList[BlockSimulation]:
        """List recent simulation results for a block in a workflow run.

        Returns the canonical
        ``{"data": [...], "list_metadata": {"before": null, "after": null}}``
        pagination envelope. Cursor pagination is not yet implemented; pass
        ``limit`` to bound the page size (default server-side: 20, max 100).
        """
        request = self.prepare_list_simulations(run_id, block_id, limit=limit)
        response = self._client._prepared_request(request)
        result = PaginatedList[BlockSimulation](**response)
        result.data = [BlockSimulation.model_validate(item) for item in result.data]
        return result

    def simulate(
        self,
        run_id: str,
        block_id: str,
        n_consensus: int | None = None,
        step_id: str | None = None,
        check_eligibility: bool = True,
    ) -> BlockSimulation:
        """Replay one block using inputs from a previous run + current draft config.

        Args:
            run_id: ID of the workflow run whose recorded inputs feed this
                block. The endpoint optionally verifies the upstream
                subgraph hasn't drifted since this run executed (see
                ``check_eligibility``).
            block_id: ID of the block to simulate.
            n_consensus: Override the block's ``n_consensus`` for this run
                only. Allowed values: 3, 5, 7. Only meaningful for
                ``extract`` / ``split`` / ``classifier`` blocks.
            step_id: For blocks inside a ``for_each`` / ``while_loop``,
                pick a specific iteration. Defaults to the base/first
                available step.
            check_eligibility: When ``True`` (default), the backend
                rejects the request with 409 if the upstream subgraph has
                changed since ``run_id`` executed. Pass ``False`` to skip
                the check.

        Returns:
            ``BlockSimulation`` with handle inputs/outputs, optional
            ``artifact`` ref to the persisted result, and timing info.

        Raises:
            APIError: 409 if the run is no longer eligible (drift), 400 if
                ``n_consensus`` isn't 3/5/7, 404 if the run or block is gone.
        """
        request = self.prepare_simulate(
            run_id,
            block_id,
            n_consensus=n_consensus,
            step_id=step_id,
            check_eligibility=check_eligibility,
        )
        response = self._client._prepared_request(request)
        return BlockSimulation.model_validate(response)


class AsyncWorkflowBlocks(AsyncAPIResource, WorkflowBlocksMixin):
    """Workflow Blocks API wrapper for asynchronous operations.

    Usage: ``await client.workflows.blocks.list(workflow_id)``
    """

    async def list(self, workflow_id: str) -> PaginatedList[WorkflowBlock]:
        """List all blocks for a workflow.

        Returns the canonical
        ``{"data": [...], "list_metadata": {"before": null, "after": null}}``
        pagination envelope.
        """
        request = self.prepare_list(workflow_id)
        response = await self._client._prepared_request(request)
        result = PaginatedList[WorkflowBlock](**response)
        result.data = [WorkflowBlock.model_validate(item) for item in result.data]
        return result

    async def config_history(self, workflow_id: str, block_id: str) -> PaginatedList[BlockConfigVersion]:
        """Return the config-version timeline for a block."""
        request = self.prepare_config_history(workflow_id, block_id)
        response = await self._client._prepared_request(request)
        result = PaginatedList[BlockConfigVersion](**response)
        result.data = [BlockConfigVersion.model_validate(item) for item in result.data]
        return result

    async def get(self, workflow_id: str, block_id: str) -> WorkflowBlock:
        """Get a single block by ID."""
        request = self.prepare_get(workflow_id, block_id)
        response = await self._client._prepared_request(request)
        return WorkflowBlock.model_validate(response)

    async def get_resolved_schemas(self, workflow_id: str, block_id: str) -> BlockResolvedSchemasResponse:
        """Get graph-derived schemas for one current-draft block."""
        request = self.prepare_get_resolved_schemas(workflow_id, block_id)
        response = await self._client._prepared_request(request)
        return BlockResolvedSchemasResponse.model_validate(response)

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
            parent_id=parent_id,
        )
        prepared_request = self.prepare_update(workflow_id, update_request)
        response = await self._client._prepared_request(prepared_request)
        return WorkflowBlock.model_validate(response)

    async def delete(self, workflow_id: str, block_id: str) -> None:
        """Delete a block and any edges connected to it."""
        request = self.prepare_delete(workflow_id, block_id)
        await self._client._prepared_request(request)

    async def list_simulations(
        self,
        run_id: str,
        block_id: str,
        limit: int | None = None,
    ) -> PaginatedList[BlockSimulation]:
        """List recent simulation results for a block in a workflow run."""
        request = self.prepare_list_simulations(run_id, block_id, limit=limit)
        response = await self._client._prepared_request(request)
        result = PaginatedList[BlockSimulation](**response)
        result.data = [BlockSimulation.model_validate(item) for item in result.data]
        return result

    async def simulate(
        self,
        run_id: str,
        block_id: str,
        n_consensus: int | None = None,
        step_id: str | None = None,
        check_eligibility: bool = True,
    ) -> BlockSimulation:
        """Replay one block using inputs from a previous run + current draft config."""
        request = self.prepare_simulate(
            run_id,
            block_id,
            n_consensus=n_consensus,
            step_id=step_id,
            check_eligibility=check_eligibility,
        )
        response = await self._client._prepared_request(request)
        return BlockSimulation.model_validate(response)

    _coerce_create_request = WorkflowBlocks._coerce_create_request
    _coerce_update_request = WorkflowBlocks._coerce_update_request
    _coerce_batch_requests = WorkflowBlocks._coerce_batch_requests
