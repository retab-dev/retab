"""
Jobs API Resource

Provides synchronous and asynchronous clients for the Jobs API.
"""

from typing import Any

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.jobs import Job, JobListResponse, JobStatus, SupportedEndpoint
from ...types.standards import PreparedRequest


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
        after: str | None = None,
        limit: int = 20,
        status: JobStatus | None = None,
        include_request: bool | None = None,
        include_response: bool | None = None,
    ) -> PreparedRequest:
        params: dict[str, Any] = {"limit": limit}
        if after is not None:
            params["after"] = after
        if status is not None:
            params["status"] = status
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
        ...         "model": "gpt-5.2",
        ...     }
        ... )
        >>>
        >>> # Poll for completion
        >>> while job.status not in ("completed", "failed", "cancelled"):
        ...     import time
        ...     time.sleep(5)
        ...     job = client.jobs.retrieve(job.id)
        >>>
        >>> if job.status == "completed":
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
        include_request: bool | None = None,
        include_response: bool | None = None,
    ) -> Job:
        """
        Retrieve a job by ID.

        Args:
            job_id: The job ID to retrieve
            include_request: Whether to include the original request payload
            include_response: Whether to include response payload/documents

        Returns:
            Job: The job with current status and result (if completed)
        """
        prepared = self._prepare_retrieve_with_options(job_id, include_request, include_response)
        response = self._client._prepared_request(prepared)
        return Job.model_validate(response)

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
        after: str | None = None,
        limit: int = 20,
        status: JobStatus | None = None,
        include_request: bool | None = None,
        include_response: bool | None = None,
    ) -> JobListResponse:
        """
        List jobs with pagination and optional status filtering.

        Args:
            after: Pagination cursor (last ID from previous page)
            limit: Number of jobs to return (1-100, default 20)
            status: Filter by job status
            include_request: Whether to include full request payloads
            include_response: Whether to include full response payloads

        Returns:
            JobListResponse: List of jobs with pagination info
        """
        prepared = self._prepare_list(after, limit, status, include_request, include_response)
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
        ...         "model": "gpt-5.2",
        ...     }
        ... )
        >>>
        >>> # Poll for completion
        >>> while job.status not in ("completed", "failed", "cancelled"):
        ...     import asyncio
        ...     await asyncio.sleep(5)
        ...     job = await client.jobs.retrieve(job.id)
        >>>
        >>> if job.status == "completed":
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
        include_request: bool | None = None,
        include_response: bool | None = None,
    ) -> Job:
        """
        Retrieve a job by ID.

        Args:
            job_id: The job ID to retrieve
            include_request: Whether to include the original request payload
            include_response: Whether to include response payload/documents

        Returns:
            Job: The job with current status and result (if completed)
        """
        prepared = self._prepare_retrieve_with_options(job_id, include_request, include_response)
        response = await self._client._prepared_request(prepared)
        return Job.model_validate(response)

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
        after: str | None = None,
        limit: int = 20,
        status: JobStatus | None = None,
        include_request: bool | None = None,
        include_response: bool | None = None,
    ) -> JobListResponse:
        """
        List jobs with pagination and optional status filtering.

        Args:
            after: Pagination cursor (last ID from previous page)
            limit: Number of jobs to return (1-100, default 20)
            status: Filter by job status
            include_request: Whether to include full request payloads
            include_response: Whether to include full response payloads

        Returns:
            JobListResponse: List of jobs with pagination info
        """
        prepared = self._prepare_list(after, limit, status, include_request, include_response)
        response = await self._client._prepared_request(prepared)
        return JobListResponse.model_validate(response)
