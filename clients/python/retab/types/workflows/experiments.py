"""Pydantic models mirroring the wire format of `/v1/workflows/{workflow_id}/experiments`.

Backend source of truth lives in
``backend/main_server/main_server/services/v1/workflows/experiments/``
(``models.py`` for storage, ``routes.py`` for request/response shapes).

Only the fields that ride the wire are mirrored here — internal storage-only
fields (e.g. ``last_run_config_fingerprint``) are intentionally omitted, and
any forensic/diagnostic blob the backend may evolve uses
``model_config = ConfigDict(extra="ignore")`` so SDK pins don't break when
the backend adds optional fields. ``handle_inputs`` is left as ``dict[str, Any]``
to match the existing `block_tests` convention (the SDK does not model the
``HandleInput`` discriminated union internally).
"""

from __future__ import annotations

import datetime
from typing import Annotated, Any, Literal, Union

from pydantic import ConfigDict, Field
from retab.types.base import RetabBaseModel


# ---------------------------------------------------------------------------
# Enums
# ---------------------------------------------------------------------------

NConsensusValue = Literal[3, 5, 7]
ExperimentBlockKind = Literal["extract", "classifier", "split", "for_each"]
ExperimentRunStatus = Literal["pending", "running", "completed", "error", "cancelled"]
ExperimentJobStatus = Literal["pending", "running", "completed", "error"]
ExperimentPublicStatus = Literal[
    "draft", "processing", "completed", "failed", "cancelled"
]
ExperimentTargetKind = Literal["field", "subdocument", "category", "key"]
ExperimentMetricView = Literal["summary", "by_document", "by_target", "votes"]
ExperimentSchemaDriftStatus = Literal["none", "drifted", "unknown"]


# ---------------------------------------------------------------------------
# Request payloads (sent by the client)
# ---------------------------------------------------------------------------


class ExperimentDocumentProvenance(RetabBaseModel):
    """Workflow execution metadata attached to a captured document."""

    model_config = ConfigDict(extra="ignore")

    workflow_run_id: str | None = None
    step_id: str | None = None


class ExperimentDocumentCaptureRequest(RetabBaseModel):
    """One document captured from workflow execution provenance.

    ``step_id`` selects a specific iteration when the source block lives
    inside a ``for_each`` container.
    """

    model_config = ConfigDict(extra="ignore")

    workflow_run_id: str
    step_id: str | None = None


class ExplicitExperimentDocumentRequest(RetabBaseModel):
    """A document with inlined ``handle_inputs`` and optional provenance.

    Use this when you want to drive an experiment with hand-crafted inputs
    or with inputs not derivable from a workflow run.
    """

    model_config = ConfigDict(extra="ignore")

    handle_inputs: dict[str, Any]
    provenance: ExperimentDocumentProvenance | None = None


# ---------------------------------------------------------------------------
# Stored entities (returned by the API)
# ---------------------------------------------------------------------------


class ExperimentDocument(RetabBaseModel):
    """A single experiment document, defined by its materialized inputs.

    Returned in nested experiment payloads (e.g. ``GET /experiments/{id}``).
    """

    model_config = ConfigDict(extra="ignore")

    id: str
    handle_inputs: dict[str, Any] = Field(default_factory=dict)
    handle_inputs_fingerprint: str = ""
    provenance: ExperimentDocumentProvenance | None = None


class StepArtifactRefMini(RetabBaseModel):
    """``(operation, id)`` pointer to a persisted backing record.

    Mirrors the backend's ``StepArtifactRef`` for experiment job content;
    the canonical record is fetched via the matching resource (e.g.
    ``client.extractions.get`` for ``operation="extract"``).
    """

    model_config = ConfigDict(extra="ignore")

    operation: str
    id: str


# ---------------------------------------------------------------------------
# Top-level experiment response
# ---------------------------------------------------------------------------


class ExperimentResponse(RetabBaseModel):
    """One row of the ``GET /experiments`` listing or single-experiment fetch."""

    model_config = ConfigDict(extra="ignore")

    id: str
    workflow_id: str
    block_id: str
    n_consensus: NConsensusValue
    document_count: int = 0
    name: str
    last_run_id: str | None = None
    created_at: datetime.datetime
    updated_at: datetime.datetime
    status: ExperimentPublicStatus = "draft"
    block_kind: ExperimentBlockKind
    score: float | None = None
    job_id: str | None = None
    is_stale: bool = False
    schema_drift: ExperimentSchemaDriftStatus = "unknown"
    schema_drift_detail: str | None = None


# ---------------------------------------------------------------------------
# Run create / response
# ---------------------------------------------------------------------------


class PreviousRunSummary(RetabBaseModel):
    """Slim summary of the previous completed run."""

    model_config = ConfigDict(extra="ignore")

    run_id: str
    definition_fingerprint: str | None = None
    score: float | None = None


class RunExperimentResponse(RetabBaseModel):
    """Response from ``POST /experiments/{id}/run``.

    Async — execution proceeds in the background; poll
    ``client.jobs.retrieve(job_id)`` (or call ``wait_for_run_completion``).
    """

    model_config = ConfigDict(extra="ignore")

    experiment_id: str
    run_id: str
    job_id: str | None
    status: str
    definition_fingerprint: str
    total_document_count: int = 0
    completed_document_count: int = 0
    document_count: int
    n_consensus: NConsensusValue
    previous_run: PreviousRunSummary | None = None
    noop: bool = False


class ExperimentRunSummary(RetabBaseModel):
    """One row of ``GET /experiments/{id}/runs``."""

    model_config = ConfigDict(extra="ignore")

    id: str
    parent_run_id: str | None = None
    block_config: dict[str, Any] | None = None
    definition_fingerprint: str
    documents_fingerprint: str
    status: ExperimentRunStatus
    block_kind: ExperimentBlockKind
    score: float | None = None
    total_document_count: int = 0
    completed_document_count: int = 0
    document_count: int = 0
    error_count: int = 0
    n_consensus: NConsensusValue
    created_at: datetime.datetime
    completed_at: datetime.datetime | None = None
    duration_ms: int | None = None
    job_id: str | None = None


class ExperimentRunListResponse(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    runs: list[ExperimentRunSummary] = Field(default_factory=list)


# ---------------------------------------------------------------------------
# Content (per-run job execution)
# ---------------------------------------------------------------------------


class ExperimentJobResponse(RetabBaseModel):
    """Execution content for one document job inside an experiment run."""

    model_config = ConfigDict(extra="ignore")

    id: str
    run_id: str
    experiment_id: str
    document_id: str
    status: ExperimentJobStatus
    block_kind: ExperimentBlockKind
    handle_inputs: dict[str, Any] = Field(default_factory=dict)
    artifact: StepArtifactRefMini | None = None
    error: str | None = None
    duration_ms: int | None = None
    created_at: datetime.datetime | None = None
    started_at: datetime.datetime | None = None
    completed_at: datetime.datetime | None = None
    attempt: int = 0
    is_placeholder: bool = False


class ExperimentContent(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    jobs: list[ExperimentJobResponse] = Field(default_factory=list)


class ExperimentContentResponse(RetabBaseModel):
    """``GET /experiments/{id}/content`` — execution content for one run."""

    model_config = ConfigDict(extra="ignore")

    experiment_id: str
    run_id: str
    content: ExperimentContent


# ---------------------------------------------------------------------------
# Metrics responses (discriminated union over ``view``)
# ---------------------------------------------------------------------------


class ExperimentSummaryMetricDocument(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    id: str
    filename: str
    score: float | None = None
    prior_score: float | None = None


class ExperimentConfusionFlowMetric(RetabBaseModel):
    model_config = ConfigDict(
        extra="ignore",
        populate_by_name=True,
        serialize_by_alias=True,
    )

    from_: str = Field(validation_alias="from", serialization_alias="from")
    to: str
    score: float


class ExperimentExtractSummaryAggregate(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    likelihoods: dict[str, float] = Field(default_factory=dict)


class ExperimentConfusionSummaryAggregate(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    diag: dict[str, float] = Field(default_factory=dict)
    flows: list[ExperimentConfusionFlowMetric] = Field(default_factory=list)


ExperimentSummaryAggregate = (
    ExperimentExtractSummaryAggregate | ExperimentConfusionSummaryAggregate
)


class ExperimentSummaryMetricsResponse(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    experiment_id: str
    run_id: str
    view: Literal["summary"] = "summary"
    definition_fingerprint: str | None = None
    block_kind: ExperimentBlockKind
    score: float | None = None
    prior_score: float | None = None
    documents: list[ExperimentSummaryMetricDocument] = Field(default_factory=list)
    aggregate: ExperimentSummaryAggregate | None = None
    prior_run_id: str | None = None


class ExperimentMetricDocumentRef(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    id: str
    filename: str


class ExperimentByDocumentTargetMetric(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    path: str
    score: float | None = None
    prior_score: float | None = None
    value: Any | None = None


class ExperimentDocumentConfusionMetric(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    diag: dict[str, float] = Field(default_factory=dict)
    flows: list[ExperimentConfusionFlowMetric] = Field(default_factory=list)


class ExperimentByDocumentMetricsResponse(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    run_id: str
    view: Literal["by_document"] = "by_document"
    document: ExperimentMetricDocumentRef
    score: float | None = None
    prior_score: float | None = None
    confusion: ExperimentDocumentConfusionMetric | None = None
    targets: list[ExperimentByDocumentTargetMetric] = Field(default_factory=list)


class ExperimentByTargetDocumentMetric(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    id: str
    filename: str
    score: float | None = None
    prior_score: float | None = None
    value: Any | None = None


class ExperimentTargetConfusionMetric(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    self: float | None = None
    flow_from: dict[str, float] = Field(default_factory=dict)
    flow_to: dict[str, float] = Field(default_factory=dict)


class ExperimentByTargetMetricsResponse(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    run_id: str
    view: Literal["by_target"] = "by_target"
    target: str
    score: float | None = None
    prior_score: float | None = None
    confusion: ExperimentTargetConfusionMetric | None = None
    documents: list[ExperimentByTargetDocumentMetric] = Field(default_factory=list)


class ExperimentVotesMetricDocument(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    id: str
    filename: str


class ExperimentVoteRow(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    consensus: Any | None = None
    votes: list[Any] = Field(default_factory=list)


class ExperimentVotesMetricsResponse(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    run_id: str
    view: Literal["votes"] = "votes"
    document: ExperimentVotesMetricDocument
    target: str
    score: float | None = None
    prior_score: float | None = None
    rows: list[ExperimentVoteRow] = Field(default_factory=list)


ExperimentMetricsViewResponse = Annotated[
    Union[
        ExperimentSummaryMetricsResponse,
        ExperimentByDocumentMetricsResponse,
        ExperimentByTargetMetricsResponse,
        ExperimentVotesMetricsResponse,
    ],
    Field(discriminator="view"),
]


class _MetricsStaleErrorLastRun(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    run_id: str
    definition_fingerprint: str | None = None
    score: float | None = None
    created_at: datetime.datetime | None = None


class ExperimentMetricsStaleError(RetabBaseModel):
    """Returned when the last run is stale vs the current draft config or
    document set. Recompute by calling ``experiments.runs.create(...)``.
    """

    model_config = ConfigDict(extra="ignore")

    error: Literal["stale_metrics"] = "stale_metrics"
    experiment_id: str
    stale_reasons: list[str] = Field(default_factory=list)
    last_run: _MetricsStaleErrorLastRun
    current_config_fingerprint: str | None = None
    message: str


class ExperimentMetricsMissingError(RetabBaseModel):
    """Returned when the experiment has no runs at all."""

    model_config = ConfigDict(extra="ignore")

    error: Literal["no_metrics"] = "no_metrics"
    experiment_id: str
    message: str


# Full response surface from ``GET /experiments/{id}/metrics``.
ExperimentMetricsResponse = Union[
    ExperimentMetricsViewResponse,
    ExperimentMetricsStaleError,
    ExperimentMetricsMissingError,
]


# ---------------------------------------------------------------------------
# Eligible blocks + run batch
# ---------------------------------------------------------------------------


class EligibleBlockSummary(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    block_id: str
    block_label: str
    block_type: str
    experiment_count: int
    drifted_experiment_count: int
    stale_experiment_count: int
    latest_run_at: datetime.datetime | None = None
    mean_score: float | None = None


class EligibleBlockListResponse(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    blocks: list[EligibleBlockSummary] = Field(default_factory=list)


class RunBatchResponse(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    block_id: str
    experiment_count: int
    runs: list[RunExperimentResponse] = Field(default_factory=list)


__all__ = [
    "NConsensusValue",
    "ExperimentBlockKind",
    "ExperimentRunStatus",
    "ExperimentJobStatus",
    "ExperimentPublicStatus",
    "ExperimentTargetKind",
    "ExperimentMetricView",
    "ExperimentSchemaDriftStatus",
    "ExperimentDocumentProvenance",
    "ExperimentDocumentCaptureRequest",
    "ExplicitExperimentDocumentRequest",
    "ExperimentDocument",
    "StepArtifactRefMini",
    "ExperimentResponse",
    "PreviousRunSummary",
    "RunExperimentResponse",
    "ExperimentRunSummary",
    "ExperimentRunListResponse",
    "ExperimentJobResponse",
    "ExperimentContent",
    "ExperimentContentResponse",
    "ExperimentSummaryMetricDocument",
    "ExperimentConfusionFlowMetric",
    "ExperimentExtractSummaryAggregate",
    "ExperimentConfusionSummaryAggregate",
    "ExperimentSummaryAggregate",
    "ExperimentSummaryMetricsResponse",
    "ExperimentMetricDocumentRef",
    "ExperimentByDocumentTargetMetric",
    "ExperimentDocumentConfusionMetric",
    "ExperimentByDocumentMetricsResponse",
    "ExperimentByTargetDocumentMetric",
    "ExperimentTargetConfusionMetric",
    "ExperimentByTargetMetricsResponse",
    "ExperimentVotesMetricDocument",
    "ExperimentVoteRow",
    "ExperimentVotesMetricsResponse",
    "ExperimentMetricsViewResponse",
    "ExperimentMetricsStaleError",
    "ExperimentMetricsMissingError",
    "ExperimentMetricsResponse",
    "EligibleBlockSummary",
    "EligibleBlockListResponse",
    "RunBatchResponse",
]
