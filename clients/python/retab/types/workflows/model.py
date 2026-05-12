import datetime
from typing import Annotated, Any, Dict, List, Literal, Optional

from pydantic import Field, ConfigDict, computed_field, field_validator
from retab.types.base import RetabBaseModel

from retab.types.mime import FileRef

# ---------------------------------------------------------------------------
# BREAKING CHANGES (workflow step artifact + StepStatus shape cutover)
# ---------------------------------------------------------------------------
# - All step shapes now share :class:`StepCore` — :class:`StepStatus`,
#   :class:`StepExecutionResponse` and :class:`WorkflowRunStep` inherit
#   from it. ``duration_ms`` / ``loop_id`` /
#   ``iteration`` are computed properties on :class:`StepCore` (derived from
#   ``started_at`` / ``completed_at`` / ``loop_containers``); they are no
#   longer flat fields. ``updated_at`` is gone.
# - The old flat error/lifecycle fields are replaced by a single
#   discriminated :data:`TerminalState` payload under ``terminal``
#   (``TerminalError`` / ``TerminalSkipped`` / ``TerminalCancelled``). The
#   ``terminal`` field is present iff ``status`` is one of
#   error/skipped/cancelled.
# - ``iteration_context`` is replaced by a flat
#   ``loop_containers: List[ContainerContextData]`` on :class:`StepCore`;
#   :class:`IterationContextData` is removed.
# - ``StepStatus.retry_count`` is now ``int`` (default ``0``); a
#   never-retried step has count ``0`` rather than ``None``.
# - ``StepStatus`` / ``WorkflowRunStep`` / ``StepExecutionResponse`` no
#   longer carry ``metadata``, ``artifacts: list[...]``,
#   ``requires_human_review``, ``human_reviewed_at`` or
#   ``human_review_approved``. They expose a singular
#   ``artifact: StepArtifactRef | None`` pointer.
# - ``WorkflowArtifactOperation`` is extended with five new operations:
#   ``conditional_evaluation``, ``hil_evaluation``, ``while_loop_termination``,
#   ``api_call_invocation``, ``function_invocation``.
#   Each points at a dedicated backing-collection record:
#   :class:`ConditionalEvaluation`, :class:`HilEvaluation`,
#   :class:`WhileLoopTermination`, :class:`ApiCallInvocation`,
#   :class:`FunctionInvocation`.
#
# Migration: callers that previously read flat terminal details should switch
# to ``step.terminal`` (a ``TerminalState``).
# Callers that read ``step.metadata.evaluations`` /
# ``step.requires_human_review`` should fetch the artifact's backing
# record (one of the record types above) and read from there.
# ---------------------------------------------------------------------------

# Schemas are accessed via ``workflows.blocks.get(block_id).resolved_schemas``,
# not via step payload data. Step executions only carry data/payload;
# user-declared block config schemas (``start_json`` / ``extract`` /
# ``function`` / ``api_call``) live on the block itself, and every other
# block's input/output schema is inferred and exposed under
# ``resolved_schemas.input_schemas`` / ``resolved_schemas.output_schemas``.


class HandlePayload(RetabBaseModel):
    """
    Payload for a single output handle.

    Each output handle on a block produces a typed payload that can be:
    - file: A document reference (PDF, image, etc.)
    - json: Structured JSON data (extracted data, etc.)
    - text: Plain text content
    """
    type: Literal["file", "json", "text"] = Field(..., description="Type of payload")
    document: Optional[FileRef] = Field(default=None, description="For file handles: document reference")
    data: Any | None = Field(default=None, description="For JSON handles: structured data")
    text: Optional[str] = Field(default=None, description="For text payloads: text content")


def _normalize_handle_payloads(value: Any) -> Any:
    if value is None:
        return {}
    return value


# Workflow run payloads can contain newer backend block types before the SDK is
# regenerated. Keep runtime validation permissive so informational step metadata
# does not break run parsing.
BlockType = str
StepExecutionStatus = Literal[
    "pending",
    "queued",
    "running",
    "completed",
    "skipped",
    "error",
    "waiting_for_human",
    "cancelled",
]
WorkflowArtifactOperation = Literal[
    # Existing persisted-ref operations.
    "extraction",
    "split",
    "classification",
    "parse",
    "edit",
    "partition",
    # New persisted-ref operations introduced by the metadata cutover. Each
    # corresponds to a dedicated backing collection (see records below).
    "conditional_evaluation",
    "hil_evaluation",
    "while_loop_termination",
    "api_call_invocation",
    "function_invocation",
]


class StepArtifactRef(RetabBaseModel):
    """Canonical persisted resource produced by a workflow step.

    Uniformly an ``(operation, id)`` ref into a backing collection. The
    artifact itself carries no payload — consumers dispatch on ``operation``
    and fetch the backing record by ``id`` (e.g. via ``client.extractions.get``,
    or one of the workflow record helpers introduced for the new operations).
    """

    operation: WorkflowArtifactOperation = Field(
        ...,
        description="Persisted resource operation; identifies the backing collection",
    )
    id: str = Field(..., description="Persisted resource identifier")


class WorkflowArtifact(RetabBaseModel):
    """Dereferenced workflow artifact record.

    Returned by ``client.workflows.artifacts.get(...)`` and
    ``client.workflows.artifacts.list(...)``. It is the persisted artifact
    record flattened with the ref's ``operation`` injected at top level. New
    operation-specific fields are ignored until the SDK types them explicitly.
    """

    model_config = ConfigDict(extra="ignore")

    operation: WorkflowArtifactOperation = Field(
        ...,
        description="Persisted resource operation; identifies the backing collection",
    )
    id: str = Field(..., description="Persisted resource identifier")


class ContainerContextData(RetabBaseModel):
    """Structured context for a single container in the hierarchy."""
    container_id: str = Field(..., description="Container ID (e.g., 'while_loop-abc')")
    iteration: int = Field(..., description="Iteration index (0-based)")
    is_parallel: bool = Field(default=False, description="Whether this container represents a parallel item")
    parallel_item_index: Optional[int] = Field(default=None, description="Parallel item index if is_parallel")


class ErrorDetails(RetabBaseModel):
    """Detailed error information for debugging.

    Captures stack traces and context about where and why an error occurred.
    Mirrors backend ``ErrorDetails`` in ``services/v1/workflows/models.py``.
    """
    model_config = ConfigDict(extra="ignore")

    stack_trace: Optional[str] = Field(default=None, description="Full Python stack trace")
    block_id: Optional[str] = Field(default=None, description="ID of the block that failed")
    block_name: Optional[str] = Field(default=None, description="Name/label of the block that failed")
    error_code: Optional[str] = Field(default=None, description="Error code if available")
    context: Optional[dict] = Field(default=None, description="Additional context about the error")


# ---------------------------------------------------------------------------
# Terminal payloads — discriminated union over status in {error, skipped, cancelled}
# ---------------------------------------------------------------------------


class TerminalError(RetabBaseModel):
    """Terminal payload for ``status="error"``."""
    kind: Literal["error"] = "error"
    message: str = Field(..., description="Human-readable error message")
    # ``stage``/``category`` are typed enums on the canonical model
    # (``ExecutionStage`` / ``ErrorCategory``); kept as ``str`` here for
    # SDK-flexibility — the canonical backend is the source of truth.
    stage: Optional[str] = Field(default=None, description="Which execution stage failed")
    category: Optional[str] = Field(default=None, description="Category of error for retry decisions")
    details: Optional[ErrorDetails] = Field(default=None, description="Structured error context")


class TerminalSkipped(RetabBaseModel):
    """Terminal payload for ``status="skipped"``."""
    kind: Literal["skipped"] = "skipped"
    reason: str = Field(..., description="Reason the step was skipped")


class TerminalCancelled(RetabBaseModel):
    """Terminal payload for ``status="cancelled"``."""
    kind: Literal["cancelled"] = "cancelled"
    reason: str = Field(..., description="Reason the step was cancelled")


TerminalState = Annotated[
    TerminalError | TerminalSkipped | TerminalCancelled,
    Field(discriminator="kind"),
]


class ConditionEvaluationPerItem(RetabBaseModel):
    """Per-item evaluation result for wildcard array conditions."""
    indices: List[int] = Field(
        default_factory=list,
        description="Hierarchical indices for nested arrays (e.g., [0, 2, 1] for items[0].subitems[2].field[1])",
    )
    actual: Any = Field(default=None, description="Actual value at this index")
    matched: bool = Field(default=False, description="Whether this item matched the condition")


class ConditionEvaluationSubCondition(RetabBaseModel):
    """Evaluation result for a sub-condition in a compound condition."""
    sub_condition_id: str = Field(default="", description="Identifier for this sub-condition")
    path: str = Field(default="", description="JSON path that was evaluated")
    operator: str = Field(default="", description="Comparison operator used")
    expected: Any = Field(default=None, description="Expected value")
    actual: Any = Field(default=None, description="Actual value found")
    matched: bool = Field(default=False, description="Whether this sub-condition matched")
    per_item: Optional[List[ConditionEvaluationPerItem]] = Field(
        default=None,
        description="Per-item breakdown if this sub-condition used a wildcard path",
    )


class ConditionEvaluationDetails(RetabBaseModel):
    """Detailed evaluation information for frontend display."""
    path: str = Field(default="", description="JSON path that was evaluated")
    operator: str = Field(default="", description="Comparison operator used")
    expected: Any = Field(default=None, description="Expected value")
    actual: Any = Field(default=None, description="Actual value found")
    matched: bool = Field(default=False, description="Whether the condition matched")
    per_item: Optional[List[ConditionEvaluationPerItem]] = Field(
        default=None,
        description="Per-item breakdown for wildcard array conditions",
    )
    sub_conditions: Optional[List[ConditionEvaluationSubCondition]] = Field(
        default=None,
        description="Sub-condition evaluations for compound conditions",
    )
    logical_operator: Optional[Literal["and", "or"]] = Field(
        default=None,
        description="Logical operator combining sub-conditions",
    )


class ConditionEvaluationResult(RetabBaseModel):
    """Complete evaluation result for a routing/termination condition."""
    condition_id: str = Field(..., description="Unique identifier for this condition")
    path: str = Field(default="", description="JSON path that was evaluated")
    operator: str = Field(default="", description="Comparison operator used")
    expected: Any = Field(default=None, description="Expected value")
    actual: Any = Field(default=None, description="Actual value found")
    matched: bool = Field(default=False, description="Whether the condition matched")
    branch_name: str = Field(default="", description="Branch name selected by this condition")
    logical_operator: Optional[Literal["and", "or"]] = Field(
        default=None,
        description="Logical operator for compound conditions",
    )
    per_item: Optional[List[ConditionEvaluationPerItem]] = Field(
        default=None,
        description="Per-item breakdown for wildcard array conditions",
    )
    sub_evaluations: Optional[List[ConditionEvaluationSubCondition]] = Field(
        default=None,
        description="Sub-condition evaluations for compound conditions",
    )
    details: ConditionEvaluationDetails = Field(
        ...,
        description="Nested details object for frontend compatibility",
    )


# ---------------------------------------------------------------------------
# Backing-collection records for the persisted-ref operations.
#
# Each record is the canonical persisted form of a step's result. The workflow
# step references the record via :attr:`StepStatus.artifact` (an
# ``(operation, id)`` pointer); consumers dereference by fetching the record
# from its dedicated collection.
# ---------------------------------------------------------------------------


class ConditionalEvaluation(RetabBaseModel):
    """Persisted record of a conditional block's branch evaluation.

    Backing record for :data:`StepArtifactRef` with
    ``operation == "conditional_evaluation"``.
    """
    model_config = ConfigDict(extra="ignore")

    id: str = Field(..., description="Unique identifier")
    workflow_run_id: str = Field(..., description="Parent workflow run ID")
    step_id: str = Field(..., description="Producing step ID")
    evaluations: List[ConditionEvaluationResult] = Field(default_factory=list)
    selected_handles: List[str] = Field(default_factory=list)
    matched_branch_id: Optional[str] = Field(default=None)
    matched_condition_ids: List[str] = Field(default_factory=list)
    created_at: datetime.datetime = Field(..., description="When the record was created")


class HilEvaluation(RetabBaseModel):
    """Persisted record of a HIL block's branch evaluation.

    Same evaluation core as :class:`ConditionalEvaluation`, plus human-review
    state. Backing record for :data:`StepArtifactRef` with
    ``operation == "hil_evaluation"``.
    """
    model_config = ConfigDict(extra="ignore")

    id: str = Field(..., description="Unique identifier")
    workflow_run_id: str = Field(..., description="Parent workflow run ID")
    step_id: str = Field(..., description="Producing step ID")
    evaluations: List[ConditionEvaluationResult] = Field(default_factory=list)
    selected_handles: List[str] = Field(default_factory=list)
    matched_branch_id: Optional[str] = Field(default=None)
    matched_condition_ids: List[str] = Field(default_factory=list)
    requires_human_review: bool = Field(default=False)
    reviewer_id: Optional[str] = Field(default=None)
    review_decision: Optional[Literal["approved", "rejected", "needs_changes"]] = Field(default=None)
    review_notes: Optional[str] = Field(default=None)
    reviewed_at: Optional[datetime.datetime] = Field(default=None)
    created_at: datetime.datetime = Field(..., description="When the record was created")


class WhileLoopTermination(RetabBaseModel):
    """Persisted record of why a while-loop terminated.

    Backing record for :data:`StepArtifactRef` with
    ``operation == "while_loop_termination"``.
    """
    model_config = ConfigDict(extra="ignore")

    id: str = Field(..., description="Unique identifier")
    workflow_run_id: str = Field(..., description="Parent workflow run ID")
    step_id: str = Field(..., description="Producing step ID")
    termination_reason: Literal["max_iterations_reached", "condition_matched", "error"] = Field(
        ...,
        description="Why the while_loop terminated",
    )
    evaluations: List[ConditionEvaluationResult] = Field(
        default_factory=list,
        description="Termination condition evaluations recorded for the final iteration",
    )
    created_at: datetime.datetime = Field(..., description="When the record was created")


class ApiCallAttempt(RetabBaseModel):
    """One attempt of an api_call (initial + retries)."""
    model_config = ConfigDict(extra="ignore")

    attempt_number: int = Field(..., description="0-based attempt index")
    request_method: str = Field(...)
    request_url: str = Field(...)
    request_headers: Dict[str, str] = Field(default_factory=dict)
    request_body: Optional[Any] = Field(default=None)
    response_status: Optional[int] = Field(default=None)
    response_headers: Dict[str, str] = Field(default_factory=dict)
    response_body: Optional[Any] = Field(default=None)
    duration_ms: Optional[int] = Field(default=None)
    error: Optional[str] = Field(default=None)
    started_at: Optional[datetime.datetime] = Field(default=None)
    completed_at: Optional[datetime.datetime] = Field(default=None)


class ApiCallInvocation(RetabBaseModel):
    """Persisted record of an api_call block's invocation (with retry trace).

    Backing record for :data:`StepArtifactRef` with
    ``operation == "api_call_invocation"``.
    """
    model_config = ConfigDict(extra="ignore")

    id: str = Field(..., description="Unique identifier")
    workflow_run_id: str = Field(..., description="Parent workflow run ID")
    step_id: str = Field(..., description="Producing step ID")
    attempts: List[ApiCallAttempt] = Field(
        default_factory=list,
        description="Full retry trace; final attempt holds the canonical request/response",
    )
    error: Optional[str] = Field(default=None, description="Final error after exhausted retries, if any")
    created_at: datetime.datetime = Field(..., description="When the record was created")


class FunctionInvocation(RetabBaseModel):
    """Persisted record of a function block's invocation.

    Backing record for :data:`StepArtifactRef` with
    ``operation == "function_invocation"``.
    """
    model_config = ConfigDict(extra="ignore")

    id: str = Field(..., description="Unique identifier")
    workflow_run_id: str = Field(..., description="Parent workflow run ID")
    step_id: str = Field(..., description="Producing step ID")
    inputs: Dict[str, Any] = Field(default_factory=dict)
    output: Optional[Any] = Field(default=None)
    duration_ms: Optional[int] = Field(default=None)
    error: Optional[str] = Field(default=None)
    created_at: datetime.datetime = Field(..., description="When the record was created")


class StepCore(RetabBaseModel):
    """Shape shared by full step docs and read-projections of them.

    Mirrors the backend ``StepCore`` shape. :class:`StepStatus` and the
    public-API view classes :class:`StepExecutionResponse` /
    :class:`WorkflowRunStep` inherit from this. ``duration_ms`` /
    ``loop_id`` / ``iteration`` are computed from timestamps and
    ``loop_containers``; do not set them as flat fields.

    Per the WorkflowRun v2 cutover, ``StepStatusSummary`` was removed —
    listing endpoints return ``StepStatus`` instances with handle payloads
    projected away server-side via Mongo ``projection``.
    """
    model_config = ConfigDict(extra="ignore")

    block_id: str = Field(..., description="Logical ID of the block")
    # SDK-side: tolerate missing ``step_id`` on intermediate construction
    # paths (canonical model requires it).
    step_id: str = Field(default="", description="Full step ID with iteration context")
    block_type: BlockType = Field(..., description="Type of the block")
    block_label: str = Field(..., description="Label of the block")
    status: StepExecutionStatus = Field(..., description="Current status")
    started_at: Optional[datetime.datetime] = Field(default=None, description="When the step started")
    completed_at: Optional[datetime.datetime] = Field(default=None, description="When the step completed")
    terminal: Optional[TerminalState] = Field(
        default=None,
        description="Structured terminal payload — present iff status is error/skipped/cancelled",
    )
    model: Optional[str] = Field(default=None, description="LLM model used by this step, when applicable")
    loop_containers: List[ContainerContextData] = Field(
        default_factory=list,
        description="Container hierarchy from outermost to innermost. Empty when not inside any container.",
    )

    @computed_field
    @property
    def duration_ms(self) -> Optional[int]:
        """Derived from ``completed_at - started_at``."""
        if self.started_at and self.completed_at:
            return int((self.completed_at - self.started_at).total_seconds() * 1000)
        return None

    @computed_field
    @property
    def loop_id(self) -> Optional[str]:
        """Innermost container's ID. Derived from ``loop_containers[-1]``."""
        return self.loop_containers[-1].container_id if self.loop_containers else None

    @computed_field
    @property
    def iteration(self) -> Optional[int]:
        """Innermost container's iteration. Derived from ``loop_containers[-1]``."""
        return self.loop_containers[-1].iteration if self.loop_containers else None


class StepStatus(StepCore):
    """Full step status with all execution details.

    Mirrors the backend ``StepStatus`` document. ``handle_inputs``,
    ``handle_outputs``, and the singular :attr:`artifact` ref are flat,
    real fields on this model. Branch-evaluation / HIL / while-loop
    termination state lives on the backing record referenced by
    ``artifact`` (e.g. :class:`HilEvaluation`), not as flat fields here.
    The ``"waiting_for_human"`` status value encodes the awaiting-review
    signal; the actual review decision is on the ``HilEvaluation``
    backing record.
    """


    # Identity — SDK-side relaxed (canonical model requires both).
    run_id: str = Field(default="", description="Parent workflow run ID")

    # Persisted-doc-creation timestamp; distinct from ``started_at`` because
    # the step doc may be created before execution begins (queue time).
    created_at: Optional[datetime.datetime] = Field(default=None, description="When the step doc was first persisted")

    # Flat execution payload — no nested StepExecutionRecord wrapper.
    handle_inputs: Dict[str, HandlePayload] = Field(
        default_factory=dict,
        description="Handle input payloads consumed by this step",
    )
    handle_outputs: Dict[str, HandlePayload] = Field(
        default_factory=dict,
        description="Handle output payloads produced by this step",
    )
    artifact: Optional[StepArtifactRef] = Field(
        default=None,
        description=(
            "Canonical persisted resource produced by this step — a "
            "(operation, id) ref into a backing collection. ``None`` for "
            "steps that produce no canonical result (start/note/merge/"
            "sentinels). Every executed step produces at most one canonical "
            "artifact, so this is a singular ref, not a list."
        ),
    )

    # Retry tracking — int (not Optional[int]); a never-retried step has count 0.
    retry_count: int = Field(default=0, description="Number of retry attempts for this step execution")

    @field_validator("handle_inputs", "handle_outputs", mode="before")
    @classmethod
    def _handle_payloads_default_to_dict(cls, value: Any) -> Any:
        return _normalize_handle_payloads(value)

    @property
    def extracted_data(self) -> Optional[dict]:
        """Get the JSON output data from the default handle (``output-json-0``)."""
        if not self.handle_outputs:
            return None
        payload = self.handle_outputs.get("output-json-0")
        if isinstance(payload, HandlePayload) and payload.type == "json":
            return payload.data
        return None


# ---------------------------------------------------------------------------
# WorkflowRun v2 — typed sub-models
# ---------------------------------------------------------------------------


class WorkflowSnapshotRef(RetabBaseModel):
    """Reference to the workflow + immutable version driving the run."""
    workflow_id: str = Field(..., description="ID of the workflow that was run")
    version_id: str = Field(..., description="Content-addressed workflow version used for this run")
    name_at_run_time: str = Field(..., description="Workflow name at run-creation time")
    requested_version: str = Field(default="production", description="Raw version selector requested for this run")


class ManualTrigger(RetabBaseModel):
    type: Literal["manual"] = "manual"
    user_id: Optional[str] = Field(default=None)


class ApiTrigger(RetabBaseModel):
    type: Literal["api"] = "api"
    api_key_id: Optional[str] = Field(default=None)


class ScheduleTrigger(RetabBaseModel):
    type: Literal["schedule"] = "schedule"
    schedule_id: str = Field(...)


class WebhookTrigger(RetabBaseModel):
    type: Literal["webhook"] = "webhook"
    webhook_id: Optional[str] = Field(default=None)


class EmailTrigger(RetabBaseModel):
    type: Literal["email"] = "email"
    sender: str = Field(...)
    subject: Optional[str] = Field(default=None)


class RestartTrigger(RetabBaseModel):
    type: Literal["restart"] = "restart"
    parent_run_id: str = Field(...)


Trigger = Annotated[
    ManualTrigger | ApiTrigger | ScheduleTrigger | WebhookTrigger | EmailTrigger | RestartTrigger,
    Field(discriminator="type"),
]


class PendingRun(RetabBaseModel):
    kind: Literal["pending"] = "pending"


class RunningRun(RetabBaseModel):
    kind: Literal["running"] = "running"


class WaitingForHumanRun(RetabBaseModel):
    kind: Literal["waiting_for_human"] = "waiting_for_human"
    waiting_for_block_ids: List[str] = Field(default_factory=list)


class CompletedTerminal(RetabBaseModel):
    kind: Literal["completed"] = "completed"


class ErrorTerminal(RetabBaseModel):
    kind: Literal["error"] = "error"
    message: str = Field(...)
    stage: Optional[str] = Field(default=None)
    category: Optional[str] = Field(default=None)
    details: Optional[ErrorDetails] = Field(default=None)
    failing_step_id: Optional[str] = Field(default=None)


class CancelledTerminal(RetabBaseModel):
    kind: Literal["cancelled"] = "cancelled"
    reason: Optional[str] = Field(default=None)


RunLifecycle = Annotated[
    PendingRun | RunningRun | WaitingForHumanRun | CompletedTerminal | ErrorTerminal | CancelledTerminal,
    Field(discriminator="kind"),
]


class RunTiming(RetabBaseModel):
    """Timing information for a workflow run."""
    created_at: datetime.datetime = Field(...)
    started_at: Optional[datetime.datetime] = Field(default=None)
    completed_at: Optional[datetime.datetime] = Field(default=None)
    human_waiting_started_at: Optional[datetime.datetime] = Field(default=None)
    accumulated_human_waiting_ms: int = Field(default=0, ge=0)

    @computed_field  # type: ignore[prop-decorator]
    @property
    def duration_ms(self) -> Optional[int]:
        if self.started_at and self.completed_at:
            return int((self.completed_at - self.started_at).total_seconds() * 1000)
        return None

    @computed_field  # type: ignore[prop-decorator]
    @property
    def active_duration_ms(self) -> Optional[int]:
        wall = self.duration_ms
        if wall is None:
            return None
        return max(0, wall - self.accumulated_human_waiting_ms)


class RunInputs(RetabBaseModel):
    documents: Dict[str, FileRef] = Field(default_factory=dict)
    json_data: Dict[str, Any] = Field(default_factory=dict)


class WorkflowRunError(Exception):
    """Raised by :meth:`WorkflowRun.raise_for_status` when a run failed."""

    def __init__(self, run: "WorkflowRun") -> None:
        self.run = run
        lifecycle = run.lifecycle
        if isinstance(lifecycle, CancelledTerminal):
            msg = f"Workflow run {run.id} was cancelled"
            if lifecycle.reason:
                msg += f": {lifecycle.reason}"
        elif isinstance(lifecycle, ErrorTerminal):
            msg = f"Workflow run {run.id} failed: {lifecycle.message}"
        else:
            msg = f"Workflow run {run.id} did not succeed"
        super().__init__(msg)


class WorkflowRun(RetabBaseModel):
    """A stored workflow run record.

    Slim, typed, discriminated. Engine state is not surfaced; the terminal
    state is encoded in :attr:`lifecycle`, not loose flat fields.
    """
    model_config = ConfigDict(extra="ignore")

    id: str = Field(...)
    workflow: WorkflowSnapshotRef = Field(...)
    trigger: Trigger = Field(...)
    lifecycle: RunLifecycle = Field(...)
    timing: RunTiming = Field(...)
    inputs: RunInputs = Field(default_factory=RunInputs)

    def raise_for_status(self) -> None:
        """Raise :class:`WorkflowRunError` if the run did not succeed."""
        if isinstance(self.lifecycle, (ErrorTerminal, CancelledTerminal)):
            raise WorkflowRunError(self)

    @property
    def is_terminal(self) -> bool:
        return isinstance(self.lifecycle, (CompletedTerminal, ErrorTerminal, CancelledTerminal))



class Workflow(RetabBaseModel):
    """A stored workflow record."""
    model_config = ConfigDict(extra="ignore")

    class Published(RetabBaseModel):
        version_id: Optional[str] = Field(default=None, description="Published workflow version ID")
        published_at: Optional[datetime.datetime] = Field(default=None, description="When the workflow was last published")

    class EmailTriggerPolicy(RetabBaseModel):
        """Workflow CONFIG for inbound email triggers (allowlist policy).

        Renamed from ``EmailTrigger`` to disambiguate from the run-level
        :class:`EmailTrigger` discriminated-union variant defined at
        module top level (``WorkflowRun.trigger`` when triggered by email).
        """
        allowed_senders: List[str] = Field(default_factory=list, description="Allowed sender email addresses")
        allowed_domains: List[str] = Field(default_factory=list, description="Allowed sender email domains")

    id: str = Field(..., description="Unique ID for this workflow")
    name: str = Field(default="Untitled Workflow", description="Workflow name")
    description: str = Field(default="", description="Workflow description")
    published: Optional[Published] = Field(default=None, description="Published workflow metadata")
    email_trigger: EmailTriggerPolicy = Field(default_factory=EmailTriggerPolicy, description="Email trigger allowlist policy")
    created_at: datetime.datetime = Field(..., description="When the workflow was created")
    updated_at: datetime.datetime = Field(..., description="When the workflow was last updated")

    @property
    def published_version_id(self) -> Optional[str]:
        return self.published.version_id if self.published is not None else None

    @property
    def published_at(self) -> Optional[datetime.datetime]:
        return self.published.published_at if self.published is not None else None

# ---------------------------------------------------------------------------
# Type aliases
# ---------------------------------------------------------------------------

WorkflowRunStatus = Literal["pending", "running", "completed", "error", "waiting_for_human", "cancelled"]
WorkflowRunTriggerType = Literal["manual", "api", "schedule", "webhook", "email", "restart"]

TERMINAL_WORKFLOW_RUN_STATUSES: tuple[str, ...] = ("completed", "error", "cancelled")


class StepExecutionResponse(StepCore):
    """Step status, handle data, and artifact ref for a specific workflow step."""


    artifact: Optional[StepArtifactRef] = Field(
        default=None,
        description=(
            "Canonical persisted resource produced by this step "
            "(operation + id ref); None for steps that produce no "
            "canonical result"
        ),
    )
    handle_outputs: Dict[str, HandlePayload] = Field(
        default_factory=dict,
        description="Handle outputs keyed by handle ID",
    )
    handle_inputs: Dict[str, HandlePayload] = Field(
        default_factory=dict,
        description="Handle inputs keyed by handle ID (what this block received)",
    )

    @field_validator("handle_inputs", "handle_outputs", mode="before")
    @classmethod
    def _handle_payloads_default_to_dict(cls, value: Any) -> Any:
        return _normalize_handle_payloads(value)

    def get_json_output(self, handle_id: str = "output-json-0") -> Optional[dict]:
        """Get JSON data from a specific output handle.

        Most extract blocks emit JSON on ``output-json-0``.
        """
        if not self.handle_outputs:
            return None
        payload = self.handle_outputs.get(handle_id)
        if isinstance(payload, HandlePayload) and payload.type == "json":
            return payload.data
        return None

    @property
    def extracted_data(self) -> Optional[dict]:
        """Shorthand for ``get_json_output()`` — the most common access pattern."""
        return self.get_json_output()


class WorkflowRunStep(StepCore):
    """Persisted public step document returned by list workflow run steps.

    Backed by the same flattened backend ``StepStatus`` shape —
    ``handle_inputs``, ``handle_outputs``, and the singular ``artifact``
    ref are flat fields. Branch-evaluation / HIL / while-loop-termination
    state lives on the backing record referenced by ``artifact`` (e.g.
    :class:`HilEvaluation`).
    """


    run_id: str = Field(..., description="Parent workflow run ID")
    artifact: Optional[StepArtifactRef] = Field(
        default=None,
        description=(
            "Canonical persisted resource produced by this step "
            "(operation + id ref); None for steps that produce no "
            "canonical result"
        ),
    )
    handle_outputs: Dict[str, HandlePayload] = Field(
        default_factory=dict,
        description="Handle outputs keyed by handle ID",
    )
    handle_inputs: Dict[str, HandlePayload] = Field(
        default_factory=dict,
        description="Handle inputs keyed by handle ID",
    )
    retry_count: int = Field(default=0, description="Retry count for this step")
    created_at: Optional[datetime.datetime] = Field(default=None, description="When the step document was created")

    @field_validator("handle_inputs", "handle_outputs", mode="before")
    @classmethod
    def _handle_payloads_default_to_dict(cls, value: Any) -> Any:
        return _normalize_handle_payloads(value)

    @property
    def extracted_data(self) -> Optional[dict]:
        """Get the JSON output data from the default handle (``output-json-0``)."""
        if not self.handle_outputs:
            return None
        payload = self.handle_outputs.get("output-json-0")
        if isinstance(payload, HandlePayload) and payload.type == "json":
            return payload.data
        return None


# ---------------------------------------------------------------------------
# Cancel / Restart / HIL decision response types
# ---------------------------------------------------------------------------

class CancelWorkflowResponse(RetabBaseModel):
    """Response from cancelling a workflow run."""
    run: WorkflowRun
    cancellation_status: Literal["cancelled", "cancellation_requested", "cancellation_failed"] = Field(
        default="cancellation_requested",
        description="Cancellation delivery state",
    )


class HILDecisionResource(RetabBaseModel):
    """Temporal-owned decision state for a workflow HIL block."""
    run_id: str = Field(..., description="Workflow run ID")
    block_id: str = Field(..., description="HIL block ID")
    block_status: Optional[str] = Field(default=None, description="Current workflow block status")
    decision_received: bool = Field(default=False, description="Whether Temporal received the decision")
    decision_applied: bool = Field(default=False, description="Whether the workflow applied the decision")
    approved: Optional[bool] = Field(default=None, description="Approved or rejected decision value")
    modified_data: Optional[Dict[str, Any]] = Field(
        default=None,
        description="Modified data retained by Temporal for approved decisions",
    )
    payload_hash: Optional[str] = Field(default=None, description="Stable hash for the decision payload")
    received_at: Optional[datetime.datetime] = Field(
        default=None,
        description="When Temporal received the decision",
    )
    applied_at: Optional[datetime.datetime] = Field(
        default=None,
        description="When the workflow applied the decision",
    )


class SubmitHILDecisionResponse(RetabBaseModel):
    """Response from submitting a HIL decision."""
    submission_status: Literal["accepted", "already_received", "already_applied"] = Field(
        ...,
        description="Decision submission lifecycle status",
    )
    decision: HILDecisionResource = Field(
        ...,
        description="Temporal-owned HIL decision state for the block",
    )


class RunCountsResponse(RetabBaseModel):
    """Run counts grouped by status."""
    total: int = 0
    completed: int = 0
    running: int = 0
    error: int = 0
    pending: int = 0
    waiting_for_human: int = 0
    cancelled: int = 0


class ExecutionOrderResponse(RetabBaseModel):
    """DAG-ordered step IDs for a workflow run."""
    run_id: str = Field(..., description="Workflow run ID")
    ordered_step_ids: List[str] = Field(default_factory=list, description="Step IDs in DAG execution order")


class DocumentSignedUrlResponse(RetabBaseModel):
    """Signed URL for downloading a document from a run step."""
    signed_url: str = Field(..., description="Signed download URL")
    filename: str = Field(..., description="Original filename")
    mime_type: Optional[str] = Field(default=None, description="MIME type")


class ExportResponse(RetabBaseModel):
    """Export payload containing CSV data."""
    csv_data: str = Field(..., description="CSV content as string")
    rows: int = Field(..., description="Number of data rows")
    columns: int = Field(..., description="Number of columns")


# ---------------------------------------------------------------------------
# Workflow graph request models
# ---------------------------------------------------------------------------


class WorkflowBlockCreateRequest(RetabBaseModel):
    """Typed request payload for creating a workflow block."""
    model_config = ConfigDict(extra="ignore")

    id: str
    type: str
    label: str = ""
    position_x: float = 0
    position_y: float = 0
    width: Optional[float] = None
    height: Optional[float] = None
    config: Optional[dict] = None
    parent_id: Optional[str] = None


class WorkflowBlockUpdateRequest(RetabBaseModel):
    """Typed request payload for updating a workflow block."""
    model_config = ConfigDict(extra="ignore")

    block_id: str
    label: Optional[str] = None
    position_x: Optional[float] = None
    position_y: Optional[float] = None
    width: Optional[float] = None
    height: Optional[float] = None
    config: Optional[dict] = None
    parent_id: Optional[str] = None


class WorkflowEdgeCreateRequest(RetabBaseModel):
    """Typed request payload for creating a workflow edge."""
    model_config = ConfigDict(extra="ignore")

    id: str
    source_block: str
    target_block: str
    source_handle: Optional[str] = None
    target_handle: Optional[str] = None


# ---------------------------------------------------------------------------
# Workflow graph response models (blocks, edges)
# ---------------------------------------------------------------------------


class ResolvedSchemas(RetabBaseModel):
    """Graph-derived schemas attached to workflow blocks in transport responses."""
    model_config = ConfigDict(extra="ignore", populate_by_name=True)

    input_schemas: Dict[str, Any] = Field(
        default_factory=dict,
        description="Input JSON schemas keyed by sidecar slot.",
    )
    output_schemas: Dict[str, Any] = Field(
        default_factory=dict,
        description="Output JSON schemas keyed by output handle.",
    )
    field_ref_drift: Optional[dict] = Field(
        default=None,
        alias="_field_ref_drift",
        description="Field reference drift metadata when present.",
    )


class WorkflowBlock(RetabBaseModel):
    """A block in a workflow graph."""
    model_config = ConfigDict(extra="ignore")

    id: str = Field(..., description="Block ID")
    workflow_id: str = Field(..., description="Parent workflow ID")
    draft_version: Optional[str] = Field(default=None, description="Draft version for the live entity")
    type: str = Field(..., description="Block type (start, extract, parse, classifier, etc.)")
    label: str = Field(default="", description="Display label")
    position_x: float = Field(default=0, description="X position on canvas")
    position_y: float = Field(default=0, description="Y position on canvas")
    width: Optional[float] = Field(default=None, description="Block width")
    height: Optional[float] = Field(default=None, description="Block height")
    config: Optional[dict] = Field(default=None, description="Block-specific configuration")
    parent_id: Optional[str] = Field(default=None, description="Parent container block ID (while_loop, for_each)")
    resolved_schemas: Optional[ResolvedSchemas] = Field(
        default=None,
        description="Graph-derived schema sidecar. Schemas for block outputs live here, not on raw step results.",
    )
    updated_at: Optional[datetime.datetime] = Field(default=None, description="Last updated timestamp")


class WorkflowEdgeDoc(RetabBaseModel):
    """A persisted edge document connecting two blocks in a workflow graph."""
    model_config = ConfigDict(extra="ignore")

    id: str = Field(..., description="Edge ID")
    workflow_id: str = Field(..., description="Parent workflow ID")
    draft_version: Optional[str] = Field(default=None, description="Draft version for the live entity")
    source_block: str = Field(..., description="Source block ID")
    target_block: str = Field(..., description="Target block ID")
    source_handle: Optional[str] = Field(default=None, description="Output handle on source block")
    target_handle: Optional[str] = Field(default=None, description="Input handle on target block")
    updated_at: Optional[datetime.datetime] = Field(default=None, description="Last updated timestamp")


class WorkflowWithEntities(RetabBaseModel):
    """Complete workflow with its graph structure (blocks and edges)."""
    model_config = ConfigDict(extra="ignore")

    workflow: Workflow
    blocks: List[WorkflowBlock] = Field(default_factory=list)
    edges: List[WorkflowEdgeDoc] = Field(default_factory=list)

    @property
    def start_blocks(self) -> List[WorkflowBlock]:
        """Document input start blocks."""
        return [b for b in self.blocks if b.type == "start"]

    @property
    def start_json_blocks(self) -> List[WorkflowBlock]:
        """JSON input start blocks."""
        return [b for b in self.blocks if b.type == "start_json"]


class DeclarativePlanSummary(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    add: int = 0
    change: int = 0
    destroy: int = 0
    replace: int = 0
    noop: int = 0
    total: int = 0
    has_changes: bool = False


class DeclarativePlanFieldChange(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    path: List[str | int]
    path_display: str
    action: str
    before: Any | None = None
    after: Any | None = None
    before_sensitive: bool = False
    after_sensitive: bool = False
    unified_diff: str | None = None


class DeclarativePlanChange(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    before: Any | None = None
    after: Any | None = None
    before_sensitive: Any = Field(default_factory=dict)
    after_sensitive: Any = Field(default_factory=dict)
    field_changes: List[DeclarativePlanFieldChange] = Field(default_factory=list)


class DeclarativePlanResourceChange(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    address: str
    target: str
    target_id: str
    name: str
    type: str
    actions: List[str]
    summary: str
    change: DeclarativePlanChange
    path: str | None = None


class DeclarativeValidationResponse(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    workflow_id: str
    block_count: int
    edge_count: int
    is_valid: bool
    diagnostics: Dict[str, Any]


class DeclarativePlanResponse(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    workflow_id: str
    action: str
    block_count: int
    edge_count: int
    diagnostics: Dict[str, Any]
    format_version: str = "workflows-plan/v1"
    summary: DeclarativePlanSummary = Field(default_factory=DeclarativePlanSummary)
    resource_changes: List[DeclarativePlanResourceChange] = Field(default_factory=list)
    rendered_plan: str = "No changes. Infrastructure is up-to-date."


class DeclarativeApplyResponse(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    workflow_id: str
    created: bool
    block_count: int
    edge_count: int
    diagnostics: Dict[str, Any]
    format_version: str = "workflows-plan/v1"
    summary: DeclarativePlanSummary = Field(default_factory=DeclarativePlanSummary)
    resource_changes: List[DeclarativePlanResourceChange] = Field(default_factory=list)
    rendered_plan: str = "No changes. Infrastructure is up-to-date."


class DeclarativeExportResponse(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    workflow_id: str
    yaml_definition: str


# ---------------------------------------------------------------------------
# Diagnose response (POST /workflows/{id}/diagnose-graph)
# ---------------------------------------------------------------------------


class WorkflowDiagnosisIssue(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    severity: Literal["error", "warning", "info"]
    code: str = Field(..., description="Stable issue code")
    message: str = Field(..., description="Human-readable issue description")
    block_id: Optional[str] = Field(default=None, description="Related block when applicable")


class WorkflowDiagnosisStats(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    total_blocks: int = 0
    total_edges: int = 0
    block_types: Dict[str, int] = Field(default_factory=dict)
    start_blocks: int = 0


class WorkflowDiagnosisResponse(RetabBaseModel):
    """Result of ``POST /workflows/{id}/diagnose-graph``."""
    model_config = ConfigDict(extra="ignore")

    is_valid: bool
    issues: List[WorkflowDiagnosisIssue] = Field(default_factory=list)
    suggestions: List[str] = Field(default_factory=list)
    stats: WorkflowDiagnosisStats = Field(default_factory=WorkflowDiagnosisStats)


# ---------------------------------------------------------------------------
# Block simulation (POST /workflows/runs/{run_id}/steps/{block_id}/simulate)
# ---------------------------------------------------------------------------


class BlockSimulationIteration(RetabBaseModel):
    """One available iteration step exposed to simulate."""
    model_config = ConfigDict(extra="ignore")

    step_id: Optional[str] = None
    iteration_index: Optional[int] = None
    label: Optional[str] = None


class BlockSimulation(RetabBaseModel):
    """Result of replaying one block with the current draft config.

    Returned by ``client.workflows.blocks.simulate(...)``. Contains the
    inputs used, the produced outputs, and a canonical ``artifact`` ref
    when the block produces a persisted resource.
    """
    model_config = ConfigDict(extra="ignore")

    id: str = Field(..., description="Unique simulation ID")
    workflow_id: str
    run_id: str
    block_id: str
    block_type: str
    success: bool
    handle_inputs: Optional[Dict[str, Any]] = None
    artifact: Optional[StepArtifactRef] = None
    handle_outputs: Optional[Dict[str, Any]] = None
    routing_decision: Optional[List[str]] = Field(
        default=None,
        description="Active output handles for routing decisions (conditional/classifier).",
    )
    error: Optional[str] = None
    duration_ms: Optional[float] = None
    skipped: bool = False
    created_at: Optional[datetime.datetime] = None
    block_config: Optional[Dict[str, Any]] = Field(
        default=None,
        description="The draft block config used for this simulation.",
    )
    step_id: Optional[str] = Field(
        default=None,
        description="Step ID whose inputs were used (carries iteration prefix when applicable).",
    )
    available_iterations: Optional[List[BlockSimulationIteration]] = None
