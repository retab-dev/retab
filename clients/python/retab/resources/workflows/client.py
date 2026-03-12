from typing import Any

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.pagination import PaginatedList, PaginationOrder
from ...types.standards import PreparedRequest
from ...types.workflows import Workflow
from .runs import WorkflowRuns, AsyncWorkflowRuns


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


class Workflows(SyncAPIResource, WorkflowsMixin):
    """Workflows API wrapper for synchronous operations.

    Sub-clients:
        runs: Workflow run operations (create, get, list)
    """

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.runs = WorkflowRuns(client=client)

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
        return PaginatedList(**response)


class AsyncWorkflows(AsyncAPIResource, WorkflowsMixin):
    """Workflows API wrapper for asynchronous operations.

    Sub-clients:
        runs: Workflow run operations (create, get, list)
    """

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.runs = AsyncWorkflowRuns(client=client)

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
        return PaginatedList(**response)
