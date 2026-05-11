"""Python SDK client for the workflow experiments API.

Mirrors the REST endpoints under ``/v1/workflows/{workflow_id}/experiments``.
The MCP tool surface (``experiments_create``, ``experiments_runs_create``,
``experiments_get_metrics``, ...) calls the same routes — this resource is
the SDK-side equivalent so callers don't have to construct raw requests.

Argument shapes accept either the typed Pydantic models or plain dicts.
Pydantic validates dicts at the boundary so callers who copy/paste from
the API docs work without import gymnastics, and callers who want type
checking can pass model instances.

Sub-resource:
    runs: per-experiment run history + run content inspection.

Layout follows the sibling ``WorkflowTests`` resource (mixin → sync → async)
so anyone reading this code recognizes the pattern.
"""

from __future__ import annotations

from typing import Any, Dict, List, Mapping, Sequence, Union

from pydantic import TypeAdapter

from ...._resource import AsyncAPIResource, SyncAPIResource
from ....types.standards import PreparedRequest
from ....types.workflows.experiments import (
    EligibleBlockListResponse,
    ExperimentContentResponse,
    ExperimentDocumentCaptureRequest,
    ExperimentMetricsResponse,
    ExperimentMetricView,
    ExperimentResponse,
    ExperimentRunListResponse,
    ExplicitExperimentDocumentRequest,
    NConsensusValue,
    RunBatchResponse,
    RunExperimentResponse,
)


# Shared TypeAdapter — reused for every metrics call so we don't re-introspect
# the discriminated-union schema on each request.
_METRICS_RESPONSE_ADAPTER: TypeAdapter[ExperimentMetricsResponse] = TypeAdapter(
    ExperimentMetricsResponse
)


def _dump_capture(
    capture: Union[ExperimentDocumentCaptureRequest, Mapping[str, Any]],
) -> Dict[str, Any]:
    if isinstance(capture, ExperimentDocumentCaptureRequest):
        return capture.model_dump(mode="json", exclude_none=True)
    return ExperimentDocumentCaptureRequest.model_validate(capture).model_dump(
        mode="json", exclude_none=True
    )


def _dump_explicit_document(
    document: Union[ExplicitExperimentDocumentRequest, Mapping[str, Any]],
) -> Dict[str, Any]:
    if isinstance(document, ExplicitExperimentDocumentRequest):
        return document.model_dump(mode="json", exclude_none=True)
    return ExplicitExperimentDocumentRequest.model_validate(document).model_dump(
        mode="json", exclude_none=True
    )


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
        return PreparedRequest(
            method="POST",
            url=f"/workflows/{workflow_id}/experiments",
            data=body,
        )

    def prepare_list(self, workflow_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url=f"/workflows/{workflow_id}/experiments",
        )

    def prepare_get(self, workflow_id: str, experiment_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url=f"/workflows/{workflow_id}/experiments/{experiment_id}",
        )

    def prepare_update(
        self,
        workflow_id: str,
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
            url=f"/workflows/{workflow_id}/experiments/{experiment_id}",
            data=body,
        )

    def prepare_delete(self, workflow_id: str, experiment_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="DELETE",
            url=f"/workflows/{workflow_id}/experiments/{experiment_id}",
        )

    def prepare_duplicate(self, workflow_id: str, experiment_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="POST",
            url=f"/workflows/{workflow_id}/experiments/{experiment_id}/duplicate",
            data={},
        )

    def prepare_cancel(self, workflow_id: str, experiment_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="POST",
            url=f"/workflows/{workflow_id}/experiments/{experiment_id}/cancel",
            data={},
        )

    def prepare_get_metrics(
        self,
        workflow_id: str,
        experiment_id: str,
        *,
        view: ExperimentMetricView = "summary",
        run_id: str | None = None,
        document_id: str | None = None,
        target_path: str | None = None,
        include_prior: bool = True,
        prior_run_id: str | None = None,
    ) -> PreparedRequest:
        # Always send ``view`` so the server doesn't have to fall back to a
        # default — the four views return different shapes and an explicit
        # value protects against accidental ``summary`` reads.
        params: Dict[str, Any] = {"view": view, "include_prior": include_prior}
        if run_id is not None:
            params["run_id"] = run_id
        if document_id is not None:
            params["document_id"] = document_id
        if target_path is not None:
            params["target_path"] = target_path
        if prior_run_id is not None:
            params["prior_run_id"] = prior_run_id
        return PreparedRequest(
            method="GET",
            url=f"/workflows/{workflow_id}/experiments/{experiment_id}/metrics",
            params=params,
        )

    def prepare_list_eligible_blocks(self, workflow_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url=f"/workflows/{workflow_id}/experiments/eligible-blocks",
        )

    def prepare_run_batch(
        self,
        workflow_id: str,
        *,
        block_id: str,
    ) -> PreparedRequest:
        data: Dict[str, Any] = {"block_id": block_id}
        return PreparedRequest(
            method="POST",
            url=f"/workflows/{workflow_id}/experiments/run-batch",
            data=data,
        )


class ExperimentRunsMixin:
    """Shared prepare_* methods for the runs sub-resource."""

    def prepare_create(
        self,
        workflow_id: str,
        experiment_id: str,
    ) -> PreparedRequest:
        # Empty-bodied ``{}`` POST is what the backend expects when no
        # overrides are passed — the route validates the request as
        # ``RunExperimentRequest`` which has all-optional fields.
        data: Dict[str, Any] = {}
        return PreparedRequest(
            method="POST",
            url=f"/workflows/{workflow_id}/experiments/{experiment_id}/run",
            data=data,
        )

    def prepare_list(self, workflow_id: str, experiment_id: str) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url=f"/workflows/{workflow_id}/experiments/{experiment_id}/runs",
        )

    def prepare_get(
        self,
        workflow_id: str,
        experiment_id: str,
        *,
        run_id: str | None = None,
    ) -> PreparedRequest:
        params: Dict[str, Any] = {}
        if run_id is not None:
            params["run_id"] = run_id
        return PreparedRequest(
            method="GET",
            url=f"/workflows/{workflow_id}/experiments/{experiment_id}/content",
            params=params or None,
        )

    def prepare_cancel_document(
        self,
        workflow_id: str,
        experiment_id: str,
        document_id: str,
    ) -> PreparedRequest:
        return PreparedRequest(
            method="POST",
            url=(
                f"/workflows/{workflow_id}/experiments/{experiment_id}"
                f"/documents/{document_id}/cancel"
            ),
            data={},
        )


# ---------------------------------------------------------------------------
# Sync
# ---------------------------------------------------------------------------


class ExperimentRuns(SyncAPIResource, ExperimentRunsMixin):
    """Per-experiment run history and per-document job inspection (sync)."""

    def create(
        self,
        workflow_id: str,
        experiment_id: str,
    ) -> RunExperimentResponse:
        """Ensure the experiment has fresh results for the current draft config.

        Args:
            workflow_id: Workflow ID
            experiment_id: Experiment to run

        Returns:
            ``RunExperimentResponse``. Fresh completed jobs are reused, failed
            or stale jobs are queued, and ``noop`` is true when no work was
            needed.
        """
        request = self.prepare_create(
            workflow_id,
            experiment_id,
        )
        response = self._client._prepared_request(request)
        return RunExperimentResponse.model_validate(response)

    def list(
        self,
        workflow_id: str,
        experiment_id: str,
    ) -> ExperimentRunListResponse:
        """List historical runs for one experiment, newest first."""
        request = self.prepare_list(workflow_id, experiment_id)
        response = self._client._prepared_request(request)
        return ExperimentRunListResponse.model_validate(response)

    def get(
        self,
        workflow_id: str,
        experiment_id: str,
        *,
        run_id: str | None = None,
    ) -> ExperimentContentResponse:
        """Get the per-document execution content of a run.

        ``run_id`` defaults to the latest run when omitted.
        """
        request = self.prepare_get(workflow_id, experiment_id, run_id=run_id)
        response = self._client._prepared_request(request)
        return ExperimentContentResponse.model_validate(response)

    def cancel_document(
        self,
        workflow_id: str,
        experiment_id: str,
        document_id: str,
    ) -> Dict[str, str]:
        """Cancel one pending/running document job in the latest active run."""
        request = self.prepare_cancel_document(workflow_id, experiment_id, document_id)
        response = self._client._prepared_request(request)
        return dict(response) if isinstance(response, dict) else {"status": str(response)}


class WorkflowExperiments(SyncAPIResource, WorkflowExperimentsMixin):
    """Workflow experiments API client (synchronous).

    Sub-clients:
        runs: per-experiment run history + per-document job inspection.

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
        >>> client.jobs.wait_for_completion(run.job_id)
        >>>
        >>> metrics = client.workflows.experiments.get_metrics(
        ...     workflow_id="wf_abc123",
        ...     experiment_id=exp.id,
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

    def list(self, workflow_id: str) -> List[ExperimentResponse]:
        """List all experiments attached to a workflow."""
        request = self.prepare_list(workflow_id)
        response = self._client._prepared_request(request)
        return [ExperimentResponse.model_validate(item) for item in response]

    def get(self, workflow_id: str, experiment_id: str) -> ExperimentResponse:
        """Fetch a single experiment by id (refreshes drift state)."""
        request = self.prepare_get(workflow_id, experiment_id)
        response = self._client._prepared_request(request)
        return ExperimentResponse.model_validate(response)

    def update(
        self,
        workflow_id: str,
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
            workflow_id,
            experiment_id,
            name=name,
            document_captures=document_captures,
            documents=documents,
            n_consensus=n_consensus,
        )
        response = self._client._prepared_request(request)
        return ExperimentResponse.model_validate(response)

    def delete(self, workflow_id: str, experiment_id: str) -> None:
        """Delete an experiment along with its runs and jobs."""
        request = self.prepare_delete(workflow_id, experiment_id)
        self._client._prepared_request(request)

    def duplicate(self, workflow_id: str, experiment_id: str) -> ExperimentResponse:
        """Duplicate an experiment, re-using the existing materialized documents."""
        request = self.prepare_duplicate(workflow_id, experiment_id)
        response = self._client._prepared_request(request)
        return ExperimentResponse.model_validate(response)

    def cancel(self, workflow_id: str, experiment_id: str) -> Dict[str, str]:
        """Cancel the latest pending/running run for an experiment.

        Returns a small ``{"status": "cancelled", "run_id": ...}`` dict.
        """
        request = self.prepare_cancel(workflow_id, experiment_id)
        response = self._client._prepared_request(request)
        return dict(response) if isinstance(response, dict) else {"status": str(response)}

    def get_metrics(
        self,
        workflow_id: str,
        experiment_id: str,
        *,
        view: ExperimentMetricView = "summary",
        run_id: str | None = None,
        document_id: str | None = None,
        target_path: str | None = None,
        include_prior: bool = True,
        prior_run_id: str | None = None,
    ) -> ExperimentMetricsResponse:
        """Get experiment metrics — consensus likelihoods (0.0–1.0).

        Views:
        - ``summary``: overall score + block-specific diagnostics.
        - ``by_document``: one document → all targets sorted ascending.
            Requires ``document_id``.
        - ``by_target``: one target → scores across all documents.
            Requires ``target_path``.
        - ``votes``: one document + selected target → consensus rows with
            flat voter values. Requires both ``document_id`` and
            ``target_path``.

        If the latest run is stale relative to the current block config or
        document set, the response carries an ``error="stale_metrics"``
        envelope; call ``runs.create(...)`` to recompute.
        """
        request = self.prepare_get_metrics(
            workflow_id,
            experiment_id,
            view=view,
            run_id=run_id,
            document_id=document_id,
            target_path=target_path,
            include_prior=include_prior,
            prior_run_id=prior_run_id,
        )
        response = self._client._prepared_request(request)
        return _METRICS_RESPONSE_ADAPTER.validate_python(response)

    def list_eligible_blocks(self, workflow_id: str) -> EligibleBlockListResponse:
        """List blocks in a workflow that support experiments, with rollups."""
        request = self.prepare_list_eligible_blocks(workflow_id)
        response = self._client._prepared_request(request)
        return EligibleBlockListResponse.model_validate(response)

    def run_batch(
        self,
        workflow_id: str,
        *,
        block_id: str,
    ) -> RunBatchResponse:
        """Trigger a run for every experiment attached to one block."""
        request = self.prepare_run_batch(
            workflow_id,
            block_id=block_id,
        )
        response = self._client._prepared_request(request)
        return RunBatchResponse.model_validate(response)


# ---------------------------------------------------------------------------
# Async
# ---------------------------------------------------------------------------


class AsyncExperimentRuns(AsyncAPIResource, ExperimentRunsMixin):
    """Per-experiment run history and per-document job inspection (async)."""

    async def create(
        self,
        workflow_id: str,
        experiment_id: str,
    ) -> RunExperimentResponse:
        request = self.prepare_create(
            workflow_id,
            experiment_id,
        )
        response = await self._client._prepared_request(request)
        return RunExperimentResponse.model_validate(response)

    async def list(
        self,
        workflow_id: str,
        experiment_id: str,
    ) -> ExperimentRunListResponse:
        request = self.prepare_list(workflow_id, experiment_id)
        response = await self._client._prepared_request(request)
        return ExperimentRunListResponse.model_validate(response)

    async def get(
        self,
        workflow_id: str,
        experiment_id: str,
        *,
        run_id: str | None = None,
    ) -> ExperimentContentResponse:
        request = self.prepare_get(workflow_id, experiment_id, run_id=run_id)
        response = await self._client._prepared_request(request)
        return ExperimentContentResponse.model_validate(response)

    async def cancel_document(
        self,
        workflow_id: str,
        experiment_id: str,
        document_id: str,
    ) -> Dict[str, str]:
        request = self.prepare_cancel_document(workflow_id, experiment_id, document_id)
        response = await self._client._prepared_request(request)
        return dict(response) if isinstance(response, dict) else {"status": str(response)}


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

    async def list(self, workflow_id: str) -> List[ExperimentResponse]:
        request = self.prepare_list(workflow_id)
        response = await self._client._prepared_request(request)
        return [ExperimentResponse.model_validate(item) for item in response]

    async def get(self, workflow_id: str, experiment_id: str) -> ExperimentResponse:
        request = self.prepare_get(workflow_id, experiment_id)
        response = await self._client._prepared_request(request)
        return ExperimentResponse.model_validate(response)

    async def update(
        self,
        workflow_id: str,
        experiment_id: str,
        *,
        name: str | None = None,
        document_captures: Sequence[Union[ExperimentDocumentCaptureRequest, Mapping[str, Any]]] | None = None,
        documents: Sequence[Union[ExplicitExperimentDocumentRequest, Mapping[str, Any]]] | None = None,
        n_consensus: NConsensusValue | None = None,
    ) -> ExperimentResponse:
        request = self.prepare_update(
            workflow_id,
            experiment_id,
            name=name,
            document_captures=document_captures,
            documents=documents,
            n_consensus=n_consensus,
        )
        response = await self._client._prepared_request(request)
        return ExperimentResponse.model_validate(response)

    async def delete(self, workflow_id: str, experiment_id: str) -> None:
        request = self.prepare_delete(workflow_id, experiment_id)
        await self._client._prepared_request(request)

    async def duplicate(
        self, workflow_id: str, experiment_id: str
    ) -> ExperimentResponse:
        request = self.prepare_duplicate(workflow_id, experiment_id)
        response = await self._client._prepared_request(request)
        return ExperimentResponse.model_validate(response)

    async def cancel(
        self, workflow_id: str, experiment_id: str
    ) -> Dict[str, str]:
        request = self.prepare_cancel(workflow_id, experiment_id)
        response = await self._client._prepared_request(request)
        return dict(response) if isinstance(response, dict) else {"status": str(response)}

    async def get_metrics(
        self,
        workflow_id: str,
        experiment_id: str,
        *,
        view: ExperimentMetricView = "summary",
        run_id: str | None = None,
        document_id: str | None = None,
        target_path: str | None = None,
        include_prior: bool = True,
        prior_run_id: str | None = None,
    ) -> ExperimentMetricsResponse:
        request = self.prepare_get_metrics(
            workflow_id,
            experiment_id,
            view=view,
            run_id=run_id,
            document_id=document_id,
            target_path=target_path,
            include_prior=include_prior,
            prior_run_id=prior_run_id,
        )
        response = await self._client._prepared_request(request)
        return _METRICS_RESPONSE_ADAPTER.validate_python(response)

    async def list_eligible_blocks(
        self, workflow_id: str
    ) -> EligibleBlockListResponse:
        request = self.prepare_list_eligible_blocks(workflow_id)
        response = await self._client._prepared_request(request)
        return EligibleBlockListResponse.model_validate(response)

    async def run_batch(
        self,
        workflow_id: str,
        *,
        block_id: str,
    ) -> RunBatchResponse:
        request = self.prepare_run_batch(
            workflow_id,
            block_id=block_id,
        )
        response = await self._client._prepared_request(request)
        return RunBatchResponse.model_validate(response)
