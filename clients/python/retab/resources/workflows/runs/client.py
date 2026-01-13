from io import IOBase
from pathlib import Path
from typing import Any, Dict

import PIL.Image
from pydantic import HttpUrl

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....utils.mime import MIMEData, prepare_mime_document
from ....types.standards import PreparedRequest
from ....types.workflows import WorkflowRun


# Type alias for document inputs
DocumentInput = Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl


class WorkflowRunsMixin:
    """Mixin providing shared methods for workflow run operations."""

    def prepare_create(
        self,
        workflow_id: str,
        documents: Dict[str, DocumentInput],
    ) -> PreparedRequest:
        """Prepare a request to run a workflow with input documents.

        Args:
            workflow_id: The ID of the workflow to run
            documents: Mapping of start node IDs to their input documents.
                       Each document can be a file path, bytes, file-like object,
                       MIMEData, PIL Image, or HttpUrl.

        Returns:
            PreparedRequest: The prepared request

        Example:
            >>> client.workflows.runs.create(
            ...     workflow_id="wf_abc123",
            ...     documents={
            ...         "start-node-1": Path("invoice.pdf"),
            ...         "start-node-2": Path("receipt.pdf"),
            ...     }
            ... )
        """
        # Convert each document to MIMEData and then to the format expected by the backend
        documents_payload: Dict[str, Dict[str, Any]] = {}
        for node_id, document in documents.items():
            mime_data = prepare_mime_document(document)
            documents_payload[node_id] = {
                "filename": mime_data.filename,
                "content": mime_data.content,
                "mime_type": mime_data.mime_type,
            }

        data = {"documents": documents_payload}
        return PreparedRequest(method="POST", url=f"/v1/workflows/{workflow_id}/run", data=data)

    def prepare_get(self, run_id: str) -> PreparedRequest:
        """Prepare a request to get a workflow run by ID.

        Args:
            run_id: The ID of the workflow run to retrieve

        Returns:
            PreparedRequest: The prepared request
        """
        return PreparedRequest(method="GET", url=f"/v1/workflows/runs/{run_id}")


class WorkflowRuns(SyncAPIResource, WorkflowRunsMixin):
    """Workflow Runs API wrapper for synchronous operations."""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    def create(
        self,
        workflow_id: str,
        documents: Dict[str, DocumentInput],
    ) -> WorkflowRun:
        """Run a workflow with the provided input documents.

        This creates a workflow run and starts execution in the background.
        The returned WorkflowRun will have status "running" - use get()
        to check for updates on the run status.

        Args:
            workflow_id: The ID of the workflow to run
            documents: Mapping of start node IDs to their input documents.
                       Each document can be a file path, bytes, file-like object,
                       MIMEData, PIL Image, or HttpUrl.

        Returns:
            WorkflowRun: The created workflow run with status "running"

        Raises:
            HTTPException: If the request fails (e.g., workflow not found,
                          missing input documents for start nodes)

        Example:
            >>> run = client.workflows.runs.create(
            ...     workflow_id="wf_abc123",
            ...     documents={
            ...         "start-node-1": Path("invoice.pdf"),
            ...         "start-node-2": Path("receipt.pdf"),
            ...     }
            ... )
            >>> print(f"Run started: {run.id}, status: {run.status}")
        """
        request = self.prepare_create(workflow_id=workflow_id, documents=documents)
        response = self._client._prepared_request(request)
        return WorkflowRun.model_validate(response)

    def get(self, run_id: str) -> WorkflowRun:
        """Get a workflow run by ID.

        Args:
            run_id: The ID of the workflow run to retrieve

        Returns:
            WorkflowRun: The workflow run

        Raises:
            HTTPException: If the request fails (e.g., run not found)
        """
        request = self.prepare_get(run_id)
        response = self._client._prepared_request(request)
        return WorkflowRun.model_validate(response)


class AsyncWorkflowRuns(AsyncAPIResource, WorkflowRunsMixin):
    """Workflow Runs API wrapper for asynchronous operations."""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    async def create(
        self,
        workflow_id: str,
        documents: Dict[str, DocumentInput],
    ) -> WorkflowRun:
        """Run a workflow with the provided input documents.

        This creates a workflow run and starts execution in the background.
        The returned WorkflowRun will have status "running" - use get()
        to check for updates on the run status.

        Args:
            workflow_id: The ID of the workflow to run
            documents: Mapping of start node IDs to their input documents.
                       Each document can be a file path, bytes, file-like object,
                       MIMEData, PIL Image, or HttpUrl.

        Returns:
            WorkflowRun: The created workflow run with status "running"

        Raises:
            HTTPException: If the request fails (e.g., workflow not found,
                          missing input documents for start nodes)

        Example:
            >>> run = await client.workflows.runs.create(
            ...     workflow_id="wf_abc123",
            ...     documents={
            ...         "start-node-1": Path("invoice.pdf"),
            ...         "start-node-2": Path("receipt.pdf"),
            ...     }
            ... )
            >>> print(f"Run started: {run.id}, status: {run.status}")
        """
        request = self.prepare_create(workflow_id=workflow_id, documents=documents)
        response = await self._client._prepared_request(request)
        return WorkflowRun.model_validate(response)

    async def get(self, run_id: str) -> WorkflowRun:
        """Get a workflow run by ID.

        Args:
            run_id: The ID of the workflow run to retrieve

        Returns:
            WorkflowRun: The workflow run

        Raises:
            HTTPException: If the request fails (e.g., run not found)
        """
        request = self.prepare_get(run_id)
        response = await self._client._prepared_request(request)
        return WorkflowRun.model_validate(response)
