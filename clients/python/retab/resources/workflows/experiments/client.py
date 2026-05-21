"""Python SDK client for the workflow experiments API.

Mirrors the REST endpoints under ``/v1/workflows/experiments``.
The MCP tool surface (``experiments_create``, ``experiments_runs_create``,
``experiments_runs_metrics_get``, ...) calls the same routes — this resource is
the SDK-side equivalent so callers don't have to construct raw requests.

Argument shapes accept either the typed Pydantic models or plain dicts.
Pydantic validates dicts at the boundary so callers who copy/paste from
the API docs work without import gymnastics, and callers who want type
checking can pass model instances.

Sub-resource:
    runs: experiment run lifecycle, child results, and metrics.

Layout follows the sibling ``WorkflowTests`` resource (mixin → sync → async)
so anyone reading this code recognizes the pattern.
"""

from __future__ import annotations

from typing import Any, Dict, Mapping, Sequence, Union

from pydantic import TypeAdapter

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.pagination import PaginatedList
from ....types.standards import PreparedRequest
from ....types.workflows.experiments import (
    ExperimentDocumentCaptureRequest,
    ExperimentMetricsResponse,
    ExperimentMetricView,
    ExperimentResult,
    ExperimentRun,
    ExperimentResponse,
    ExplicitExperimentDocumentRequest,
    NConsensusValue,
)


# Shared TypeAdapter — reused for every metrics call so we don't re-introspect
# the discriminated-union schema on each request.
_METRICS_RESPONSE_ADAPTER: TypeAdapter[ExperimentMetricsResponse] = TypeAdapter(ExperimentMetricsResponse)


def _dump_capture(
    capture: Union[ExperimentDocumentCaptureRequest, Mapping[str, Any]],
) -> Dict[str, Any]:
    if isinstance(capture, ExperimentDocumentCaptureRequest):
        return capture.model_dump(mode="json", exclude_none=True)
    return ExperimentDocumentCaptureRequest.model_validate(capture).model_dump(mode="json", exclude_none=True)


def _dump_explicit_document(
    document: Union[ExplicitExperimentDocumentRequest, Mapping[str, Any]],
) -> Dict[str, Any]:
    if isinstance(document, ExplicitExperimentDocumentRequest):
        return document.model_dump(mode="json", exclude_none=True)
    return ExplicitExperimentDocumentRequest.model_validate(document).model_dump(mode="json", exclude_none=True)


def _build_create_or_update_body(
    *,
    block_id: str | None,
    name: str | None,
    document_captures: Sequence[Union[ExperimentDocumentCaptureRequest, Mapping[str, Any]]] | None,
    documents: Sequence[Union[ExplicitExperimentDocumentRequest, Mapping[str, Any]]] | None,
    n_consensus: NConsensusValue | None,
) -> Dict[str, Any]:
    body: Dict[str, Any] = {}
    if block_id is not None:
        body["block_id"] = block_id
    if name is not None:
        body["name"] = name
    if document_captures is not None:
        body["document_captures"] = [_dump_capture(c) for c in document_captures]
    if documents is not None:
        body["documents"] = [_dump_explicit_document(d) for d in documents]
    if n_consensus is not None:
        body["n_consensus"] = n_consensus
    return body


# ---------------------------------------------------------------------------
# Mixins
# ---------------------------------------------------------------------------


class WorkflowExperimentsMixin:
    """Shared prepare_* methods used by both sync and async classes."""

    def prepare_create(
        self,
        workflow_id: str,
        *,
        block_id: str,
        name: str,
        document_captures: Sequence[Union[ExperimentDocumentCaptureRequest, Mapping[str, Any]]] | None = None,
        documents: Sequence[Union[ExplicitExperimentDocumentRequest, Mapping[str, Any]]] | None = None,
        n_consensus: NConsensusValue = 5,
    ) -> PreparedRequest:
        body = _build_create_or_update_body(
            block_id=block_id,
            name=name,
            document_captures=document_captures,
            documents=documents,
            n_consensus=n_consensus,
        )
        body["workflow_id"] = workflow_id
        return PreparedRequest(
            method="POST",
            url="/workflows/experiments",
            data=body,
        )

    def prepare_list(self, workflow_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url=f"/workflows/experiments?workflow_id={workflow_id}",
        )

    def prepare_get(self, experiment_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url=f"/workflows/experiments/{experiment_id}",
        )

    def prepare_update(
        self,
        experiment_id: str,
        *,
        name: str | None = None,
        document_captures: Sequence[Union[ExperimentDocumentCaptureRequest, Mapping[str, Any]]] | None = None,
        documents: Sequence[Union[ExplicitExperimentDocumentRequest, Mapping[str, Any]]] | None = None,
        n_consensus: NConsensusValue | None = None,
    ) -> PreparedRequest:
        body = _build_create_or_update_body(
            block_id=None,
            name=name,
            document_captures=document_captures,
            documents=documents,
            n_consensus=n_consensus,
        )
        return PreparedRequest(
            method="PATCH",
            url=f"/workflows/experiments/{experiment_id}",
            data=body,
        )

    def prepare_delete(self, experiment_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="DELETE",
            url=f"/workflows/experiments/{experiment_id}",
        )


def _join_csv(value: Sequence[str] | str | None) -> str | None:
    if value is None or isinstance(value, str):
        return value
    return ",".join(value)


class ExperimentRunsMixin:
    """Shared prepare_* methods for the runs sub-resource."""

    def prepare_create(
        self,
        workflow_id: str | None = None,
        experiment_id: str | None = None,
    ) -> PreparedRequest:
        if experiment_id is None:
            if workflow_id is None:
                raise TypeError("experiment_id is required")
            experiment_id = workflow_id
            workflow_id = None

        data: Dict[str, Any] = {"experiment_id": experiment_id}
        if workflow_id is not None:
            data["workflow_id"] = workflow_id
        return PreparedRequest(
            method="POST",
            url="/workflows/experiments/runs",
            data=data,
        )

    def prepare_list(
        self,
        *,
        workflow_id: str | None = None,
        experiment_id: str | None = None,
        block_id: str | None = None,
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
            "experiment_id": experiment_id,
            "block_id": block_id,
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
            url="/workflows/experiments/runs",
            params=params,
        )

    def prepare_get(self, run_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url=f"/workflows/experiments/runs/{run_id}",
        )

    def prepare_cancel(self, run_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="POST",
            url=f"/workflows/experiments/runs/{run_id}/cancel",
            data={},
        )


class ExperimentRunResultsMixin:
    def prepare_list(self, run_id: str, *, limit: int = 20) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url=f"/workflows/experiments/runs/{run_id}/results",
            params={"limit": limit},
        )


class ExperimentRunMetricsMixin:
    def prepare_get(
        self,
        run_id: str,
        *,
        view: ExperimentMetricView = "summary",
        document_id: str | None = None,
        target_path: str | None = None,
        include_prior: bool = True,
        prior_run_id: str | None = None,
    ) -> PreparedRequest:
        params: Dict[str, Any] = {"view": view, "include_prior": include_prior}
        if document_id is not None:
            params["document_id"] = document_id
        if target_path is not None:
            params["target_path"] = target_path
        if prior_run_id is not None:
            params["prior_run_id"] = prior_run_id
        return PreparedRequest(
            method="GET",
            url=f"/workflows/experiments/runs/{run_id}/metrics",
            params=params,
        )


# ---------------------------------------------------------------------------
# Sync
# ---------------------------------------------------------------------------


class ExperimentRuns(SyncAPIResource, ExperimentRunsMixin):
    """Experiment run lifecycle and per-document results (sync)."""

    def __init__(self, *args: Any, **kwargs: Any) -> None:
        super().__init__(*args, **kwargs)
        self.results = ExperimentRunResults(client=self._client)
        self.metrics = ExperimentRunMetrics(client=self._client)

    def create(
        self,
        workflow_id: str | None = None,
        experiment_id: str | None = None,
    ) -> ExperimentRun:
        """Ensure the experiment has fresh results for the current draft config.

        Args:
            experiment_id: Experiment to run
            workflow_id: Optional workflow ID

        Returns:
            ``ExperimentRun``. Inspect child result rows via
            ``client.workflows.experiments.runs.results``.
        """
        request = self.prepare_create(
            workflow_id,
            experiment_id,
        )
        response = self._client._prepared_request(request)
        return ExperimentRun.model_validate(response)

    def list(
        self,
        *,
        workflow_id: str | None = None,
        experiment_id: str | None = None,
        block_id: str | None = None,
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
    ) -> PaginatedList[ExperimentRun]:
        """List historical runs for one experiment, newest first."""
        request = self.prepare_list(
            workflow_id=workflow_id,
            experiment_id=experiment_id,
            block_id=block_id,
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
        return PaginatedList[ExperimentRun].model_validate(response)

    def get(self, run_id: str) -> ExperimentRun:
        request = self.prepare_get(run_id)
        response = self._client._prepared_request(request)
        return ExperimentRun.model_validate(response)

    def cancel(self, run_id: str) -> ExperimentRun:
        request = self.prepare_cancel(run_id)
        response = self._client._prepared_request(request)
        return ExperimentRun.model_validate(response)


class ExperimentRunResults(SyncAPIResource, ExperimentRunResultsMixin):
    def list(self, run_id: str, *, limit: int = 20) -> PaginatedList[ExperimentResult]:
        request = self.prepare_list(run_id, limit=limit)
        response = self._client._prepared_request(request)
        return PaginatedList[ExperimentResult].model_validate(response)


class ExperimentRunMetrics(SyncAPIResource, ExperimentRunMetricsMixin):
    def get(
        self,
        run_id: str,
        *,
        view: ExperimentMetricView = "summary",
        document_id: str | None = None,
        target_path: str | None = None,
        include_prior: bool = True,
        prior_run_id: str | None = None,
    ) -> ExperimentMetricsResponse:
        request = self.prepare_get(
            run_id,
            view=view,
            document_id=document_id,
            target_path=target_path,
            include_prior=include_prior,
            prior_run_id=prior_run_id,
        )
        response = self._client._prepared_request(request)
        return _METRICS_RESPONSE_ADAPTER.validate_python(response)


class WorkflowExperiments(SyncAPIResource, WorkflowExperimentsMixin):
    """Workflow experiments API client (synchronous).

    Sub-clients:
        runs: experiment run lifecycle, child results, and metrics.

    Example:
        >>> from retab import Retab
        >>> client = Retab(api_key="your-api-key")
        >>>
        >>> exp = client.workflows.experiments.create(
        ...     workflow_id="wf_abc123",
        ...     block_id="block_extract",
        ...     name="Q1 invoice totals",
        ...     document_captures=[
        ...         {"workflow_run_id": "run_xyz"},
        ...     ],
        ...     n_consensus=5,
        ... )
        >>>
        >>> run = client.workflows.experiments.runs.create(
        ...     workflow_id="wf_abc123", experiment_id=exp.id,
        ... )
        >>> print(run.id, run.lifecycle.status)
        >>>
        >>> metrics = client.workflows.experiments.runs.metrics.get(
        ...     run.id,
        ...     view="summary",
        ... )
    """

    def __init__(self, *args: Any, **kwargs: Any) -> None:
        super().__init__(*args, **kwargs)
        self.runs = ExperimentRuns(client=self._client)

    def create(
        self,
        workflow_id: str,
        *,
        block_id: str,
        name: str,
        document_captures: Sequence[Union[ExperimentDocumentCaptureRequest, Mapping[str, Any]]] | None = None,
        documents: Sequence[Union[ExplicitExperimentDocumentRequest, Mapping[str, Any]]] | None = None,
        n_consensus: NConsensusValue = 5,
    ) -> ExperimentResponse:
        """Create a consensus experiment on a supported block.

        Supported block kinds: ``extract``, ``classifier``, ``split``, and
        ``for_each`` configured with ``map_method="split_by_key"``.

        Provide documents via ``document_captures`` (provenance from a
        workflow run) and/or ``documents`` (explicit ``handle_inputs``).
        At least one document is required.

        Returns:
            ``ExperimentResponse``. Note: this does NOT trigger a run —
            call ``client.workflows.experiments.runs.create(...)`` next.
        """
        request = self.prepare_create(
            workflow_id,
            block_id=block_id,
            name=name,
            document_captures=document_captures,
            documents=documents,
            n_consensus=n_consensus,
        )
        response = self._client._prepared_request(request)
        return ExperimentResponse.model_validate(response)

    def list(self, workflow_id: str) -> PaginatedList[ExperimentResponse]:
        """List all experiments attached to a workflow."""
        request = self.prepare_list(workflow_id)
        response = self._client._prepared_request(request)
        return PaginatedList[ExperimentResponse].model_validate(response)

    def get(self, experiment_id: str) -> ExperimentResponse:
        """Fetch a single experiment by id (refreshes drift state)."""
        request = self.prepare_get(experiment_id)
        response = self._client._prepared_request(request)
        return ExperimentResponse.model_validate(response)

    def update(
        self,
        experiment_id: str,
        *,
        name: str | None = None,
        document_captures: Sequence[Union[ExperimentDocumentCaptureRequest, Mapping[str, Any]]] | None = None,
        documents: Sequence[Union[ExplicitExperimentDocumentRequest, Mapping[str, Any]]] | None = None,
        n_consensus: NConsensusValue | None = None,
    ) -> ExperimentResponse:
        """Patch the name, document set, and/or n_consensus.

        Changing the document set or n_consensus invalidates existing
        metrics — call ``runs.create(...)`` afterwards to recompute.
        """
        request = self.prepare_update(
            experiment_id,
            name=name,
            document_captures=document_captures,
            documents=documents,
            n_consensus=n_consensus,
        )
        response = self._client._prepared_request(request)
        return ExperimentResponse.model_validate(response)

    def delete(self, experiment_id: str) -> None:
        """Delete an experiment along with its runs and results."""
        request = self.prepare_delete(experiment_id)
        self._client._prepared_request(request)


# ---------------------------------------------------------------------------
# Async
# ---------------------------------------------------------------------------


class AsyncExperimentRuns(AsyncAPIResource, ExperimentRunsMixin):
    """Experiment run lifecycle and per-document results (async)."""

    def __init__(self, *args: Any, **kwargs: Any) -> None:
        super().__init__(*args, **kwargs)
        self.results = AsyncExperimentRunResults(client=self._client)
        self.metrics = AsyncExperimentRunMetrics(client=self._client)

    async def create(
        self,
        workflow_id: str | None = None,
        experiment_id: str | None = None,
    ) -> ExperimentRun:
        request = self.prepare_create(
            workflow_id,
            experiment_id,
        )
        response = await self._client._prepared_request(request)
        return ExperimentRun.model_validate(response)

    async def list(
        self,
        *,
        workflow_id: str | None = None,
        experiment_id: str | None = None,
        block_id: str | None = None,
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
    ) -> PaginatedList[ExperimentRun]:
        request = self.prepare_list(
            workflow_id=workflow_id,
            experiment_id=experiment_id,
            block_id=block_id,
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
        return PaginatedList[ExperimentRun].model_validate(response)

    async def get(self, run_id: str) -> ExperimentRun:
        request = self.prepare_get(run_id)
        response = await self._client._prepared_request(request)
        return ExperimentRun.model_validate(response)

    async def cancel(self, run_id: str) -> ExperimentRun:
        request = self.prepare_cancel(run_id)
        response = await self._client._prepared_request(request)
        return ExperimentRun.model_validate(response)


class AsyncExperimentRunResults(AsyncAPIResource, ExperimentRunResultsMixin):
    async def list(self, run_id: str, *, limit: int = 20) -> PaginatedList[ExperimentResult]:
        request = self.prepare_list(run_id, limit=limit)
        response = await self._client._prepared_request(request)
        return PaginatedList[ExperimentResult].model_validate(response)


class AsyncExperimentRunMetrics(AsyncAPIResource, ExperimentRunMetricsMixin):
    async def get(
        self,
        run_id: str,
        *,
        view: ExperimentMetricView = "summary",
        document_id: str | None = None,
        target_path: str | None = None,
        include_prior: bool = True,
        prior_run_id: str | None = None,
    ) -> ExperimentMetricsResponse:
        request = self.prepare_get(
            run_id,
            view=view,
            document_id=document_id,
            target_path=target_path,
            include_prior=include_prior,
            prior_run_id=prior_run_id,
        )
        response = await self._client._prepared_request(request)
        return _METRICS_RESPONSE_ADAPTER.validate_python(response)


class AsyncWorkflowExperiments(AsyncAPIResource, WorkflowExperimentsMixin):
    """Workflow experiments API client (asynchronous)."""

    def __init__(self, *args: Any, **kwargs: Any) -> None:
        super().__init__(*args, **kwargs)
        self.runs = AsyncExperimentRuns(client=self._client)

    async def create(
        self,
        workflow_id: str,
        *,
        block_id: str,
        name: str,
        document_captures: Sequence[Union[ExperimentDocumentCaptureRequest, Mapping[str, Any]]] | None = None,
        documents: Sequence[Union[ExplicitExperimentDocumentRequest, Mapping[str, Any]]] | None = None,
        n_consensus: NConsensusValue = 5,
    ) -> ExperimentResponse:
        request = self.prepare_create(
            workflow_id,
            block_id=block_id,
            name=name,
            document_captures=document_captures,
            documents=documents,
            n_consensus=n_consensus,
        )
        response = await self._client._prepared_request(request)
        return ExperimentResponse.model_validate(response)

    async def list(self, workflow_id: str) -> PaginatedList[ExperimentResponse]:
        request = self.prepare_list(workflow_id)
        response = await self._client._prepared_request(request)
        return PaginatedList[ExperimentResponse].model_validate(response)

    async def get(self, experiment_id: str) -> ExperimentResponse:
        request = self.prepare_get(experiment_id)
        response = await self._client._prepared_request(request)
        return ExperimentResponse.model_validate(response)

    async def update(
        self,
        experiment_id: str,
        *,
        name: str | None = None,
        document_captures: Sequence[Union[ExperimentDocumentCaptureRequest, Mapping[str, Any]]] | None = None,
        documents: Sequence[Union[ExplicitExperimentDocumentRequest, Mapping[str, Any]]] | None = None,
        n_consensus: NConsensusValue | None = None,
    ) -> ExperimentResponse:
        request = self.prepare_update(
            experiment_id,
            name=name,
            document_captures=document_captures,
            documents=documents,
            n_consensus=n_consensus,
        )
        response = await self._client._prepared_request(request)
        return ExperimentResponse.model_validate(response)

    async def delete(self, experiment_id: str) -> None:
        request = self.prepare_delete(experiment_id)
        await self._client._prepared_request(request)
