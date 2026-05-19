from typing import Any, Dict, List, Optional

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.pagination import PaginatedList, PaginationOrder
from ...types.standards import PreparedRequest
from ...types.workflows import (
    Workflow,
    WorkflowDiagnosisResponse,
    WorkflowResolvedSchemasResponse,
    WorkflowSnapshot,
    WorkflowWithEntities,
)
from .runs import WorkflowRuns, AsyncWorkflowRuns
from .reviews import WorkflowReviews, AsyncWorkflowReviews
from .artifacts import AsyncWorkflowArtifacts, WorkflowArtifacts
from .blocks import WorkflowBlocks, AsyncWorkflowBlocks
from .edges import WorkflowEdges, AsyncWorkflowEdges
from .experiments import AsyncWorkflowExperiments, WorkflowExperiments
from .specs import AsyncWorkflowSpecs, WorkflowSpecs
from .tests import AsyncWorkflowTests, WorkflowTests


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
        email_trigger: Dict[str, List[str]] | None = None,
    ) -> PreparedRequest:
        """Prepare a request to update a workflow."""
        data: Dict[str, Any] = {}
        if name is not None:
            data["name"] = name
        if description is not None:
            data["description"] = description
        if email_trigger is not None:
            data["email_trigger"] = email_trigger
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
        """Prepare a request to get a workflow with all its entities (blocks and edges)."""
        return PreparedRequest(method="GET", url=f"/workflows/{workflow_id}/entities")

    def prepare_get_resolved_schemas(self, workflow_id: str) -> PreparedRequest:
        """Prepare a request to get graph-derived schemas for all current-draft blocks."""
        return PreparedRequest(method="GET", url=f"/workflows/{workflow_id}/resolved-schemas")

    def prepare_list_snapshots(
        self,
        workflow_id: str,
        limit: int | None = None,
    ) -> PreparedRequest:
        """Prepare a request to list published snapshots for a workflow."""
        params: Dict[str, Any] = {}
        if limit is not None:
            params["limit"] = limit
        return PreparedRequest(
            method="GET",
            url=f"/workflows/{workflow_id}/snapshots",
            params=params or None,
        )

    def prepare_diagnose(
        self,
        workflow_id: str,
        blocks: List[Dict[str, Any]],
        edges: List[Dict[str, Any]],
        re_propagate: bool = True,
    ) -> PreparedRequest:
        """Prepare a request to diagnose a workflow's graph structure.

        ``blocks`` and ``edges`` are the in-memory graph payload — for
        diagnosing the persisted draft state, fetch with ``get_entities()``
        first and pass the results.
        """
        data: Dict[str, Any] = {
            "blocks": blocks,
            "edges": edges,
            "re_propagate": re_propagate,
        }
        return PreparedRequest(
            method="POST",
            url=f"/workflows/{workflow_id}/diagnose-graph",
            data=data,
        )


class Workflows(SyncAPIResource, WorkflowsMixin):
    """Workflows API wrapper for synchronous operations.

    Sub-clients:
        runs: Workflow run operations (create, get, list, cancel, restart, resume)
        reviews: HIL review overlay operations (queue, approve, reject)
        blocks: Workflow block CRUD operations
        edges: Workflow edge CRUD operations
        artifacts: Workflow artifact dereference operations
        specs: Declarative workflow YAML validation, planning, apply, and export
    """

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.runs = WorkflowRuns(client=client)
        self.reviews = WorkflowReviews(client=client)
        self.artifacts = WorkflowArtifacts(client=client)
        self.blocks = WorkflowBlocks(client=client)
        self.edges = WorkflowEdges(client=client)
        self.tests = WorkflowTests(client=client)
        self.experiments = WorkflowExperiments(client=client)
        self.specs = WorkflowSpecs(client=client)

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
            Workflow: The created workflow (unpublished, with a default start block)
        """
        request = self.prepare_create(name=name, description=description)
        response = self._client._prepared_request(request)
        return Workflow.model_validate(response)

    def update(
        self,
        workflow_id: str,

        name: str | None = None,
        description: str | None = None,
        email_trigger: Dict[str, List[str]] | None = None,
    ) -> Workflow:
        """Update a workflow's metadata.

        Args:
            workflow_id: The workflow ID
            name: New name (optional)
            description: New description (optional)
            email_trigger: Email trigger allowlist policy (optional)

        Returns:
            Workflow: The updated workflow
        """
        request = self.prepare_update(
            workflow_id, name=name, description=description,
            email_trigger=email_trigger,
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
        """Get a workflow with all its entities (blocks and edges).

        Returns:
            WorkflowWithEntities: The workflow with its full graph structure.
                Use ``.start_blocks`` and ``.start_json_blocks`` to discover input blocks.
        """
        request = self.prepare_get_entities(workflow_id)
        response = self._client._prepared_request(request)
        return WorkflowWithEntities.model_validate(response)

    def get_resolved_schemas(self, workflow_id: str) -> WorkflowResolvedSchemasResponse:
        """Get graph-derived schemas for all current-draft blocks in a workflow."""
        request = self.prepare_get_resolved_schemas(workflow_id)
        response = self._client._prepared_request(request)
        return WorkflowResolvedSchemasResponse.model_validate(response)

    def list_snapshots(
        self,
        workflow_id: str,
        limit: int | None = None,
    ) -> PaginatedList[WorkflowSnapshot]:
        """List published snapshots for a workflow (newest first).

        Returns the canonical
        ``{"data": [...], "list_metadata": {"before": null, "after": null}}``
        pagination envelope. ``limit`` bounds the page size (server default:
        50, max 100). ID pagination is not yet implemented for this
        endpoint.
        """
        request = self.prepare_list_snapshots(workflow_id, limit=limit)
        response = self._client._prepared_request(request)
        result = PaginatedList[WorkflowSnapshot](**response)
        result.data = [WorkflowSnapshot.model_validate(item) for item in result.data]
        return result

    def diagnose(
        self,
        workflow_id: str,
        re_propagate: bool = True,
    ) -> WorkflowDiagnosisResponse:
        """Diagnose the workflow's draft graph for structural issues.

        Fetches the persisted draft entities first, then POSTs them to the
        diagnose-graph endpoint. Returns a list of ``issues`` (errors must
        be fixed before publish; warnings are advisory) and ``stats``.

        For diagnosing an in-memory editor graph that hasn't been saved
        yet, build the request directly with :meth:`prepare_diagnose` and
        pass your own ``blocks`` / ``edges`` payloads.
        """
        entities = self.get_entities(workflow_id)
        blocks = [
            {
                "id": block.id,
                "type": block.type,
                "label": block.label,
                "config": block.config,
                "position": {"x": block.position_x, "y": block.position_y},
                "width": block.width,
                "height": block.height,
                "parent_id": block.parent_id,
            }
            for block in entities.blocks
        ]
        edges = [
            {
                "id": edge.id,
                "source": edge.source_block,
                "target": edge.target_block,
                "source_handle": edge.source_handle,
                "target_handle": edge.target_handle,
            }
            for edge in entities.edges
        ]
        request = self.prepare_diagnose(workflow_id, blocks, edges, re_propagate=re_propagate)
        response = self._client._prepared_request(request)
        return WorkflowDiagnosisResponse.model_validate(response)


class AsyncWorkflows(AsyncAPIResource, WorkflowsMixin):
    """Workflows API wrapper for asynchronous operations.

    Sub-clients:
        runs: Workflow run operations (create, get, list, cancel, restart, resume)
        reviews: HIL review overlay operations (queue, approve, reject)
        blocks: Workflow block CRUD operations
        edges: Workflow edge CRUD operations
        specs: Declarative workflow YAML validation, planning, apply, and export
    """

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.runs = AsyncWorkflowRuns(client=client)
        self.reviews = AsyncWorkflowReviews(client=client)
        self.artifacts = AsyncWorkflowArtifacts(client=client)
        self.blocks = AsyncWorkflowBlocks(client=client)
        self.edges = AsyncWorkflowEdges(client=client)
        self.tests = AsyncWorkflowTests(client=client)
        self.experiments = AsyncWorkflowExperiments(client=client)
        self.specs = AsyncWorkflowSpecs(client=client)

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
        email_trigger: Dict[str, List[str]] | None = None,
    ) -> Workflow:
        """Update a workflow's metadata."""
        request = self.prepare_update(
            workflow_id, name=name, description=description,
            email_trigger=email_trigger,
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
        """Get a workflow with all its entities (blocks and edges)."""
        request = self.prepare_get_entities(workflow_id)
        response = await self._client._prepared_request(request)
        return WorkflowWithEntities.model_validate(response)

    async def get_resolved_schemas(self, workflow_id: str) -> WorkflowResolvedSchemasResponse:
        """Get graph-derived schemas for all current-draft blocks in a workflow."""
        request = self.prepare_get_resolved_schemas(workflow_id)
        response = await self._client._prepared_request(request)
        return WorkflowResolvedSchemasResponse.model_validate(response)

    async def list_snapshots(
        self,
        workflow_id: str,
        limit: int | None = None,
    ) -> PaginatedList[WorkflowSnapshot]:
        """List published snapshots for a workflow (newest first)."""
        request = self.prepare_list_snapshots(workflow_id, limit=limit)
        response = await self._client._prepared_request(request)
        result = PaginatedList[WorkflowSnapshot](**response)
        result.data = [WorkflowSnapshot.model_validate(item) for item in result.data]
        return result

    async def diagnose(
        self,
        workflow_id: str,
        re_propagate: bool = True,
    ) -> WorkflowDiagnosisResponse:
        """Diagnose the workflow's draft graph for structural issues."""
        entities = await self.get_entities(workflow_id)
        blocks = [
            {
                "id": block.id,
                "type": block.type,
                "label": block.label,
                "config": block.config,
                "position": {"x": block.position_x, "y": block.position_y},
                "width": block.width,
                "height": block.height,
                "parent_id": block.parent_id,
            }
            for block in entities.blocks
        ]
        edges = [
            {
                "id": edge.id,
                "source": edge.source_block,
                "target": edge.target_block,
                "source_handle": edge.source_handle,
                "target_handle": edge.target_handle,
            }
            for edge in entities.edges
        ]
        request = self.prepare_diagnose(workflow_id, blocks, edges, re_propagate=re_propagate)
        response = await self._client._prepared_request(request)
        return WorkflowDiagnosisResponse.model_validate(response)
