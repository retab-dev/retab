"""
Jobs API Resource

Provides synchronous and asynchronous clients for the Jobs API.
"""

import asyncio
import json
import time
from typing import Any, Sequence

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.jobs import Job, JobListOrder, JobListResponse, JobListSource, JobStatus, SupportedEndpoint
from ...types.standards import PreparedRequest

TERMINAL_JOB_STATUSES: tuple[JobStatus, ...] = ("completed", "failed", "cancelled", "expired")


class BaseJobsMixin:
    """Shared methods for preparing Jobs API requests."""

    def _prepare_create(
        self,
        endpoint: SupportedEndpoint,
        request: dict[str, Any],
        metadata: dict[str, str] | None = None,
    ) -> PreparedRequest:
        data = {
            "endpoint": endpoint,
            "request": request,
        }
        if metadata is not None:
            data["metadata"] = metadata
        return PreparedRequest(method="POST", url="/v1/jobs", data=data)

    def _prepare_retrieve(self, job_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/jobs/{job_id}")

    def _prepare_retrieve_with_options(
        self,
        job_id: str,
        include_request: bool | None = None,
        include_response: bool | None = None,
    ) -> PreparedRequest:
        params: dict[str, Any] = {}
        if include_request is not None:
            params["include_request"] = include_request
        if include_response is not None:
            params["include_response"] = include_response
        return PreparedRequest(method="GET", url=f"/v1/jobs/{job_id}", params=params or None)

    def _prepare_cancel(self, job_id: str) -> PreparedRequest:
        return PreparedRequest(method="POST", url=f"/v1/jobs/{job_id}/cancel")

    def _prepare_retry(self, job_id: str) -> PreparedRequest:
        return PreparedRequest(method="POST", url=f"/v1/jobs/{job_id}/retry")

    def _prepare_list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 20,
        order: JobListOrder | None = "desc",
        id: str | None = None,
        status: JobStatus | None = None,
        endpoint: SupportedEndpoint | None = None,
        source: JobListSource | None = None,
        project_id: str | None = None,
        workflow_id: str | None = None,
        workflow_node_id: str | None = None,
        model: str | None = None,
        filename_regex: str | None = None,
        filename_contains: str | None = None,
        document_type: Sequence[str] | None = None,
        from_date: str | None = None,
        to_date: str | None = None,
        metadata: dict[str, str] | None = None,
        include_request: bool | None = None,
        include_response: bool | None = None,
    ) -> PreparedRequest:
        params: dict[str, Any] = {"limit": limit}
        if before is not None:
            params["before"] = before
        if after is not None:
            params["after"] = after
        if order is not None:
            params["order"] = order
        if id is not None:
            params["id"] = id
        if status is not None:
            params["status"] = status
        if endpoint is not None:
            params["endpoint"] = endpoint
        if source is not None:
            params["source"] = source
        if project_id is not None:
            params["project_id"] = project_id
        if workflow_id is not None:
            params["workflow_id"] = workflow_id
        if workflow_node_id is not None:
            params["workflow_node_id"] = workflow_node_id
        if model is not None:
            params["model"] = model
        if filename_regex is not None:
            params["filename_regex"] = filename_regex
        if filename_contains is not None:
            params["filename_contains"] = filename_contains
        if document_type:
            params["document_type"] = list(document_type)
        if from_date is not None:
            params["from_date"] = from_date
        if to_date is not None:
            params["to_date"] = to_date
        if metadata is not None:
            params["metadata"] = json.dumps(metadata)
        if include_request is not None:
            params["include_request"] = include_request
        if include_response is not None:
            params["include_response"] = include_response
        return PreparedRequest(method="GET", url="/v1/jobs", params=params)


class Jobs(SyncAPIResource, BaseJobsMixin):
    """
    Synchronous Jobs API client.

    The Jobs API allows you to submit long-running extract or parse operations
    asynchronously and poll for their results.

    Example:
        >>> from retab import Retab
        >>> client = Retab(api_key="your-api-key")
        >>>
        >>> # Create an async extraction job
        >>> job = client.jobs.create(
        ...     endpoint="/v1/documents/extract",
        ...     request={
        ...         "document": {"content": "...", "mime_type": "application/pdf"},
        ...         "json_schema": {"type": "object", ...},
        ...         "model": "retab-small",
        ...     }
        ... )
        >>>
        >>> # Poll for completion
        >>> job = client.jobs.wait_for_completion(job.id)
        >>> if job.status == "completed" and job.response:
        ...     print(job.response.body)
    """

    def create(
        self,
        endpoint: SupportedEndpoint,
        request: dict[str, Any],
        metadata: dict[str, str] | None = None,
    ) -> Job:
        """
        Create a new asynchronous job.

        Args:
            endpoint: The API endpoint to call ("/v1/documents/extract" or "/v1/documents/parse")
            request: The full request body for the target endpoint
            metadata: Optional metadata (max 16 pairs; keys ≤64 chars, values ≤512 chars)

        Returns:
            Job: The created job with status "queued"
        """
        prepared = self._prepare_create(endpoint, request, metadata)
        response = self._client._prepared_request(prepared)
        return Job.model_validate(response)

    def retrieve(
        self,
        job_id: str,
        include_request: bool | None = False,
        include_response: bool | None = False,
    ) -> Job:
        """
        Retrieve a job by ID.

        Args:
            job_id: The job ID to retrieve
            include_request: Whether to include the original request payload (defaults to False when omitted)
            include_response: Whether to include response payload/documents (defaults to False when omitted)

        Returns:
            Job: The job with current status and result (if completed)
        """
        prepared = self._prepare_retrieve_with_options(job_id, include_request, include_response)
        response = self._client._prepared_request(prepared)
        return Job.model_validate(response)

    def retrieve_full(self, job_id: str) -> Job:
        """Retrieve a job by ID including full request and response payloads."""
        return self.retrieve(
            job_id,
            include_request=True,
            include_response=True,
        )

    def wait_for_completion(
        self,
        job_id: str,
        *,
        poll_interval_seconds: float = 2.0,
        timeout_seconds: float = 600.0,
        include_request: bool = False,
        include_response: bool = True,
    ) -> Job:
        """
        Poll a job until it reaches a terminal state, then fetch the final payload once.

        Polling remains lightweight by default and only the final retrieve can include
        request/response payloads.
        """
        if poll_interval_seconds <= 0:
            raise ValueError("poll_interval_seconds must be > 0")
        if timeout_seconds <= 0:
            raise ValueError("timeout_seconds must be > 0")

        started_at = time.monotonic()
        deadline = started_at + timeout_seconds
        while True:
            job = self.retrieve(
                job_id,
                include_request=False,
                include_response=False,
            )
            if job.status in TERMINAL_JOB_STATUSES:
                if include_request or include_response:
                    return self.retrieve(
                        job_id,
                        include_request=include_request,
                        include_response=include_response,
                    )
                return job

            now = time.monotonic()
            if now >= deadline:
                raise TimeoutError(f"Timed out waiting for job {job_id} completion after {timeout_seconds} seconds")
            sleep_for = min(poll_interval_seconds, max(deadline - now, 0.0))
            time.sleep(sleep_for)

    def cancel(self, job_id: str) -> Job:
        """
        Cancel a queued or in-progress job.

        Args:
            job_id: The job ID to cancel

        Returns:
            Job: The updated job with status "cancelled"
        """
        prepared = self._prepare_cancel(job_id)
        response = self._client._prepared_request(prepared)
        return Job.model_validate(response)

    def retry(self, job_id: str) -> Job:
        """
        Retry a failed, cancelled, or expired job.

        Args:
            job_id: The job ID to retry

        Returns:
            Job: The updated job with status "queued"
        """
        prepared = self._prepare_retry(job_id)
        response = self._client._prepared_request(prepared)
        return Job.model_validate(response)

    def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 20,
        order: JobListOrder | None = "desc",
        id: str | None = None,
        status: JobStatus | None = None,
        endpoint: SupportedEndpoint | None = None,
        source: JobListSource | None = None,
        project_id: str | None = None,
        workflow_id: str | None = None,
        workflow_node_id: str | None = None,
        model: str | None = None,
        filename_regex: str | None = None,
        filename_contains: str | None = None,
        document_type: Sequence[str] | None = None,
        from_date: str | None = None,
        to_date: str | None = None,
        metadata: dict[str, str] | None = None,
        include_request: bool | None = None,
        include_response: bool | None = None,
    ) -> JobListResponse:
        """
        List jobs with pagination and optional filtering.

        Args:
            before: Pagination cursor (first ID from current page)
            after: Pagination cursor (last ID from previous page)
            limit: Number of jobs to return (1-100, default 20)
            order: Sort order by created_at ("asc" or "desc")
            id: Filter by job ID
            status: Filter by job status
            endpoint: Filter by endpoint
            source: High-level source filter (api/project/workflow)
            project_id: Filter by request.project_id
            workflow_id: Filter by metadata.workflow_id
            workflow_node_id: Filter by metadata.workflow_node_id or metadata.node_id
            model: Filter by request.model
            filename_regex: Regex/plain-text filename filter
            filename_contains: Substring filename filter
            document_type: Document type filters (can pass multiple)
            from_date: Filter jobs created on or after date (YYYY-MM-DD)
            to_date: Filter jobs created on or before date (YYYY-MM-DD)
            metadata: Exact metadata key/value filters
            include_request: Whether to include full request payloads
            include_response: Whether to include full response payloads

        Returns:
            JobListResponse: List of jobs with pagination info
        """
        prepared = self._prepare_list(
            before=before,
            after=after,
            limit=limit,
            order=order,
            id=id,
            status=status,
            endpoint=endpoint,
            source=source,
            project_id=project_id,
            workflow_id=workflow_id,
            workflow_node_id=workflow_node_id,
            model=model,
            filename_regex=filename_regex,
            filename_contains=filename_contains,
            document_type=document_type,
            from_date=from_date,
            to_date=to_date,
            metadata=metadata,
            include_request=include_request,
            include_response=include_response,
        )
        response = self._client._prepared_request(prepared)
        return JobListResponse.model_validate(response)


class AsyncJobs(AsyncAPIResource, BaseJobsMixin):
    """
    Asynchronous Jobs API client.

    The Jobs API allows you to submit long-running extract or parse operations
    asynchronously and poll for their results.

    Example:
        >>> from retab import AsyncRetab
        >>> client = AsyncRetab(api_key="your-api-key")
        >>>
        >>> # Create an async extraction job
        >>> job = await client.jobs.create(
        ...     endpoint="/v1/documents/extract",
        ...     request={
        ...         "document": {"content": "...", "mime_type": "application/pdf"},
        ...         "json_schema": {"type": "object", ...},
        ...         "model": "retab-small",
        ...     }
        ... )
        >>>
        >>> # Poll for completion
        >>> job = await client.jobs.wait_for_completion(job.id)
        >>> if job.status == "completed" and job.response:
        ...     print(job.response.body)
    """

    async def create(
        self,
        endpoint: SupportedEndpoint,
        request: dict[str, Any],
        metadata: dict[str, str] | None = None,
    ) -> Job:
        """
        Create a new asynchronous job.

        Args:
            endpoint: The API endpoint to call ("/v1/documents/extract" or "/v1/documents/parse")
            request: The full request body for the target endpoint
            metadata: Optional metadata (max 16 pairs; keys ≤64 chars, values ≤512 chars)

        Returns:
            Job: The created job with status "queued"
        """
        prepared = self._prepare_create(endpoint, request, metadata)
        response = await self._client._prepared_request(prepared)
        return Job.model_validate(response)

    async def retrieve(
        self,
        job_id: str,
        include_request: bool | None = False,
        include_response: bool | None = False,
    ) -> Job:
        """
        Retrieve a job by ID.

        Args:
            job_id: The job ID to retrieve
            include_request: Whether to include the original request payload (defaults to False when omitted)
            include_response: Whether to include response payload/documents (defaults to False when omitted)

        Returns:
            Job: The job with current status and result (if completed)
        """
        prepared = self._prepare_retrieve_with_options(job_id, include_request, include_response)
        response = await self._client._prepared_request(prepared)
        return Job.model_validate(response)

    async def retrieve_full(self, job_id: str) -> Job:
        """Retrieve a job by ID including full request and response payloads."""
        return await self.retrieve(
            job_id,
            include_request=True,
            include_response=True,
        )

    async def wait_for_completion(
        self,
        job_id: str,
        *,
        poll_interval_seconds: float = 2.0,
        timeout_seconds: float = 600.0,
        include_request: bool = False,
        include_response: bool = True,
    ) -> Job:
        """
        Poll a job until it reaches a terminal state, then fetch the final payload once.

        Polling remains lightweight by default and only the final retrieve can include
        request/response payloads.
        """
        if poll_interval_seconds <= 0:
            raise ValueError("poll_interval_seconds must be > 0")
        if timeout_seconds <= 0:
            raise ValueError("timeout_seconds must be > 0")

        started_at = time.monotonic()
        deadline = started_at + timeout_seconds
        while True:
            job = await self.retrieve(
                job_id,
                include_request=False,
                include_response=False,
            )
            if job.status in TERMINAL_JOB_STATUSES:
                if include_request or include_response:
                    return await self.retrieve(
                        job_id,
                        include_request=include_request,
                        include_response=include_response,
                    )
                return job

            now = time.monotonic()
            if now >= deadline:
                raise TimeoutError(f"Timed out waiting for job {job_id} completion after {timeout_seconds} seconds")
            sleep_for = min(poll_interval_seconds, max(deadline - now, 0.0))
            await asyncio.sleep(sleep_for)

    async def cancel(self, job_id: str) -> Job:
        """
        Cancel a queued or in-progress job.

        Args:
            job_id: The job ID to cancel

        Returns:
            Job: The updated job with status "cancelled"
        """
        prepared = self._prepare_cancel(job_id)
        response = await self._client._prepared_request(prepared)
        return Job.model_validate(response)

    async def retry(self, job_id: str) -> Job:
        """
        Retry a failed, cancelled, or expired job.

        Args:
            job_id: The job ID to retry

        Returns:
            Job: The updated job with status "queued"
        """
        prepared = self._prepare_retry(job_id)
        response = await self._client._prepared_request(prepared)
        return Job.model_validate(response)

    async def list(
        self,
        before: str | None = None,
        after: str | None = None,
        limit: int = 20,
        order: JobListOrder | None = "desc",
        id: str | None = None,
        status: JobStatus | None = None,
        endpoint: SupportedEndpoint | None = None,
        source: JobListSource | None = None,
        project_id: str | None = None,
        workflow_id: str | None = None,
        workflow_node_id: str | None = None,
        model: str | None = None,
        filename_regex: str | None = None,
        filename_contains: str | None = None,
        document_type: Sequence[str] | None = None,
        from_date: str | None = None,
        to_date: str | None = None,
        metadata: dict[str, str] | None = None,
        include_request: bool | None = None,
        include_response: bool | None = None,
    ) -> JobListResponse:
        """
        List jobs with pagination and optional filtering.

        Args:
            before: Pagination cursor (first ID from current page)
            after: Pagination cursor (last ID from previous page)
            limit: Number of jobs to return (1-100, default 20)
            order: Sort order by created_at ("asc" or "desc")
            id: Filter by job ID
            status: Filter by job status
            endpoint: Filter by endpoint
            source: High-level source filter (api/project/workflow)
            project_id: Filter by request.project_id
            workflow_id: Filter by metadata.workflow_id
            workflow_node_id: Filter by metadata.workflow_node_id or metadata.node_id
            model: Filter by request.model
            filename_regex: Regex/plain-text filename filter
            filename_contains: Substring filename filter
            document_type: Document type filters (can pass multiple)
            from_date: Filter jobs created on or after date (YYYY-MM-DD)
            to_date: Filter jobs created on or before date (YYYY-MM-DD)
            metadata: Exact metadata key/value filters
            include_request: Whether to include full request payloads
            include_response: Whether to include full response payloads

        Returns:
            JobListResponse: List of jobs with pagination info
        """
        prepared = self._prepare_list(
            before=before,
            after=after,
            limit=limit,
            order=order,
            id=id,
            status=status,
            endpoint=endpoint,
            source=source,
            project_id=project_id,
            workflow_id=workflow_id,
            workflow_node_id=workflow_node_id,
            model=model,
            filename_regex=filename_regex,
            filename_contains=filename_contains,
            document_type=document_type,
            from_date=from_date,
            to_date=to_date,
            metadata=metadata,
            include_request=include_request,
            include_response=include_response,
        )
        response = await self._client._prepared_request(prepared)
        return JobListResponse.model_validate(response)
