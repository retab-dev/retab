"""Python SDK client for the workflow tests API.

Mirrors the backend endpoints under ``/v1/workflows/tests``
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

from typing import Any, Dict, Mapping, Sequence, Union

from pydantic import TypeAdapter

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.pagination import PaginatedList
from ....types.standards import PreparedRequest
from ....types.workflows.tests import (
    AssertionSpec,
    ManualWorkflowTestSource,
    RunStepWorkflowTestSource,
    WorkflowTest,
    WorkflowTestBlockTarget,
    WorkflowTestResult,
    WorkflowTestRun,
    WorkflowTestSource,
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


def _prepare_create_run(
    workflow_id: str,
    *,
    test_id: str | None = None,
    target: Union[WorkflowTestBlockTarget, Mapping[str, Any], None] = None,
    n_consensus: int | None = None,
) -> PreparedRequest:
    data: Dict[str, Any] = {"workflow_id": workflow_id}
    if test_id is not None:
        data["test_id"] = test_id
    if target is not None:
        data["target"] = _dump_target(target)
    if n_consensus is not None:
        data["n_consensus"] = n_consensus
    return PreparedRequest(
        method="POST",
        url="/workflows/tests/runs",
        data=data,
    )


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
            "workflow_id": workflow_id,
            "target": _dump_target(target),
            "source": _dump_source(source),
            "assertion": _dump_assertion(assertion),
        }
        if name is not None:
            data["name"] = name
        return PreparedRequest(
            method="POST",
            url="/workflows/tests",
            data=data,
        )

    def prepare_get(self, test_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url=f"/workflows/tests/{test_id}",
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
            url=f"/workflows/tests?workflow_id={workflow_id}",
            params=params,
        )

    def prepare_update(
        self,
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
            url=f"/workflows/tests/{test_id}",
            data=data,
        )

    def prepare_delete(self, test_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="DELETE",
            url=f"/workflows/tests/{test_id}",
        )

    def prepare_create_run(
        self,
        workflow_id: str,
        *,
        test_id: str | None = None,
        target: Union[WorkflowTestBlockTarget, Mapping[str, Any], None] = None,
        n_consensus: int | None = None,
    ) -> PreparedRequest:
        return _prepare_create_run(
            workflow_id,
            test_id=test_id,
            target=target,
            n_consensus=n_consensus,
        )


class WorkflowTestRunsMixin:
    """Mixin for canonical workflow-test run lifecycle routes."""

    def prepare_list(
        self,
        *,
        workflow_id: str | None = None,
        test_id: str | None = None,
        target_block_id: str | None = None,
        status: str | None = None,
        statuses: Sequence[str] | str | None = None,
        exclude_status: str | None = None,
        trigger_type: str | None = None,
        trigger_types: Sequence[str] | str | None = None,
        from_date: str | None = None,
        to_date: str | None = None,
        sort_by: str | None = None,
        fields: Sequence[str] | str | None = None,
        before: str | None = None,
        after: str | None = None,
        limit: int = 20,
        order: str | None = None,
    ) -> PreparedRequest:
        params: Dict[str, Any] = {"limit": limit}
        for key, value in {
            "workflow_id": workflow_id,
            "test_id": test_id,
            "target_block_id": target_block_id,
            "status": status,
            "statuses": _join_csv(statuses),
            "exclude_status": exclude_status,
            "trigger_type": trigger_type,
            "trigger_types": _join_csv(trigger_types),
            "from_date": from_date,
            "to_date": to_date,
            "sort_by": sort_by,
            "fields": _join_csv(fields),
            "before": before,
            "after": after,
            "order": order,
        }.items():
            if value is not None:
                params[key] = value
        return PreparedRequest(
            method="GET",
            url="/workflows/tests/runs",
            params=params,
        )

    def prepare_create(
        self,
        workflow_id: str,
        *,
        test_id: str | None = None,
        target: Union[WorkflowTestBlockTarget, Mapping[str, Any], None] = None,
        n_consensus: int | None = None,
    ) -> PreparedRequest:
        return _prepare_create_run(
            workflow_id,
            test_id=test_id,
            target=target,
            n_consensus=n_consensus,
        )

    def prepare_get(self, run_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url=f"/workflows/tests/runs/{run_id}",
        )

    def prepare_cancel(self, run_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="POST",
            url=f"/workflows/tests/runs/{run_id}/cancel",
            data={},
        )


class WorkflowTestRunResultsMixin:
    """Mixin for workflow-test result rows filtered by run."""

    def prepare_list(self, run_id: str, *, limit: int = 20) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url="/workflows/tests/results",
            params={"run_id": run_id, "limit": limit},
        )

    def prepare_get(self, result_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url=f"/workflows/tests/results/{result_id}",
        )


def _join_csv(value: Sequence[str] | str | None) -> str | None:
    if value is None or isinstance(value, str):
        return value
    return ",".join(value)


class WorkflowTestRunResults(SyncAPIResource, WorkflowTestRunResultsMixin):
    def list(self, run_id: str, *, limit: int = 20) -> PaginatedList[WorkflowTestResult]:
        request = self.prepare_list(run_id, limit=limit)
        response = self._client._prepared_request(request)
        return PaginatedList[WorkflowTestResult].model_validate(response)

    def get(self, result_id: str) -> WorkflowTestResult:
        request = self.prepare_get(result_id)
        response = self._client._prepared_request(request)
        return WorkflowTestResult.model_validate(response)


class AsyncWorkflowTestRunResults(AsyncAPIResource, WorkflowTestRunResultsMixin):
    async def list(self, run_id: str, *, limit: int = 20) -> PaginatedList[WorkflowTestResult]:
        request = self.prepare_list(run_id, limit=limit)
        response = await self._client._prepared_request(request)
        return PaginatedList[WorkflowTestResult].model_validate(response)

    async def get(self, result_id: str) -> WorkflowTestResult:
        request = self.prepare_get(result_id)
        response = await self._client._prepared_request(request)
        return WorkflowTestResult.model_validate(response)


class WorkflowTestRuns(SyncAPIResource, WorkflowTestRunsMixin):
    """Synchronous access to workflow-test run lifecycle and results."""

    def __init__(self, *args: Any, **kwargs: Any) -> None:
        super().__init__(*args, **kwargs)
        self.results = WorkflowTestRunResults(client=self._client)

    def create(
        self,
        workflow_id: str,
        *,
        test_id: str | None = None,
        target: Union[WorkflowTestBlockTarget, Mapping[str, Any], None] = None,
        n_consensus: int | None = None,
    ) -> WorkflowTestRun:
        request = self.prepare_create(
            workflow_id,
            test_id=test_id,
            target=target,
            n_consensus=n_consensus,
        )
        response = self._client._prepared_request(request)
        return WorkflowTestRun.model_validate(response)

    def list(
        self,
        *,
        workflow_id: str | None = None,
        test_id: str | None = None,
        target_block_id: str | None = None,
        status: str | None = None,
        statuses: Sequence[str] | str | None = None,
        exclude_status: str | None = None,
        trigger_type: str | None = None,
        trigger_types: Sequence[str] | str | None = None,
        from_date: str | None = None,
        to_date: str | None = None,
        sort_by: str | None = None,
        fields: Sequence[str] | str | None = None,
        before: str | None = None,
        after: str | None = None,
        limit: int = 20,
        order: str | None = None,
    ) -> PaginatedList[WorkflowTestRun]:
        request = self.prepare_list(
            workflow_id=workflow_id,
            test_id=test_id,
            target_block_id=target_block_id,
            status=status,
            statuses=statuses,
            exclude_status=exclude_status,
            trigger_type=trigger_type,
            trigger_types=trigger_types,
            from_date=from_date,
            to_date=to_date,
            sort_by=sort_by,
            fields=fields,
            before=before,
            after=after,
            limit=limit,
            order=order,
        )
        response = self._client._prepared_request(request)
        return PaginatedList[WorkflowTestRun].model_validate(response)

    def get(self, run_id: str) -> WorkflowTestRun:
        request = self.prepare_get(run_id)
        response = self._client._prepared_request(request)
        return WorkflowTestRun.model_validate(response)

    def cancel(self, run_id: str) -> WorkflowTestRun:
        request = self.prepare_cancel(run_id)
        response = self._client._prepared_request(request)
        return WorkflowTestRun.model_validate(response)


class AsyncWorkflowTestRuns(AsyncAPIResource, WorkflowTestRunsMixin):
    """Asynchronous access to workflow-test run lifecycle and results."""

    def __init__(self, *args: Any, **kwargs: Any) -> None:
        super().__init__(*args, **kwargs)
        self.results = AsyncWorkflowTestRunResults(client=self._client)

    async def create(
        self,
        workflow_id: str,
        *,
        test_id: str | None = None,
        target: Union[WorkflowTestBlockTarget, Mapping[str, Any], None] = None,
        n_consensus: int | None = None,
    ) -> WorkflowTestRun:
        request = self.prepare_create(
            workflow_id,
            test_id=test_id,
            target=target,
            n_consensus=n_consensus,
        )
        response = await self._client._prepared_request(request)
        return WorkflowTestRun.model_validate(response)

    async def list(
        self,
        *,
        workflow_id: str | None = None,
        test_id: str | None = None,
        target_block_id: str | None = None,
        status: str | None = None,
        statuses: Sequence[str] | str | None = None,
        exclude_status: str | None = None,
        trigger_type: str | None = None,
        trigger_types: Sequence[str] | str | None = None,
        from_date: str | None = None,
        to_date: str | None = None,
        sort_by: str | None = None,
        fields: Sequence[str] | str | None = None,
        before: str | None = None,
        after: str | None = None,
        limit: int = 20,
        order: str | None = None,
    ) -> PaginatedList[WorkflowTestRun]:
        request = self.prepare_list(
            workflow_id=workflow_id,
            test_id=test_id,
            target_block_id=target_block_id,
            status=status,
            statuses=statuses,
            exclude_status=exclude_status,
            trigger_type=trigger_type,
            trigger_types=trigger_types,
            from_date=from_date,
            to_date=to_date,
            sort_by=sort_by,
            fields=fields,
            before=before,
            after=after,
            limit=limit,
            order=order,
        )
        response = await self._client._prepared_request(request)
        return PaginatedList[WorkflowTestRun].model_validate(response)

    async def get(self, run_id: str) -> WorkflowTestRun:
        request = self.prepare_get(run_id)
        response = await self._client._prepared_request(request)
        return WorkflowTestRun.model_validate(response)

    async def cancel(self, run_id: str) -> WorkflowTestRun:
        request = self.prepare_cancel(run_id)
        response = await self._client._prepared_request(request)
        return WorkflowTestRun.model_validate(response)


class WorkflowTests(SyncAPIResource, WorkflowTestsMixin):
    """Workflow tests API client (synchronous).

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
        >>> run = client.workflows.tests.runs.create("wf_abc123")
        >>> print(run.id, run.lifecycle.status)
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
        """Create a new workflow test against a single block in the workflow."""
        request = self.prepare_create(
            workflow_id,
            target=target,
            source=source,
            assertion=assertion,
            name=name,
        )
        response = self._client._prepared_request(request)
        return WorkflowTest.model_validate(response)

    def get(self, test_id: str) -> WorkflowTest:
        """Fetch a single workflow test by id (refreshes drift state)."""
        request = self.prepare_get(test_id)
        response = self._client._prepared_request(request)
        return WorkflowTest.model_validate(response)

    def list(
        self,
        workflow_id: str,
        *,
        target_block_id: str | None = None,
        limit: int = 50,
    ) -> PaginatedList[WorkflowTest]:
        """List all workflow tests for a workflow."""
        request = self.prepare_list(
            workflow_id,
            target_block_id=target_block_id,
            limit=limit,
        )
        response = self._client._prepared_request(request)
        return PaginatedList[WorkflowTest].model_validate(response)

    def update(
        self,
        test_id: str,
        *,
        name: str | None = None,
        assertion: Union[AssertionSpec, Mapping[str, Any], None] = None,
        source: Union[WorkflowTestSource, ManualWorkflowTestSource, RunStepWorkflowTestSource, Mapping[str, Any], None] = None,
    ) -> WorkflowTest:
        """Patch the name, assertion, and/or source of a workflow test.

        Setting ``assertion=None`` is rejected by the backend — use
        ``delete()`` to remove the test entirely.
        """
        request = self.prepare_update(
            test_id,
            name=name,
            assertion=assertion,
            source=source,
        )
        response = self._client._prepared_request(request)
        return WorkflowTest.model_validate(response)

    def delete(self, test_id: str) -> None:
        """Delete a workflow test."""
        request = self.prepare_delete(test_id)
        self._client._prepared_request(request)


class AsyncWorkflowTests(AsyncAPIResource, WorkflowTestsMixin):
    """Workflow tests API client (asynchronous)."""

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

    async def get(self, test_id: str) -> WorkflowTest:
        request = self.prepare_get(test_id)
        response = await self._client._prepared_request(request)
        return WorkflowTest.model_validate(response)

    async def list(
        self,
        workflow_id: str,
        *,
        target_block_id: str | None = None,
        limit: int = 50,
    ) -> PaginatedList[WorkflowTest]:
        request = self.prepare_list(
            workflow_id,
            target_block_id=target_block_id,
            limit=limit,
        )
        response = await self._client._prepared_request(request)
        return PaginatedList[WorkflowTest].model_validate(response)

    async def update(
        self,
        test_id: str,
        *,
        name: str | None = None,
        assertion: Union[AssertionSpec, Mapping[str, Any], None] = None,
        source: Union[WorkflowTestSource, ManualWorkflowTestSource, RunStepWorkflowTestSource, Mapping[str, Any], None] = None,
    ) -> WorkflowTest:
        request = self.prepare_update(
            test_id,
            name=name,
            assertion=assertion,
            source=source,
        )
        response = await self._client._prepared_request(request)
        return WorkflowTest.model_validate(response)

    async def delete(self, test_id: str) -> None:
        request = self.prepare_delete(test_id)
        await self._client._prepared_request(request)
