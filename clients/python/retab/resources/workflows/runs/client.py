import asyncio
import inspect
import time
from datetime import date
from io import IOBase
from pathlib import Path
from typing import Any, Awaitable, Callable, Dict, List, Literal, Optional, Sequence, TypeAlias

import PIL.Image
from pydantic import HttpUrl

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....utils.mime import MIMEData, prepare_mime_document
from ....types.standards import PreparedRequest
from ....types.pagination import PaginatedList, PaginationOrder
from ....types.workflows import (
    WorkflowRun,
    TERMINAL_WORKFLOW_RUN_STATUSES,
    CancelWorkflowResponse,
    HILDecisionResource,
    SubmitHILDecisionResponse,
    ExportResponse,
    WorkflowRunStatus,
    WorkflowRunTriggerType,
)
from .steps import WorkflowSteps, AsyncWorkflowSteps


# Type alias for document inputs
DocumentInput = Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl
DateInput: TypeAlias = str | date
ProgressCallback: TypeAlias = Callable[[WorkflowRun], Any]
AsyncProgressCallback: TypeAlias = Callable[[WorkflowRun], Any | Awaitable[Any]]


def _normalize_csv_param(value: str | Sequence[str] | None) -> str | None:
    if value is None:
        return None
    if isinstance(value, str):
        return value
    return ",".join(value)


def _normalize_date_param(value: DateInput | None) -> str | None:
    if value is None:
        return None
    if isinstance(value, date):
        return value.isoformat()
    return value


class WorkflowRunsMixin:
    """Mixin providing shared methods for workflow run operations."""

    def prepare_create(
        self,
        workflow_id: str,
        documents: Optional[Dict[str, DocumentInput]] = None,
        json_inputs: Optional[Dict[str, Dict[str, Any]]] = None,
    ) -> PreparedRequest:
        """Prepare a request to run a workflow with input documents, JSON data, and/or text data.

        Args:
            workflow_id: The ID of the workflow to run
            documents: Mapping of start node IDs to their input documents.
                       Each document can be a file path, bytes, file-like object,
                       MIMEData, PIL Image, or HttpUrl.
            json_inputs: Mapping of start_json node IDs to their input JSON data.

        Returns:
            PreparedRequest: The prepared request

        Example:
            >>> client.workflows.runs.create(
            ...     workflow_id="wf_abc123",
            ...     documents={
            ...         "start-node-1": Path("invoice.pdf"),
            ...         "start-node-2": Path("receipt.pdf"),
            ...     },
            ...     json_inputs={
            ...         "json-node-1": {"key": "value"},
            ...     },            ... )
        """
        data: Dict[str, Any] = {}

        # Convert each document to MIMEData and then to the format expected by the backend
        if documents:
            documents_payload: Dict[str, Dict[str, Any]] = {}
            for node_id, document in documents.items():
                mime_data = prepare_mime_document(document)
                documents_payload[node_id] = {
                    "filename": mime_data.filename,
                    "content": mime_data.content,
                    "mime_type": mime_data.mime_type,
                }
            data["documents"] = documents_payload

        # Add JSON inputs directly
        if json_inputs:
            data["json_inputs"] = json_inputs
        return PreparedRequest(method="POST", url=f"/workflows/{workflow_id}/run", data=data)

    def prepare_get(self, run_id: str) -> PreparedRequest:
        """Prepare a request to get a workflow run by ID."""
        return PreparedRequest(method="GET", url=f"/workflows/runs/{run_id}")

    def prepare_list(
        self,
        workflow_id: str | None = None,
        status: WorkflowRunStatus | None = None,
        statuses: str | Sequence[WorkflowRunStatus] | None = None,
        exclude_status: WorkflowRunStatus | None = None,
        trigger_type: WorkflowRunTriggerType | None = None,
        trigger_types: str | Sequence[WorkflowRunTriggerType] | None = None,
        from_date: DateInput | None = None,
        to_date: DateInput | None = None,
        min_cost: float | None = None,
        max_cost: float | None = None,
        min_duration: int | None = None,
        max_duration: int | None = None,
        search: str | None = None,
        sort_by: str = "created_at",
        fields: str | Sequence[str] | None = None,
        before: str | None = None,
        after: str | None = None,
        limit: int = 20,
        order: PaginationOrder = "desc",
    ) -> PreparedRequest:
        """Prepare a request to list workflow runs with filtering and pagination."""
        params: Dict[str, Any] = {"limit": limit, "order": order, "sort_by": sort_by}
        if workflow_id is not None:
            params["workflow_id"] = workflow_id
        if status is not None:
            params["status"] = status
        normalized_statuses = _normalize_csv_param(statuses)
        if normalized_statuses is not None:
            params["statuses"] = normalized_statuses
        if exclude_status is not None:
            params["exclude_status"] = exclude_status
        if trigger_type is not None:
            params["trigger_type"] = trigger_type
        normalized_trigger_types = _normalize_csv_param(trigger_types)
        if normalized_trigger_types is not None:
            params["trigger_types"] = normalized_trigger_types
        normalized_from_date = _normalize_date_param(from_date)
        if normalized_from_date is not None:
            params["from_date"] = normalized_from_date
        normalized_to_date = _normalize_date_param(to_date)
        if normalized_to_date is not None:
            params["to_date"] = normalized_to_date
        if min_cost is not None:
            params["min_cost"] = min_cost
        if max_cost is not None:
            params["max_cost"] = max_cost
        if min_duration is not None:
            params["min_duration"] = min_duration
        if max_duration is not None:
            params["max_duration"] = max_duration
        if search is not None:
            params["search"] = search
        normalized_fields = _normalize_csv_param(fields)
        if normalized_fields is not None:
            params["fields"] = normalized_fields
        if before is not None:
            params["before"] = before
        if after is not None:
            params["after"] = after
        return PreparedRequest(method="GET", url="/workflows/runs", params=params)

    def prepare_delete(self, run_id: str) -> PreparedRequest:
        """Prepare a request to delete a workflow run."""
        return PreparedRequest(method="DELETE", url=f"/workflows/runs/{run_id}")

    def prepare_cancel(self, run_id: str, command_id: str | None = None) -> PreparedRequest:
        """Prepare a request to cancel a workflow run."""
        data = None
        if command_id is not None:
            data = {"command_id": command_id}
        return PreparedRequest(method="POST", url=f"/workflows/runs/{run_id}/cancel", data=data)

    def prepare_restart(self, run_id: str, command_id: str | None = None) -> PreparedRequest:
        """Prepare a request to restart a workflow run."""
        data = None
        if command_id is not None:
            data = {"command_id": command_id}
        return PreparedRequest(method="POST", url=f"/workflows/runs/{run_id}/restart", data=data)

    def prepare_submit_hil_decision(
        self,
        run_id: str,
        node_id: str,
        approved: bool,
        modified_data: dict | None = None,
        command_id: str | None = None,
    ) -> PreparedRequest:
        """Prepare a request to submit a HIL decision for a workflow run."""
        data: Dict[str, Any] = {"node_id": node_id, "approved": approved}
        if modified_data is not None:
            data["modified_data"] = modified_data
        if command_id is not None:
            data["command_id"] = command_id
        return PreparedRequest(
            method="POST",
            url=f"/workflows/runs/{run_id}/hil-decisions",
            data=data,
        )

    def prepare_get_hil_decision(self, run_id: str, node_id: str) -> PreparedRequest:
        """Prepare a request to get the authoritative HIL decision state for a node."""
        return PreparedRequest(
            method="GET",
            url=f"/workflows/runs/{run_id}/hil-decisions/{node_id}",
        )

class WorkflowRuns(SyncAPIResource, WorkflowRunsMixin):
    """Workflow Runs API wrapper for synchronous operations.

    Sub-clients:
        steps: Step output operations (get, list)

    Example:
        >>> from retab import Retab
        >>> client = Retab(api_key="your-api-key")
        >>>
        >>> # Run a workflow and wait for completion
        >>> run = client.workflows.runs.create(
        ...     workflow_id="wf_abc123",
        ...     documents={"start-node-1": Path("invoice.pdf")},
        ... )
        >>> run = client.workflows.runs.wait_for_completion(run.id)
        >>>
        >>> # Get outputs from a specific step
        >>> step = client.workflows.runs.steps.get(run.id, "extract-node-1")
        >>> print(step.handle_outputs)
    """

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.steps = WorkflowSteps(client=self._client)

    def create(
        self,
        workflow_id: str,
        documents: Optional[Dict[str, DocumentInput]] = None,
        json_inputs: Optional[Dict[str, Dict[str, Any]]] = None,
    ) -> WorkflowRun:
        """Run a workflow with the provided inputs.

        This creates a workflow run and starts execution in the background.
        The returned WorkflowRun will have status "running" - use get()
        to check for updates on the run status.

        Args:
            workflow_id: The ID of the workflow to run
            documents: Mapping of start node IDs to their input documents.
                       Each document can be a file path, bytes, file-like object,
                       MIMEData, PIL Image, or HttpUrl.
            json_inputs: Mapping of start_json node IDs to their input JSON data.

        Returns:
            WorkflowRun: The created workflow run with status "running"

        Example:
            >>> run = client.workflows.runs.create(
            ...     workflow_id="wf_abc123",
            ...     documents={
            ...         "start-node-1": Path("invoice.pdf"),
            ...     },
            ... )
            >>> print(f"Run started: {run.id}, status: {run.status}")
        """
        request = self.prepare_create(
            workflow_id=workflow_id,
            documents=documents,
            json_inputs=json_inputs,
        )
        response = self._client._prepared_request(request)
        return WorkflowRun.model_validate(response)

    def get(self, run_id: str) -> WorkflowRun:
        """Get a workflow run by ID.

        Args:
            run_id: The ID of the workflow run to retrieve

        Returns:
            WorkflowRun: The workflow run
        """
        request = self.prepare_get(run_id)
        response = self._client._prepared_request(request)
        return WorkflowRun.model_validate(response)

    def list(
        self,
        workflow_id: str | None = None,
        status: WorkflowRunStatus | None = None,
        statuses: str | Sequence[WorkflowRunStatus] | None = None,
        exclude_status: WorkflowRunStatus | None = None,
        trigger_type: WorkflowRunTriggerType | None = None,
        trigger_types: str | Sequence[WorkflowRunTriggerType] | None = None,
        from_date: DateInput | None = None,
        to_date: DateInput | None = None,
        min_cost: float | None = None,
        max_cost: float | None = None,
        min_duration: int | None = None,
        max_duration: int | None = None,
        search: str | None = None,
        sort_by: str = "created_at",
        fields: str | Sequence[str] | None = None,
        before: str | None = None,
        after: str | None = None,
        limit: int = 20,
        order: PaginationOrder = "desc",
    ) -> PaginatedList:
        """List workflow runs with pagination and filtering.

        Args:
            workflow_id: Filter by workflow ID
            status: Filter by single run status
            statuses: Filter by multiple statuses (comma-separated, e.g. "completed,error")
            exclude_status: Exclude runs with this status
            trigger_type: Filter by single trigger type
            trigger_types: Filter by multiple trigger types (comma-separated)
            from_date: Filter runs created on or after this date (YYYY-MM-DD)
            to_date: Filter runs created on or before this date (YYYY-MM-DD)
            min_cost: Filter runs with cost >= this value
            max_cost: Filter runs with cost <= this value
            min_duration: Filter runs with duration >= this value (milliseconds)
            max_duration: Filter runs with duration <= this value (milliseconds)
            search: Search by run ID (partial match)
            sort_by: Field to sort by (default: "created_at")
            fields: Comma-separated list of fields to return
            before: Pagination cursor (first ID from current page)
            after: Pagination cursor (last ID from previous page)
            limit: Items per page (1-100, default 20)
            order: Sort order ("asc" or "desc")

        Returns:
            PaginatedList: Paginated list of workflow runs
        """
        request = self.prepare_list(
            workflow_id=workflow_id,
            status=status,
            statuses=statuses,
            exclude_status=exclude_status,
            trigger_type=trigger_type,
            trigger_types=trigger_types,
            from_date=from_date,
            to_date=to_date,
            min_cost=min_cost,
            max_cost=max_cost,
            min_duration=min_duration,
            max_duration=max_duration,
            search=search,
            sort_by=sort_by,
            fields=fields,
            before=before,
            after=after,
            limit=limit,
            order=order,
        )
        response = self._client._prepared_request(request)
        result = PaginatedList(**response)
        result.data = [WorkflowRun.model_validate(item) if isinstance(item, dict) else item for item in result.data]
        return result

    def delete(self, run_id: str) -> None:
        """Delete a workflow run and its associated step data.

        Args:
            run_id: The ID of the workflow run to delete
        """
        request = self.prepare_delete(run_id)
        self._client._prepared_request(request)

    def cancel(self, run_id: str, command_id: str | None = None) -> CancelWorkflowResponse:
        """Cancel a running or pending workflow run.

        Args:
            run_id: The ID of the workflow run to cancel
            command_id: Optional idempotency key for deduplicating cancel commands

        Returns:
            CancelWorkflowResponse: The updated run and cancellation status
        """
        request = self.prepare_cancel(run_id, command_id=command_id)
        response = self._client._prepared_request(request)
        return CancelWorkflowResponse.model_validate(response)

    def restart(self, run_id: str, command_id: str | None = None) -> WorkflowRun:
        """Restart a completed or failed workflow run with the same inputs.

        Args:
            run_id: The ID of the workflow run to restart
            command_id: Optional idempotency key for deduplicating restart commands

        Returns:
            WorkflowRun: The newly created run with status "running"
        """
        request = self.prepare_restart(run_id, command_id=command_id)
        response = self._client._prepared_request(request)
        return WorkflowRun.model_validate(response)

    def submit_hil_decision(
        self,
        run_id: str,
        node_id: str,
        approved: bool,
        modified_data: dict | None = None,
        command_id: str | None = None,
    ) -> SubmitHILDecisionResponse:
        """Submit a human-in-the-loop (HIL) decision for a workflow run.

        Args:
            run_id: The ID of the workflow run
            node_id: The ID of the HIL node being approved/rejected
            approved: Whether the human approved the data
            modified_data: Optional modified data if the human made changes
            command_id: Optional idempotency key for deduplicating decision submissions

        Returns:
            SubmitHILDecisionResponse: The decision submission result
        """
        request = self.prepare_submit_hil_decision(
            run_id,
            node_id=node_id,
            approved=approved,
            modified_data=modified_data,
            command_id=command_id,
        )
        response = self._client._prepared_request(request)
        return SubmitHILDecisionResponse.model_validate(response)

    def get_hil_decision(self, run_id: str, node_id: str) -> HILDecisionResource:
        """Get the authoritative HIL decision state for a workflow run node."""
        request = self.prepare_get_hil_decision(run_id, node_id)
        response = self._client._prepared_request(request)
        return HILDecisionResource.model_validate(response)

    def wait_for_completion(
        self,
        run_id: str,

        poll_interval_seconds: float = 2.0,
        timeout_seconds: float = 600.0,
        on_status: ProgressCallback | None = None,
    ) -> WorkflowRun:
        """Poll a workflow run until it reaches a terminal state.

        Stops when the run completes, fails, is cancelled, or pauses for human review.

        Args:
            run_id: The ID of the workflow run to wait for
            poll_interval_seconds: Seconds between polls (default 2.0)
            timeout_seconds: Maximum time to wait (default 600.0)
            on_status: Optional callback invoked with the ``WorkflowRun`` on each poll.
                Useful for logging progress or updating a progress bar.

        Returns:
            WorkflowRun: The completed, cancelled, failed, or waiting-for-human workflow run

        Raises:
            TimeoutError: If the run doesn't complete within timeout_seconds
            ValueError: If poll_interval_seconds or timeout_seconds are <= 0
        """
        if poll_interval_seconds <= 0:
            raise ValueError("poll_interval_seconds must be > 0")
        if timeout_seconds <= 0:
            raise ValueError("timeout_seconds must be > 0")

        started_at = time.monotonic()
        deadline = started_at + timeout_seconds
        while True:
            run = self.get(run_id)
            if on_status is not None:
                on_status(run)
            if run.status in TERMINAL_WORKFLOW_RUN_STATUSES or run.status == "waiting_for_human":
                return run

            now = time.monotonic()
            if now >= deadline:
                raise TimeoutError(
                    f"Timed out waiting for workflow run {run_id} after {timeout_seconds}s"
                )
            sleep_for = min(poll_interval_seconds, max(deadline - now, 0.0))
            time.sleep(sleep_for)

    def export(
        self,
        workflow_id: str,
        node_id: str,
        export_source: Literal["outputs", "inputs"] = "outputs",

        selected_run_ids: List[str] | None = None,
        status: str | None = None,
        exclude_status: str | None = None,
        from_date: str | None = None,
        to_date: str | None = None,
        trigger_types: List[str] | None = None,
        preferred_columns: List[str] | None = None,
    ) -> ExportResponse:
        """Export run results as structured CSV data.

        Args:
            workflow_id: The workflow ID
            node_id: The node ID to export outputs from
            export_source: Whether to export "outputs" or "inputs" (default "outputs")
            selected_run_ids: Limit export to specific run IDs (optional)
            status: Filter by status (optional)
            exclude_status: Exclude specific status (optional)
            from_date: Filter from date (YYYY-MM-DD, optional)
            to_date: Filter to date (YYYY-MM-DD, optional)
            trigger_types: Filter by trigger types (optional)
            preferred_columns: Column ordering (optional)

        Returns:
            Dict with csv_data (str), rows (int), columns (int)
        """
        data: Dict[str, Any] = {
            "workflow_id": workflow_id,
            "node_id": node_id,
            "export_source": export_source,
            "preferred_columns": preferred_columns or [],
        }
        if selected_run_ids is not None:
            data["selected_run_ids"] = selected_run_ids
        if status is not None:
            data["status"] = status
        if exclude_status is not None:
            data["exclude_status"] = exclude_status
        if from_date is not None:
            data["from_date"] = from_date
        if to_date is not None:
            data["to_date"] = to_date
        if trigger_types is not None:
            data["trigger_types"] = trigger_types
        request = PreparedRequest(method="POST", url="/workflows/runs/export_payload", data=data)
        response = self._client._prepared_request(request)
        return ExportResponse.model_validate(response)


class AsyncWorkflowRuns(AsyncAPIResource, WorkflowRunsMixin):
    """Workflow Runs API wrapper for asynchronous operations.

    Sub-clients:
        steps: Step output operations (get, list)

    Example:
        >>> from retab import AsyncRetab
        >>> client = AsyncRetab(api_key="your-api-key")
        >>>
        >>> # Run a workflow and wait for completion
        >>> run = await client.workflows.runs.create(
        ...     workflow_id="wf_abc123",
        ...     documents={"start-node-1": Path("invoice.pdf")},
        ... )
        >>> run = await client.workflows.runs.wait_for_completion(run.id)
        >>>
        >>> # Get outputs from a specific step
        >>> step = await client.workflows.runs.steps.get(run.id, "extract-node-1")
        >>> print(step.handle_outputs)
    """

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.steps = AsyncWorkflowSteps(client=self._client)

    async def create(
        self,
        workflow_id: str,
        documents: Optional[Dict[str, DocumentInput]] = None,
        json_inputs: Optional[Dict[str, Dict[str, Any]]] = None,
    ) -> WorkflowRun:
        """Run a workflow with the provided inputs.

        This creates a workflow run and starts execution in the background.
        The returned WorkflowRun will have status "running" - use get()
        to check for updates on the run status.

        Args:
            workflow_id: The ID of the workflow to run
            documents: Mapping of start node IDs to their input documents.
                       Each document can be a file path, bytes, file-like object,
                       MIMEData, PIL Image, or HttpUrl.
            json_inputs: Mapping of start_json node IDs to their input JSON data.

        Returns:
            WorkflowRun: The created workflow run with status "running"

        Example:
            >>> run = await client.workflows.runs.create(
            ...     workflow_id="wf_abc123",
            ...     documents={"start-node-1": Path("invoice.pdf")},
            ... )
            >>> print(f"Run started: {run.id}, status: {run.status}")
        """
        request = self.prepare_create(
            workflow_id=workflow_id,
            documents=documents,
            json_inputs=json_inputs,
        )
        response = await self._client._prepared_request(request)
        return WorkflowRun.model_validate(response)

    async def get(self, run_id: str) -> WorkflowRun:
        """Get a workflow run by ID.

        Args:
            run_id: The ID of the workflow run to retrieve

        Returns:
            WorkflowRun: The workflow run
        """
        request = self.prepare_get(run_id)
        response = await self._client._prepared_request(request)
        return WorkflowRun.model_validate(response)

    async def list(
        self,
        workflow_id: str | None = None,
        status: WorkflowRunStatus | None = None,
        statuses: str | Sequence[WorkflowRunStatus] | None = None,
        exclude_status: WorkflowRunStatus | None = None,
        trigger_type: WorkflowRunTriggerType | None = None,
        trigger_types: str | Sequence[WorkflowRunTriggerType] | None = None,
        from_date: DateInput | None = None,
        to_date: DateInput | None = None,
        min_cost: float | None = None,
        max_cost: float | None = None,
        min_duration: int | None = None,
        max_duration: int | None = None,
        search: str | None = None,
        sort_by: str = "created_at",
        fields: str | Sequence[str] | None = None,
        before: str | None = None,
        after: str | None = None,
        limit: int = 20,
        order: PaginationOrder = "desc",
    ) -> PaginatedList:
        """List workflow runs with pagination and filtering.

        Args:
            workflow_id: Filter by workflow ID
            status: Filter by single run status
            statuses: Filter by multiple statuses (comma-separated, e.g. "completed,error")
            exclude_status: Exclude runs with this status
            trigger_type: Filter by single trigger type
            trigger_types: Filter by multiple trigger types (comma-separated)
            from_date: Filter runs created on or after this date (YYYY-MM-DD)
            to_date: Filter runs created on or before this date (YYYY-MM-DD)
            min_cost: Filter runs with cost >= this value
            max_cost: Filter runs with cost <= this value
            min_duration: Filter runs with duration >= this value (milliseconds)
            max_duration: Filter runs with duration <= this value (milliseconds)
            search: Search by run ID (partial match)
            sort_by: Field to sort by (default: "created_at")
            fields: Comma-separated list of fields to return
            before: Pagination cursor (first ID from current page)
            after: Pagination cursor (last ID from previous page)
            limit: Items per page (1-100, default 20)
            order: Sort order ("asc" or "desc")

        Returns:
            PaginatedList: Paginated list of workflow runs
        """
        request = self.prepare_list(
            workflow_id=workflow_id,
            status=status,
            statuses=statuses,
            exclude_status=exclude_status,
            trigger_type=trigger_type,
            trigger_types=trigger_types,
            from_date=from_date,
            to_date=to_date,
            min_cost=min_cost,
            max_cost=max_cost,
            min_duration=min_duration,
            max_duration=max_duration,
            search=search,
            sort_by=sort_by,
            fields=fields,
            before=before,
            after=after,
            limit=limit,
            order=order,
        )
        response = await self._client._prepared_request(request)
        result = PaginatedList(**response)
        result.data = [WorkflowRun.model_validate(item) if isinstance(item, dict) else item for item in result.data]
        return result

    async def delete(self, run_id: str) -> None:
        """Delete a workflow run and its associated step data.

        Args:
            run_id: The ID of the workflow run to delete
        """
        request = self.prepare_delete(run_id)
        await self._client._prepared_request(request)

    async def cancel(self, run_id: str, command_id: str | None = None) -> CancelWorkflowResponse:
        """Cancel a running or pending workflow run.

        Args:
            run_id: The ID of the workflow run to cancel
            command_id: Optional idempotency key for deduplicating cancel commands

        Returns:
            CancelWorkflowResponse: The updated run and cancellation status
        """
        request = self.prepare_cancel(run_id, command_id=command_id)
        response = await self._client._prepared_request(request)
        return CancelWorkflowResponse.model_validate(response)

    async def restart(self, run_id: str, command_id: str | None = None) -> WorkflowRun:
        """Restart a completed or failed workflow run with the same inputs.

        Args:
            run_id: The ID of the workflow run to restart
            command_id: Optional idempotency key for deduplicating restart commands

        Returns:
            WorkflowRun: The newly created run with status "running"
        """
        request = self.prepare_restart(run_id, command_id=command_id)
        response = await self._client._prepared_request(request)
        return WorkflowRun.model_validate(response)

    async def submit_hil_decision(
        self,
        run_id: str,
        node_id: str,
        approved: bool,
        modified_data: dict | None = None,
        command_id: str | None = None,
    ) -> SubmitHILDecisionResponse:
        """Submit a human-in-the-loop (HIL) decision for a workflow run.

        Args:
            run_id: The ID of the workflow run
            node_id: The ID of the HIL node being approved/rejected
            approved: Whether the human approved the data
            modified_data: Optional modified data if the human made changes
            command_id: Optional idempotency key for deduplicating decision submissions

        Returns:
            SubmitHILDecisionResponse: The decision submission result
        """
        request = self.prepare_submit_hil_decision(
            run_id,
            node_id=node_id,
            approved=approved,
            modified_data=modified_data,
            command_id=command_id,
        )
        response = await self._client._prepared_request(request)
        return SubmitHILDecisionResponse.model_validate(response)

    async def get_hil_decision(self, run_id: str, node_id: str) -> HILDecisionResource:
        """Get the authoritative HIL decision state for a workflow run node."""
        request = self.prepare_get_hil_decision(run_id, node_id)
        response = await self._client._prepared_request(request)
        return HILDecisionResource.model_validate(response)

    async def wait_for_completion(
        self,
        run_id: str,

        poll_interval_seconds: float = 2.0,
        timeout_seconds: float = 600.0,
        on_status: AsyncProgressCallback | None = None,
    ) -> WorkflowRun:
        """Poll a workflow run until it reaches a terminal state.

        Stops when the run completes, fails, is cancelled, or pauses for human review.

        Args:
            run_id: The ID of the workflow run to wait for
            poll_interval_seconds: Seconds between polls (default 2.0)
            timeout_seconds: Maximum time to wait (default 600.0)
            on_status: Optional callback invoked with the ``WorkflowRun`` on each poll.

        Returns:
            WorkflowRun: The completed, cancelled, failed, or waiting-for-human workflow run

        Raises:
            TimeoutError: If the run doesn't complete within timeout_seconds
            ValueError: If poll_interval_seconds or timeout_seconds are <= 0
        """
        if poll_interval_seconds <= 0:
            raise ValueError("poll_interval_seconds must be > 0")
        if timeout_seconds <= 0:
            raise ValueError("timeout_seconds must be > 0")

        started_at = time.monotonic()
        deadline = started_at + timeout_seconds
        while True:
            run = await self.get(run_id)
            if on_status is not None:
                callback_result = on_status(run)
                if inspect.isawaitable(callback_result):
                    await callback_result
            if run.status in TERMINAL_WORKFLOW_RUN_STATUSES or run.status == "waiting_for_human":
                return run

            now = time.monotonic()
            if now >= deadline:
                raise TimeoutError(
                    f"Timed out waiting for workflow run {run_id} after {timeout_seconds}s"
                )
            sleep_for = min(poll_interval_seconds, max(deadline - now, 0.0))
            await asyncio.sleep(sleep_for)

    async def export(
        self,
        workflow_id: str,
        node_id: str,
        export_source: Literal["outputs", "inputs"] = "outputs",

        selected_run_ids: List[str] | None = None,
        status: str | None = None,
        exclude_status: str | None = None,
        from_date: str | None = None,
        to_date: str | None = None,
        trigger_types: List[str] | None = None,
        preferred_columns: List[str] | None = None,
    ) -> ExportResponse:
        """Export run results as structured CSV data."""
        data: Dict[str, Any] = {
            "workflow_id": workflow_id,
            "node_id": node_id,
            "export_source": export_source,
            "preferred_columns": preferred_columns or [],
        }
        if selected_run_ids is not None:
            data["selected_run_ids"] = selected_run_ids
        if status is not None:
            data["status"] = status
        if exclude_status is not None:
            data["exclude_status"] = exclude_status
        if from_date is not None:
            data["from_date"] = from_date
        if to_date is not None:
            data["to_date"] = to_date
        if trigger_types is not None:
            data["trigger_types"] = trigger_types
        request = PreparedRequest(method="POST", url="/workflows/runs/export_payload", data=data)
        response = await self._client._prepared_request(request)
        return ExportResponse.model_validate(response)
