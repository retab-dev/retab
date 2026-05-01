"""Python SDK client for the workflow block-tests API.

Mirrors the eight backend endpoints under ``/v1/workflows/{workflow_id}/block-tests``
plus the nested run endpoints. Sync and async variants share a `WorkflowTestsMixin`
that builds `PreparedRequest`s — same pattern as `WorkflowRuns` /
`WorkflowRunsMixin` in the sibling `runs` resource.

Argument shapes accept either the typed Pydantic models or plain dicts.
Pydantic validates dicts at the boundary, so callers who copy/paste from
the docs (`{"type": "block", "block_id": "..."}` etc.) work without any
import gymnastics, AND callers who want type checking can pass model
instances.
"""

from __future__ import annotations

import asyncio
import time
from typing import Any, Dict, Mapping, Union

from pydantic import TypeAdapter

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.standards import PreparedRequest
from ....types.workflows.block_tests import (
    AssertionSpec,
    BlockTestBatchExecutionResult,
    BlockTestListResponse,
    BlockTestRunListResponse,
    ExecuteBlockTestsResponse,
    ManualWorkflowTestSource,
    RunStepWorkflowTestSource,
    WorkflowTest,
    WorkflowTestBlockTarget,
    WorkflowTestRunRecord,
    WorkflowTestSource,
)


# Job statuses that indicate the runner is done — terminal in the same sense
# as `TERMINAL_WORKFLOW_RUN_STATUSES` for workflow runs. Mirrors the backend's
# `JobStatus` literal in `retab/types/jobs.py`. Failed/cancelled/expired make
# the wait raise; only `completed` returns the parsed payload.
_TERMINAL_JOB_STATUSES: frozenset[str] = frozenset(
    {"completed", "failed", "cancelled", "expired"}
)

# Module-level TypeAdapters — building one per call would re-introspect the
# model's schema on every SDK call. With `Annotated[Union, discriminator]`
# the adapter validates the `type` field and dispatches to the right variant
# in one pass, with a clearer error than a hand-rolled if/elif tree.
_SOURCE_ADAPTER: TypeAdapter[WorkflowTestSource] = TypeAdapter(WorkflowTestSource)


# ``mode="json"`` so the dict has the wire shape the backend expects
# (datetimes as ISO strings, etc.) regardless of whether the caller passed
# a model instance or a dict.
def _dump_target(
    target: Union[WorkflowTestBlockTarget, Mapping[str, Any]],
) -> dict[str, Any]:
    if isinstance(target, WorkflowTestBlockTarget):
        return target.model_dump(mode="json")
    return WorkflowTestBlockTarget.model_validate(target).model_dump(mode="json")


def _dump_source(
    source: Union[WorkflowTestSource, ManualWorkflowTestSource, RunStepWorkflowTestSource, Mapping[str, Any]],
) -> dict[str, Any]:
    if isinstance(source, (ManualWorkflowTestSource, RunStepWorkflowTestSource)):
        return source.model_dump(mode="json")
    # Plain dicts go through the discriminated-union TypeAdapter, which
    # validates `type` and dispatches to the right variant in one pass.
    return _SOURCE_ADAPTER.validate_python(source).model_dump(mode="json")


def _dump_assertion(
    assertion: Union[AssertionSpec, Mapping[str, Any], None],
) -> dict[str, Any] | None:
    if assertion is None:
        return None
    if isinstance(assertion, AssertionSpec):
        return assertion.model_dump(mode="json", exclude_none=True)
    return AssertionSpec.model_validate(assertion).model_dump(mode="json", exclude_none=True)


class WorkflowTestsMixin:
    """Mixin shared by `WorkflowTests` (sync) and `AsyncWorkflowTests` (async)."""

    def prepare_create(
        self,
        workflow_id: str,
        *,
        target: Union[WorkflowTestBlockTarget, Mapping[str, Any]],
        source: Union[WorkflowTestSource, ManualWorkflowTestSource, RunStepWorkflowTestSource, Mapping[str, Any]],
        assertion: Union[AssertionSpec, Mapping[str, Any]],
        name: str | None = None,
    ) -> PreparedRequest:
        data: Dict[str, Any] = {
            "target": _dump_target(target),
            "source": _dump_source(source),
            "assertion": _dump_assertion(assertion),
        }
        if name is not None:
            data["name"] = name
        return PreparedRequest(
            method="POST",
            url=f"/workflows/{workflow_id}/block-tests",
            data=data,
        )

    def prepare_get(self, workflow_id: str, test_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url=f"/workflows/{workflow_id}/block-tests/{test_id}",
        )

    def prepare_list(
        self,
        workflow_id: str,
        *,
        target_block_id: str | None = None,
        limit: int = 50,
    ) -> PreparedRequest:
        params: Dict[str, Any] = {"limit": limit}
        if target_block_id is not None:
            params["target_block_id"] = target_block_id
        return PreparedRequest(
            method="GET",
            url=f"/workflows/{workflow_id}/block-tests",
            params=params,
        )

    def prepare_update(
        self,
        workflow_id: str,
        test_id: str,
        *,
        name: str | None = None,
        assertion: Union[AssertionSpec, Mapping[str, Any], None] = None,
        source: Union[WorkflowTestSource, ManualWorkflowTestSource, RunStepWorkflowTestSource, Mapping[str, Any], None] = None,
    ) -> PreparedRequest:
        # Only include fields the caller actually passed — the backend treats
        # missing fields as "leave untouched" and explicitly rejects an
        # ``assertion: null`` payload.
        data: Dict[str, Any] = {}
        if name is not None:
            data["name"] = name
        if assertion is not None:
            data["assertion"] = _dump_assertion(assertion)
        if source is not None:
            data["source"] = _dump_source(source)
        return PreparedRequest(
            method="PATCH",
            url=f"/workflows/{workflow_id}/block-tests/{test_id}",
            data=data,
        )

    def prepare_delete(self, workflow_id: str, test_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="DELETE",
            url=f"/workflows/{workflow_id}/block-tests/{test_id}",
        )

    def prepare_execute(
        self,
        workflow_id: str,
        *,
        test_id: str | None = None,
        target: Union[WorkflowTestBlockTarget, Mapping[str, Any], None] = None,
        n_consensus: int | None = None,
    ) -> PreparedRequest:
        data: Dict[str, Any] = {}
        if test_id is not None:
            data["test_id"] = test_id
        if target is not None:
            data["target"] = _dump_target(target)
        if n_consensus is not None:
            data["n_consensus"] = n_consensus
        return PreparedRequest(
            method="POST",
            url=f"/workflows/{workflow_id}/block-tests/execute",
            data=data,
        )


class WorkflowTestRunsMixin:
    """Mixin for the nested runs sub-resource."""

    def prepare_list(
        self,
        workflow_id: str,
        test_id: str,
        *,
        limit: int = 20,
    ) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url=f"/workflows/{workflow_id}/block-tests/{test_id}/runs",
            params={"limit": limit},
        )

    def prepare_get(
        self,
        workflow_id: str,
        test_id: str,
        run_id: str,
    ) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url=f"/workflows/{workflow_id}/block-tests/{test_id}/runs/{run_id}",
        )


class WorkflowTestRuns(SyncAPIResource, WorkflowTestRunsMixin):
    """Synchronous read-only access to a test's execution history."""

    def list(
        self,
        workflow_id: str,
        test_id: str,
        *,
        limit: int = 20,
    ) -> BlockTestRunListResponse:
        request = self.prepare_list(workflow_id, test_id, limit=limit)
        response = self._client._prepared_request(request)
        return BlockTestRunListResponse.model_validate(response)

    def get(
        self,
        workflow_id: str,
        test_id: str,
        run_id: str,
    ) -> WorkflowTestRunRecord:
        request = self.prepare_get(workflow_id, test_id, run_id)
        response = self._client._prepared_request(request)
        return WorkflowTestRunRecord.model_validate(response)


class AsyncWorkflowTestRuns(AsyncAPIResource, WorkflowTestRunsMixin):
    """Asynchronous read-only access to a test's execution history."""

    async def list(
        self,
        workflow_id: str,
        test_id: str,
        *,
        limit: int = 20,
    ) -> BlockTestRunListResponse:
        request = self.prepare_list(workflow_id, test_id, limit=limit)
        response = await self._client._prepared_request(request)
        return BlockTestRunListResponse.model_validate(response)

    async def get(
        self,
        workflow_id: str,
        test_id: str,
        run_id: str,
    ) -> WorkflowTestRunRecord:
        request = self.prepare_get(workflow_id, test_id, run_id)
        response = await self._client._prepared_request(request)
        return WorkflowTestRunRecord.model_validate(response)


class WorkflowTests(SyncAPIResource, WorkflowTestsMixin):
    """Workflow block-tests API client (synchronous).

    Sub-clients:
        runs: read-only access to per-test execution history.

    Example:
        >>> from retab import Retab
        >>> client = Retab(api_key="your-api-key")
        >>>
        >>> test = client.workflows.tests.create(
        ...     workflow_id="wf_abc123",
        ...     target={"type": "block", "block_id": "block_extract"},
        ...     source={"type": "manual", "handle_inputs": {}},
        ...     assertion={
        ...         "target": {"output_handle_id": "output-json-0", "path": "total"},
        ...         "condition": {"kind": "equals", "expected": 1234.56},
        ...     },
        ...     name="Q1 invoice total",
        ... )
        >>>
        >>> # Run all tests for the workflow asynchronously.
        >>> batch = client.workflows.tests.execute(workflow_id="wf_abc123")
        >>> print(batch.batch_id, batch.job_id)
    """

    def __init__(self, *args: Any, **kwargs: Any) -> None:
        super().__init__(*args, **kwargs)
        self.runs = WorkflowTestRuns(client=self._client)

    def create(
        self,
        workflow_id: str,
        *,
        target: Union[WorkflowTestBlockTarget, Mapping[str, Any]],
        source: Union[WorkflowTestSource, ManualWorkflowTestSource, RunStepWorkflowTestSource, Mapping[str, Any]],
        assertion: Union[AssertionSpec, Mapping[str, Any]],
        name: str | None = None,
    ) -> WorkflowTest:
        """Create a new block test against a single block in the workflow."""
        request = self.prepare_create(
            workflow_id,
            target=target,
            source=source,
            assertion=assertion,
            name=name,
        )
        response = self._client._prepared_request(request)
        return WorkflowTest.model_validate(response)

    def get(self, workflow_id: str, test_id: str) -> WorkflowTest:
        """Fetch a single block test by id (refreshes drift state)."""
        request = self.prepare_get(workflow_id, test_id)
        response = self._client._prepared_request(request)
        return WorkflowTest.model_validate(response)

    def list(
        self,
        workflow_id: str,
        *,
        target_block_id: str | None = None,
        limit: int = 50,
    ) -> BlockTestListResponse:
        """List all block tests for a workflow."""
        request = self.prepare_list(
            workflow_id,
            target_block_id=target_block_id,
            limit=limit,
        )
        response = self._client._prepared_request(request)
        return BlockTestListResponse.model_validate(response)

    def update(
        self,
        workflow_id: str,
        test_id: str,
        *,
        name: str | None = None,
        assertion: Union[AssertionSpec, Mapping[str, Any], None] = None,
        source: Union[WorkflowTestSource, ManualWorkflowTestSource, RunStepWorkflowTestSource, Mapping[str, Any], None] = None,
    ) -> WorkflowTest:
        """Patch the name, assertion, and/or source of a block test.

        Setting ``assertion=None`` is rejected by the backend — use
        ``delete()`` to remove the test entirely.
        """
        request = self.prepare_update(
            workflow_id,
            test_id,
            name=name,
            assertion=assertion,
            source=source,
        )
        response = self._client._prepared_request(request)
        return WorkflowTest.model_validate(response)

    def delete(self, workflow_id: str, test_id: str) -> None:
        """Delete a block test."""
        request = self.prepare_delete(workflow_id, test_id)
        self._client._prepared_request(request)

    def execute(
        self,
        workflow_id: str,
        *,
        test_id: str | None = None,
        target: Union[WorkflowTestBlockTarget, Mapping[str, Any], None] = None,
        n_consensus: int | None = None,
    ) -> ExecuteBlockTestsResponse:
        """Run one block test, all tests for a single block, or every test in a workflow.

        Provide EXACTLY ONE of:
          - ``test_id`` — run a single test by id.
          - ``target`` — run all tests for a single block.
          - neither — run every test in the workflow.

        ``n_consensus`` is optional; allowed values are 3, 5, or 7.

        Execution is asynchronous: the response carries a ``batch_id`` +
        ``job_id``. Either poll ``client.jobs.retrieve(job_id)`` directly,
        or call :meth:`wait_for_completion` for a typed convenience.
        """
        request = self.prepare_execute(
            workflow_id,
            test_id=test_id,
            target=target,
            n_consensus=n_consensus,
        )
        response = self._client._prepared_request(request)
        return ExecuteBlockTestsResponse.model_validate(response)

    def wait_for_completion(
        self,
        job_id: str,
        *,
        poll_interval_seconds: float = 2.0,
        timeout_seconds: float = 600.0,
    ) -> BlockTestBatchExecutionResult:
        """Poll the test-batch job until it reaches a terminal state and
        return the parsed result.

        Args:
            job_id: The ``job_id`` returned by :meth:`execute`.
            poll_interval_seconds: Seconds between polls (default 2.0).
            timeout_seconds: Maximum time to wait (default 600.0).

        Returns:
            ``BlockTestBatchExecutionResult`` parsed from the completed
            job's response body.

        Raises:
            RuntimeError: if the job ends in ``failed`` / ``cancelled`` /
                ``expired``. The message includes the backend's error
                payload when available.
            TimeoutError: if the job doesn't reach terminal within
                ``timeout_seconds``.
            ValueError: on invalid ``poll_interval_seconds`` /
                ``timeout_seconds``.
        """
        if poll_interval_seconds <= 0:
            raise ValueError("poll_interval_seconds must be > 0")
        if timeout_seconds <= 0:
            raise ValueError("timeout_seconds must be > 0")

        deadline = time.monotonic() + timeout_seconds
        while True:
            job = self._client.jobs.retrieve(job_id)
            if job.status in _TERMINAL_JOB_STATUSES:
                if job.status != "completed":
                    raise RuntimeError(
                        f"Test batch job {job_id} ended in status "
                        f"{job.status!r}: {job.error!r}"
                    )
                payload = job.response.body if job.response is not None else {}
                return BlockTestBatchExecutionResult.model_validate(payload)
            now = time.monotonic()
            if now >= deadline:
                raise TimeoutError(
                    f"Test batch job {job_id} did not complete within "
                    f"{timeout_seconds}s"
                )
            time.sleep(min(poll_interval_seconds, max(deadline - now, 0.0)))


class AsyncWorkflowTests(AsyncAPIResource, WorkflowTestsMixin):
    """Workflow block-tests API client (asynchronous)."""

    def __init__(self, *args: Any, **kwargs: Any) -> None:
        super().__init__(*args, **kwargs)
        self.runs = AsyncWorkflowTestRuns(client=self._client)

    async def create(
        self,
        workflow_id: str,
        *,
        target: Union[WorkflowTestBlockTarget, Mapping[str, Any]],
        source: Union[WorkflowTestSource, ManualWorkflowTestSource, RunStepWorkflowTestSource, Mapping[str, Any]],
        assertion: Union[AssertionSpec, Mapping[str, Any]],
        name: str | None = None,
    ) -> WorkflowTest:
        request = self.prepare_create(
            workflow_id,
            target=target,
            source=source,
            assertion=assertion,
            name=name,
        )
        response = await self._client._prepared_request(request)
        return WorkflowTest.model_validate(response)

    async def get(self, workflow_id: str, test_id: str) -> WorkflowTest:
        request = self.prepare_get(workflow_id, test_id)
        response = await self._client._prepared_request(request)
        return WorkflowTest.model_validate(response)

    async def list(
        self,
        workflow_id: str,
        *,
        target_block_id: str | None = None,
        limit: int = 50,
    ) -> BlockTestListResponse:
        request = self.prepare_list(
            workflow_id,
            target_block_id=target_block_id,
            limit=limit,
        )
        response = await self._client._prepared_request(request)
        return BlockTestListResponse.model_validate(response)

    async def update(
        self,
        workflow_id: str,
        test_id: str,
        *,
        name: str | None = None,
        assertion: Union[AssertionSpec, Mapping[str, Any], None] = None,
        source: Union[WorkflowTestSource, ManualWorkflowTestSource, RunStepWorkflowTestSource, Mapping[str, Any], None] = None,
    ) -> WorkflowTest:
        request = self.prepare_update(
            workflow_id,
            test_id,
            name=name,
            assertion=assertion,
            source=source,
        )
        response = await self._client._prepared_request(request)
        return WorkflowTest.model_validate(response)

    async def delete(self, workflow_id: str, test_id: str) -> None:
        request = self.prepare_delete(workflow_id, test_id)
        await self._client._prepared_request(request)

    async def execute(
        self,
        workflow_id: str,
        *,
        test_id: str | None = None,
        target: Union[WorkflowTestBlockTarget, Mapping[str, Any], None] = None,
        n_consensus: int | None = None,
    ) -> ExecuteBlockTestsResponse:
        request = self.prepare_execute(
            workflow_id,
            test_id=test_id,
            target=target,
            n_consensus=n_consensus,
        )
        response = await self._client._prepared_request(request)
        return ExecuteBlockTestsResponse.model_validate(response)

    async def wait_for_completion(
        self,
        job_id: str,
        *,
        poll_interval_seconds: float = 2.0,
        timeout_seconds: float = 600.0,
    ) -> BlockTestBatchExecutionResult:
        """Async equivalent of :meth:`WorkflowTests.wait_for_completion`."""
        if poll_interval_seconds <= 0:
            raise ValueError("poll_interval_seconds must be > 0")
        if timeout_seconds <= 0:
            raise ValueError("timeout_seconds must be > 0")

        deadline = time.monotonic() + timeout_seconds
        while True:
            job = await self._client.jobs.retrieve(job_id)
            if job.status in _TERMINAL_JOB_STATUSES:
                if job.status != "completed":
                    raise RuntimeError(
                        f"Test batch job {job_id} ended in status "
                        f"{job.status!r}: {job.error!r}"
                    )
                payload = job.response.body if job.response is not None else {}
                return BlockTestBatchExecutionResult.model_validate(payload)
            now = time.monotonic()
            if now >= deadline:
                raise TimeoutError(
                    f"Test batch job {job_id} did not complete within "
                    f"{timeout_seconds}s"
                )
            await asyncio.sleep(min(poll_interval_seconds, max(deadline - now, 0.0)))
