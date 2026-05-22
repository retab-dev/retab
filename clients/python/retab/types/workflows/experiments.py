"""Pydantic models mirroring the wire format of `/v1/workflows/experiments`.

Backend source of truth lives in
``backend/main_server/main_server/services/v1/workflows/experiments/``
(``models.py`` for storage, ``routes.py`` for request/response shapes).

Only the fields that ride the wire are mirrored here — internal storage-only
fields (e.g. ``last_run_config_fingerprint``) are intentionally omitted, and
any forensic/diagnostic blob the backend may evolve uses
``model_config = ConfigDict(extra="ignore")`` so SDK pins don't break when
the backend adds optional fields.
"""

from __future__ import annotations

import datetime
from typing import Annotated, Any, Literal, Union

from pydantic import ConfigDict, Field
from retab.types.base import RetabBaseModel
from retab.types.workflows.model import ErrorDetails, HandleInput, WorkflowSnapshotRef


# ---------------------------------------------------------------------------
# Enums
# ---------------------------------------------------------------------------

NConsensusValue = Literal[3, 5, 7]
ExperimentBlockType = Literal["extract", "classifier", "split", "for_each"]
ExperimentRunStatus = Literal["pending", "running", "completed", "error", "cancelled"]
ExperimentPublicStatus = Literal["draft", "processing", "completed", "failed", "cancelled"]
ExperimentTargetKind = Literal["field", "subdocument", "category", "key"]
ExperimentMetricView = Literal["summary", "by_document", "by_target", "votes"]
ExperimentSchemaDriftStatus = Literal["none", "partial", "drifted", "unknown"]


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

    handle_inputs: dict[str, HandleInput]
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
    handle_inputs: dict[str, HandleInput] = Field(default_factory=dict)
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


class WorkflowExperiment(RetabBaseModel):
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
    block_type: ExperimentBlockType
    score: float | None = None
    is_stale: bool = False
    schema_drift: ExperimentSchemaDriftStatus = "unknown"
    schema_drift_detail: str | None = None


# ---------------------------------------------------------------------------
# Run create / response
# ---------------------------------------------------------------------------


class ExperimentRunTrigger(RetabBaseModel):
    """Experiment-run trigger source.

    Experiment runs keep the same top-level ``trigger`` concept as workflow
    runs, but use experiment-specific trigger strings such as ``manual_run``,
    and ``auto_retry:metrics``.
    """

    model_config = ConfigDict(extra="ignore")

    type: str | None = None


# --- ExperimentRun lifecycle discriminated union -----------------------------
#
# Mirror of the backend wire shape. Per the meta-pattern blueprint principle #3
# (invariants in shape, not prose): ``lifecycle`` references a discriminated
# union of typed states so per-state fields can carry data
# (``ErrorWorkflowExperimentRun.message`` / ``CancelledWorkflowExperimentRun.reason``).


class PendingWorkflowExperimentRun(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    status: Literal["pending"] = "pending"


class QueuedWorkflowExperimentRun(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    status: Literal["queued"] = "queued"


class RunningWorkflowExperimentRun(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    status: Literal["running"] = "running"


class CompletedWorkflowExperimentRun(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    status: Literal["completed"] = "completed"


class ErrorWorkflowExperimentRun(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    status: Literal["error"] = "error"
    message: str = Field(default="(no message)")
    details: ErrorDetails | None = None


class CancelledWorkflowExperimentRun(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    status: Literal["cancelled"] = "cancelled"
    reason: str | None = None


ExperimentRunLifecycle = Annotated[
    Union[
        PendingWorkflowExperimentRun,
        QueuedWorkflowExperimentRun,
        RunningWorkflowExperimentRun,
        CompletedWorkflowExperimentRun,
        ErrorWorkflowExperimentRun,
        CancelledWorkflowExperimentRun,
    ],
    Field(discriminator="status"),
]


class ExperimentRunTiming(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    created_at: datetime.datetime
    started_at: datetime.datetime | None = None
    completed_at: datetime.datetime | None = None
    duration_ms: int | None = None


class ExperimentRun(RetabBaseModel):
    """Parent experiment run.

    The run identity is ``id`` and lifecycle state lives under
    ``lifecycle.status``, matching workflow runs.
    """

    model_config = ConfigDict(extra="ignore")

    id: str
    workflow: WorkflowSnapshotRef
    trigger: "ExperimentRunTrigger"
    lifecycle: "ExperimentRunLifecycle"
    timing: "ExperimentRunTiming"
    experiment_id: str
    block_id: str
    block_type: ExperimentBlockType
    n_consensus: NConsensusValue
    definition_fingerprint: str
    documents_fingerprint: str
    score: float | None = None
    total_document_count: int = 0
    completed_document_count: int = 0
    document_count: int = 0
    error_count: int = 0


class CancelWorkflowExperimentRunResponse(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    id: str
    lifecycle: "ExperimentRunLifecycle"


# ---------------------------------------------------------------------------
# Results (per-document execution)
# ---------------------------------------------------------------------------


# --- ExperimentResult lifecycle discriminated union --------------------------


class PendingWorkflowExperimentResult(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    status: Literal["pending"] = "pending"


class QueuedWorkflowExperimentResult(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    status: Literal["queued"] = "queued"


class RunningWorkflowExperimentResult(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    status: Literal["running"] = "running"


class CompletedWorkflowExperimentResult(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    status: Literal["completed"] = "completed"


class ErrorWorkflowExperimentResult(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    status: Literal["error"] = "error"
    message: str = Field(default="(no message)")
    details: ErrorDetails | None = None


class CancelledWorkflowExperimentResult(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    status: Literal["cancelled"] = "cancelled"
    reason: str | None = None


ExperimentResultLifecycle = Annotated[
    Union[
        PendingWorkflowExperimentResult,
        QueuedWorkflowExperimentResult,
        RunningWorkflowExperimentResult,
        CompletedWorkflowExperimentResult,
        ErrorWorkflowExperimentResult,
        CancelledWorkflowExperimentResult,
    ],
    Field(discriminator="status"),
]


class ExperimentResultTiming(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    created_at: datetime.datetime | None = None
    started_at: datetime.datetime | None = None
    completed_at: datetime.datetime | None = None
    duration_ms: int | None = None


class ExperimentResult(RetabBaseModel):
    """Execution result for one document inside an experiment run.

    Result rows are addressed by ``document_id`` inside their parent run. The
    ``id`` field is retained as an internal row identifier.
    """

    model_config = ConfigDict(extra="ignore")

    id: str
    run_id: str
    experiment_id: str
    document_id: str
    lifecycle: "ExperimentResultLifecycle"
    timing: "ExperimentResultTiming"
    block_type: ExperimentBlockType
    handle_inputs: dict[str, HandleInput] = Field(default_factory=dict)
    artifact: StepArtifactRefMini | None = None
    duration_ms: int | None = None
    created_at: datetime.datetime | None = None
    started_at: datetime.datetime | None = None
    completed_at: datetime.datetime | None = None
    attempt: int = 0
    is_placeholder: bool = False


# ---------------------------------------------------------------------------
# Metrics responses (discriminated union over ``kind``)
# ---------------------------------------------------------------------------


class ExperimentSummaryMetricDocument(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    id: str
    filename: str
    score: float | None = None
    prior_score: float | None = None


class ExperimentConfusionFlowMetric(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    source: str
    target: str
    score: float


class ExperimentExtractSummaryAggregate(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    likelihoods: dict[str, float] = Field(default_factory=dict)


class ExperimentConfusionSummaryAggregate(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    diag: dict[str, float] = Field(default_factory=dict)
    flows: list[ExperimentConfusionFlowMetric] = Field(default_factory=list)


ExperimentSummaryAggregate = ExperimentExtractSummaryAggregate | ExperimentConfusionSummaryAggregate


class ExperimentSummaryMetricsResponse(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    experiment_id: str
    run_id: str
    kind: Literal["summary"] = "summary"
    view: Literal["summary"] = "summary"
    definition_fingerprint: str | None = None
    block_type: ExperimentBlockType
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
    kind: Literal["by_document"] = "by_document"
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
    kind: Literal["by_target"] = "by_target"
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
    score: float | None = None
    row_presence_score: float | None = None
    present_voter_count: int | None = None
    total_voter_count: int | None = None


class ExperimentVotesMetricsResponse(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    run_id: str
    kind: Literal["votes"] = "votes"
    view: Literal["votes"] = "votes"
    document: ExperimentVotesMetricDocument
    target: str
    score: float | None = None
    prior_score: float | None = None
    rows: list[ExperimentVoteRow] = Field(default_factory=list)


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

    kind: Literal["stale_metrics"] = "stale_metrics"
    error: Literal["stale_metrics"] = "stale_metrics"
    experiment_id: str
    stale_reasons: list[str] = Field(default_factory=list)
    last_run: _MetricsStaleErrorLastRun
    current_config_fingerprint: str | None = None
    message: str


class ExperimentMetricsMissingError(RetabBaseModel):
    """Returned when the experiment has no runs at all."""

    model_config = ConfigDict(extra="ignore")

    kind: Literal["no_metrics"] = "no_metrics"
    error: Literal["no_metrics"] = "no_metrics"
    experiment_id: str
    message: str


# Full response surface from ``GET /workflows/experiments/metrics``.
ExperimentMetricsResponse = Annotated[
    Union[
        ExperimentSummaryMetricsResponse,
        ExperimentByDocumentMetricsResponse,
        ExperimentByTargetMetricsResponse,
        ExperimentVotesMetricsResponse,
        ExperimentMetricsStaleError,
        ExperimentMetricsMissingError,
    ],
    Field(discriminator="kind"),
]


# ---------------------------------------------------------------------------
# Eligible blocks
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


__all__ = [
    "NConsensusValue",
    "ExperimentBlockType",
    "ExperimentRunStatus",
    "ExperimentPublicStatus",
    "ExperimentTargetKind",
    "ExperimentMetricView",
    "ExperimentSchemaDriftStatus",
    "ExperimentDocumentProvenance",
    "ExperimentDocumentCaptureRequest",
    "ExplicitExperimentDocumentRequest",
    "ExperimentDocument",
    "StepArtifactRefMini",
    "WorkflowExperiment",
    "ExperimentRunTrigger",
    "ExperimentRunLifecycle",
    "PendingWorkflowExperimentRun",
    "QueuedWorkflowExperimentRun",
    "RunningWorkflowExperimentRun",
    "CompletedWorkflowExperimentRun",
    "ErrorWorkflowExperimentRun",
    "CancelledWorkflowExperimentRun",
    "ExperimentRunTiming",
    "ExperimentRun",
    "CancelWorkflowExperimentRunResponse",
    "ExperimentResultLifecycle",
    "PendingWorkflowExperimentResult",
    "QueuedWorkflowExperimentResult",
    "RunningWorkflowExperimentResult",
    "CompletedWorkflowExperimentResult",
    "ErrorWorkflowExperimentResult",
    "CancelledWorkflowExperimentResult",
    "ExperimentResultTiming",
    "ExperimentResult",
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
    "ExperimentMetricsStaleError",
    "ExperimentMetricsMissingError",
    "ExperimentMetricsResponse",
    "EligibleBlockSummary",
    "EligibleBlockListResponse",
]
