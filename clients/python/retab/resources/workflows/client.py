from typing import Any, Dict, List, Optional

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.pagination import PaginatedList, PaginationOrder
from ...types.standards import PreparedRequest
from ...types.workflows import Workflow, WorkflowWithEntities
from .runs import WorkflowRuns, AsyncWorkflowRuns
from .blocks import WorkflowBlocks, AsyncWorkflowBlocks
from .edges import WorkflowEdges, AsyncWorkflowEdges


class WorkflowsMixin:
    """Mixin providing shared methods for workflow operations."""

    def prepare_get(self, workflow_id: str) -> PreparedRequest:
        """Prepare a request to get a workflow by ID."""
        return PreparedRequest(method="GET", url=f"/workflows/{workflow_id}")

    def prepare_list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: PaginationOrder = "desc",
        sort_by: str = "updated_at",
        fields: str | None = None,
    ) -> PreparedRequest:
        """Prepare a request to list workflows with pagination."""
        params = {
            "before": before,
            "after": after,
            "limit": limit,
            "order": order,
            "sort_by": sort_by,
            "fields": fields,
        }
        params = {key: value for key, value in params.items() if value is not None}
        return PreparedRequest(method="GET", url="/workflows", params=params)

    def prepare_create(
        self,
        name: str = "Untitled Workflow",
        description: str = "",
    ) -> PreparedRequest:
        """Prepare a request to create a new workflow."""
        data: Dict[str, Any] = {"name": name, "description": description}
        return PreparedRequest(method="POST", url="/workflows", data=data)

    def prepare_update(
        self,
        workflow_id: str,
        name: str | None = None,
        description: str | None = None,
        email_senders_whitelist: List[str] | None = None,
        email_domains_whitelist: List[str] | None = None,
    ) -> PreparedRequest:
        """Prepare a request to update a workflow."""
        data: Dict[str, Any] = {}
        if name is not None:
            data["name"] = name
        if description is not None:
            data["description"] = description
        if email_senders_whitelist is not None or email_domains_whitelist is not None:
            data["email_trigger"] = {}
            if email_senders_whitelist is not None:
                data["email_trigger"]["allowed_senders"] = email_senders_whitelist
            if email_domains_whitelist is not None:
                data["email_trigger"]["allowed_domains"] = email_domains_whitelist
        return PreparedRequest(method="PATCH", url=f"/workflows/{workflow_id}", data=data)

    def prepare_delete(self, workflow_id: str) -> PreparedRequest:
        """Prepare a request to delete a workflow."""
        return PreparedRequest(method="DELETE", url=f"/workflows/{workflow_id}")

    def prepare_publish(self, workflow_id: str, description: str = "") -> PreparedRequest:
        """Prepare a request to publish a workflow."""
        data: Dict[str, Any] = {"description": description}
        return PreparedRequest(method="POST", url=f"/workflows/{workflow_id}/publish", data=data)

    def prepare_duplicate(self, workflow_id: str) -> PreparedRequest:
        """Prepare a request to duplicate a workflow."""
        return PreparedRequest(method="POST", url=f"/workflows/{workflow_id}/duplicate", data={})

    def prepare_get_entities(self, workflow_id: str) -> PreparedRequest:
        """Prepare a request to get a workflow with all its entities (blocks, edges, subflows)."""
        return PreparedRequest(method="GET", url=f"/workflows/{workflow_id}/entities")


class Workflows(SyncAPIResource, WorkflowsMixin):
    """Workflows API wrapper for synchronous operations.

    Sub-clients:
        runs: Workflow run operations (create, get, list, cancel, restart, resume)
        blocks: Workflow block CRUD operations
        edges: Workflow edge CRUD operations
    """

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.runs = WorkflowRuns(client=client)
        self.blocks = WorkflowBlocks(client=client)
        self.edges = WorkflowEdges(client=client)

    def get(self, workflow_id: str) -> Workflow:
        """Get a workflow by ID."""
        request = self.prepare_get(workflow_id)
        response = self._client._prepared_request(request)
        return Workflow.model_validate(response)

    def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: PaginationOrder = "desc",
        sort_by: str = "updated_at",
        fields: str | None = None,
    ) -> PaginatedList:
        """List workflows with pagination."""
        request = self.prepare_list(
            before=before,
            after=after,
            limit=limit,
            order=order,
            sort_by=sort_by,
            fields=fields,
        )
        response = self._client._prepared_request(request)
        result = PaginatedList(**response)
        if fields is None:
            result.data = [Workflow.model_validate(item) if isinstance(item, dict) else item for item in result.data]
        return result

    def create(self, name: str = "Untitled Workflow", description: str = "") -> Workflow:
        """Create a new workflow.

        Args:
            name: Workflow name
            description: Workflow description

        Returns:
            Workflow: The created workflow (unpublished, with a default start node)
        """
        request = self.prepare_create(name=name, description=description)
        response = self._client._prepared_request(request)
        return Workflow.model_validate(response)

    def update(
        self,
        workflow_id: str,

        name: str | None = None,
        description: str | None = None,
        email_senders_whitelist: List[str] | None = None,
        email_domains_whitelist: List[str] | None = None,
    ) -> Workflow:
        """Update a workflow's metadata.

        Args:
            workflow_id: The workflow ID
            name: New name (optional)
            description: New description (optional)
            email_senders_whitelist: Allowed sender emails for email triggers (optional)
            email_domains_whitelist: Allowed sender domains for email triggers (optional)

        Returns:
            Workflow: The updated workflow
        """
        request = self.prepare_update(
            workflow_id, name=name, description=description,
            email_senders_whitelist=email_senders_whitelist,
            email_domains_whitelist=email_domains_whitelist,
        )
        response = self._client._prepared_request(request)
        return Workflow.model_validate(response)

    def delete(self, workflow_id: str) -> None:
        """Delete a workflow and all its associated entities (blocks, edges, snapshots)."""
        request = self.prepare_delete(workflow_id)
        self._client._prepared_request(request)

    def publish(self, workflow_id: str, description: str = "") -> Workflow:
        """Publish a workflow's current draft as a new version.

        Args:
            workflow_id: The workflow ID
            description: Optional description for this published version

        Returns:
            Workflow: The updated workflow with new published metadata
        """
        request = self.prepare_publish(workflow_id, description=description)
        response = self._client._prepared_request(request)
        return Workflow.model_validate(response)

    def duplicate(self, workflow_id: str) -> Workflow:
        """Duplicate a workflow with all its blocks and edges.

        Returns:
            Workflow: The new duplicated workflow (unpublished)
        """
        request = self.prepare_duplicate(workflow_id)
        response = self._client._prepared_request(request)
        return Workflow.model_validate(response)

    def get_entities(self, workflow_id: str) -> WorkflowWithEntities:
        """Get a workflow with all its entities (blocks, edges, subflows).

        Returns:
            WorkflowWithEntities: The workflow with its full graph structure.
                Use ``.start_nodes`` and ``.start_json_nodes`` to discover input nodes.
        """
        request = self.prepare_get_entities(workflow_id)
        response = self._client._prepared_request(request)
        return WorkflowWithEntities.model_validate(response)


class AsyncWorkflows(AsyncAPIResource, WorkflowsMixin):
    """Workflows API wrapper for asynchronous operations.

    Sub-clients:
        runs: Workflow run operations (create, get, list, cancel, restart, resume)
        blocks: Workflow block CRUD operations
        edges: Workflow edge CRUD operations
    """

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.runs = AsyncWorkflowRuns(client=client)
        self.blocks = AsyncWorkflowBlocks(client=client)
        self.edges = AsyncWorkflowEdges(client=client)

    async def get(self, workflow_id: str) -> Workflow:
        """Get a workflow by ID."""
        request = self.prepare_get(workflow_id)
        response = await self._client._prepared_request(request)
        return Workflow.model_validate(response)

    async def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 10,
        order: PaginationOrder = "desc",
        sort_by: str = "updated_at",
        fields: str | None = None,
    ) -> PaginatedList:
        """List workflows with pagination."""
        request = self.prepare_list(
            before=before,
            after=after,
            limit=limit,
            order=order,
            sort_by=sort_by,
            fields=fields,
        )
        response = await self._client._prepared_request(request)
        result = PaginatedList(**response)
        if fields is None:
            result.data = [Workflow.model_validate(item) if isinstance(item, dict) else item for item in result.data]
        return result

    async def create(self, name: str = "Untitled Workflow", description: str = "") -> Workflow:
        """Create a new workflow."""
        request = self.prepare_create(name=name, description=description)
        response = await self._client._prepared_request(request)
        return Workflow.model_validate(response)

    async def update(
        self,
        workflow_id: str,

        name: str | None = None,
        description: str | None = None,
        email_senders_whitelist: List[str] | None = None,
        email_domains_whitelist: List[str] | None = None,
    ) -> Workflow:
        """Update a workflow's metadata."""
        request = self.prepare_update(
            workflow_id, name=name, description=description,
            email_senders_whitelist=email_senders_whitelist,
            email_domains_whitelist=email_domains_whitelist,
        )
        response = await self._client._prepared_request(request)
        return Workflow.model_validate(response)

    async def delete(self, workflow_id: str) -> None:
        """Delete a workflow and all its associated entities."""
        request = self.prepare_delete(workflow_id)
        await self._client._prepared_request(request)

    async def publish(self, workflow_id: str, description: str = "") -> Workflow:
        """Publish a workflow's current draft as a new version."""
        request = self.prepare_publish(workflow_id, description=description)
        response = await self._client._prepared_request(request)
        return Workflow.model_validate(response)

    async def duplicate(self, workflow_id: str) -> Workflow:
        """Duplicate a workflow with all its blocks and edges."""
        request = self.prepare_duplicate(workflow_id)
        response = await self._client._prepared_request(request)
        return Workflow.model_validate(response)

    async def get_entities(self, workflow_id: str) -> WorkflowWithEntities:
        """Get a workflow with all its entities (blocks, edges, subflows)."""
        request = self.prepare_get_entities(workflow_id)
        response = await self._client._prepared_request(request)
        return WorkflowWithEntities.model_validate(response)
