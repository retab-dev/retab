import datetime
from typing import Any, ClassVar, Dict, List, Literal, Optional, Set

from pydantic import BaseModel, Field, ConfigDict, model_validator

from retab.types.mime import FileRef

# ---------------------------------------------------------------------------
# BREAKING CHANGES (workflow step artifact cutover)
# ---------------------------------------------------------------------------
# - ``StepArtifactRef`` was renamed to :class:`StepArtifact` (same shape).
# - ``StepArtifactView``, ``StepExecutionMetadata``, ``RoutingMetadata``,
#   ``ContainerMetadata``, ``EvaluationMetadata`` and ``DebugMetadata`` were
#   removed entirely. There is no compatibility shim — callers must migrate.
# - ``StepStatus`` / ``WorkflowRunStep`` / ``StepExecutionResponse`` no longer
#   carry ``metadata``, ``artifacts: list[...]``, ``requires_human_review``,
#   ``human_reviewed_at`` or ``human_review_approved``. They now expose:
#       * ``artifact: StepArtifact | None`` (singular pointer into a backing
#         collection — fetch the typed record via the matching resource)
#       * ``skip_reason: str | None``
#       * ``cancel_reason: str | None``
#   ``StepStatusSummary`` likewise drops ``requires_human_review`` (the status
#   value ``"waiting_for_human"`` already encodes that signal).
# - ``WorkflowArtifactOperation`` is extended with six new operations:
#   ``conditional_evaluation``, ``hil_evaluation``, ``while_loop_termination``,
#   ``api_call_invocation``, ``function_invocation``, ``webhook_invocation``.
#   Each points at a dedicated backing-collection record:
#   :class:`ConditionalEvaluation`, :class:`HilEvaluation`,
#   :class:`WhileLoopTermination`, :class:`ApiCallInvocation`,
#   :class:`FunctionInvocation`, :class:`WebhookInvocation`.
#
# Migration: callers that previously read ``step.metadata.evaluations`` /
# ``step.requires_human_review`` / ``step.human_review_approved`` should fetch
# the artifact's backing record (one of the new record types above) and read
# from there. ``step.artifact`` is the (operation, id) pointer.
# ---------------------------------------------------------------------------

# Schemas are accessed via ``workflows.blocks.get(block_id).resolved_schemas``,
# not via step artifact view data. Step executions only carry data/payload;
# user-declared block config schemas (``start_json`` / ``extract`` /
# ``function`` / ``api_call``) live on the block itself, and every other
# block's input/output schema is inferred and exposed under
# ``resolved_schemas.input_schemas`` / ``resolved_schemas.output_schemas``.


class HandlePayload(BaseModel):
    """
    Payload for a single output handle.

    Each output handle on a block produces a typed payload that can be:
    - file: A document reference (PDF, image, etc.)
    - json: Structured JSON data (extracted data, etc.)
    - text: Plain text content
    """
    type: Literal["file", "json", "text"] = Field(..., description="Type of payload")
    document: Optional[FileRef] = Field(default=None, description="For file handles: document reference")
    data: Optional[dict] = Field(default=None, description="For JSON handles: structured data")
    text: Optional[str] = Field(default=None, description="For text payloads: text content")


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
    "webhook_invocation",
]


class StepArtifact(BaseModel):
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


class ContainerContextData(BaseModel):
    """Structured context for a single container in the hierarchy."""
    container_id: str = Field(..., description="Container ID (e.g., 'while_loop-abc')")
    iteration: int = Field(..., description="Iteration index (0-based)")
    is_parallel: bool = Field(default=False, description="Whether this container represents a parallel item")
    parallel_item_index: Optional[int] = Field(default=None, description="Parallel item index if is_parallel")


class IterationContextData(BaseModel):
    """Structured container hierarchy for a step."""
    containers: List[ContainerContextData] = Field(
        default_factory=list,
        description="Container hierarchy from outermost to innermost",
    )


class ConditionEvaluationPerItem(BaseModel):
    """Per-item evaluation result for wildcard array conditions."""
    indices: List[int] = Field(
        default_factory=list,
        description="Hierarchical indices for nested arrays (e.g., [0, 2, 1] for items[0].subitems[2].field[1])",
    )
    actual: Any = Field(default=None, description="Actual value at this index")
    matched: bool = Field(default=False, description="Whether this item matched the condition")


class ConditionEvaluationSubCondition(BaseModel):
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


class ConditionEvaluationDetails(BaseModel):
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


class ConditionEvaluationResult(BaseModel):
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


class ConditionalEvaluation(BaseModel):
    """Persisted record of a conditional block's branch evaluation.

    Backing record for :data:`StepArtifact` with
    ``operation == "conditional_evaluation"``.
    """
    model_config = ConfigDict(extra="ignore")

    id: str = Field(..., description="Unique identifier")
    organization_id: str = Field(..., description="Owning organization")
    workflow_run_id: str = Field(..., description="Parent workflow run ID")
    step_id: str = Field(..., description="Producing step ID")
    evaluations: List[ConditionEvaluationResult] = Field(default_factory=list)
    selected_handles: List[str] = Field(default_factory=list)
    matched_branch_id: Optional[str] = Field(default=None)
    matched_condition_ids: List[str] = Field(default_factory=list)
    created_at: datetime.datetime = Field(..., description="When the record was created")


class HilEvaluation(BaseModel):
    """Persisted record of a HIL block's branch evaluation.

    Same evaluation core as :class:`ConditionalEvaluation`, plus human-review
    state. Backing record for :data:`StepArtifact` with
    ``operation == "hil_evaluation"``.
    """
    model_config = ConfigDict(extra="ignore")

    id: str = Field(..., description="Unique identifier")
    organization_id: str = Field(..., description="Owning organization")
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


class WhileLoopTermination(BaseModel):
    """Persisted record of why a while-loop terminated.

    Backing record for :data:`StepArtifact` with
    ``operation == "while_loop_termination"``.
    """
    model_config = ConfigDict(extra="ignore")

    id: str = Field(..., description="Unique identifier")
    organization_id: str = Field(..., description="Owning organization")
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


class ApiCallAttempt(BaseModel):
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


class ApiCallInvocation(BaseModel):
    """Persisted record of an api_call block's invocation (with retry trace).

    Backing record for :data:`StepArtifact` with
    ``operation == "api_call_invocation"``.
    """
    model_config = ConfigDict(extra="ignore")

    id: str = Field(..., description="Unique identifier")
    organization_id: str = Field(..., description="Owning organization")
    workflow_run_id: str = Field(..., description="Parent workflow run ID")
    step_id: str = Field(..., description="Producing step ID")
    attempts: List[ApiCallAttempt] = Field(
        default_factory=list,
        description="Full retry trace; final attempt holds the canonical request/response",
    )
    error: Optional[str] = Field(default=None, description="Final error after exhausted retries, if any")
    created_at: datetime.datetime = Field(..., description="When the record was created")


class FunctionInvocation(BaseModel):
    """Persisted record of a function block's invocation.

    Backing record for :data:`StepArtifact` with
    ``operation == "function_invocation"``.
    """
    model_config = ConfigDict(extra="ignore")

    id: str = Field(..., description="Unique identifier")
    organization_id: str = Field(..., description="Owning organization")
    workflow_run_id: str = Field(..., description="Parent workflow run ID")
    step_id: str = Field(..., description="Producing step ID")
    inputs: Dict[str, Any] = Field(default_factory=dict)
    output: Optional[Any] = Field(default=None)
    duration_ms: Optional[int] = Field(default=None)
    error: Optional[str] = Field(default=None)
    created_at: datetime.datetime = Field(..., description="When the record was created")


class WebhookInvocation(BaseModel):
    """Persisted record of an end-block webhook dispatch.

    Backing record for :data:`StepArtifact` with
    ``operation == "webhook_invocation"``.
    """
    model_config = ConfigDict(extra="ignore")

    id: str = Field(..., description="Unique identifier")
    organization_id: str = Field(..., description="Owning organization")
    workflow_run_id: str = Field(..., description="Parent workflow run ID")
    step_id: str = Field(..., description="Producing step ID")
    url: str = Field(...)
    request_body: Optional[Any] = Field(default=None)
    status_code: Optional[int] = Field(default=None)
    response_body: Optional[Any] = Field(default=None)
    delivered_at: Optional[datetime.datetime] = Field(default=None)
    error: Optional[str] = Field(default=None)
    created_at: datetime.datetime = Field(..., description="When the record was created")


class StepStatusSummary(BaseModel):
    """Lightweight step status embedded in :class:`WorkflowRun`.

    Carries only the row-in-a-list fields needed to display run progress.
    Full step details (handle payloads, the ``artifact`` pointer) live in the
    persisted ``WorkflowRunStep`` collection and are fetched on demand.
    """
    model_config = ConfigDict(extra="ignore")

    block_id: str = Field(..., description="Logical ID of the block")
    step_id: str = Field(..., description="Full step ID with iteration context")
    block_type: BlockType = Field(..., description="Type of the block")
    block_label: str = Field(..., description="Label of the block")
    status: StepExecutionStatus = Field(..., description="Current status")
    started_at: Optional[datetime.datetime] = Field(default=None, description="When the step started")
    completed_at: Optional[datetime.datetime] = Field(default=None, description="When the step completed")
    duration_ms: Optional[int] = Field(default=None, description="Duration in milliseconds")
    error: Optional[str] = Field(default=None, description="Error message if failed")
    error_stage: Optional[str] = Field(default=None, description="Which execution stage failed")
    error_category: Optional[str] = Field(default=None, description="Category of error for retry decisions")
    # ``requires_human_review`` is no longer carried — the status value
    # ``"waiting_for_human"`` already encodes this signal.
    loop_id: Optional[str] = Field(default=None, description="ID of the containing while_loop (if inside a loop)")
    iteration: Optional[int] = Field(default=None, description="Iteration number (0-based) if inside a loop")
    iteration_context: Optional[IterationContextData] = Field(
        default=None,
        description="Structured container hierarchy for this step",
    )
    model: Optional[str] = Field(default=None, description="LLM model used")
    cost: Optional[Dict[str, Any]] = Field(default=None, description="Cost breakdown for this step")
    tokens: Optional[Dict[str, Any]] = Field(default=None, description="Token usage for this step")

    @property
    def extracted_data(self) -> Optional[dict]:
        """StepStatusSummary intentionally omits handle payloads.

        Use ``client.workflows.runs.steps.get(run_id, block_id)`` to fetch the
        full :class:`WorkflowRunStep` (or :class:`StepExecutionResponse`) and
        read ``handle_outputs['output-json-0'].data`` from there.
        """
        return None


class StepStatus(BaseModel):
    """Full step status with all execution details.

    Mirrors the backend ``StepStatus`` document. The canonical persisted
    result of executing a step is the singular :attr:`artifact` pointer — an
    ``(operation, id)`` ref into a backing collection. There is no
    ``artifacts`` list, no ``metadata`` block, and no ``requires_human_review``
    / ``human_reviewed_at`` / ``human_review_approved`` flags: the
    ``"waiting_for_human"`` status value plus the artifact's backing record
    (e.g. :class:`HilEvaluation`) carry that information instead.
    """
    model_config = ConfigDict(extra="ignore")

    # Hint to ``retab.generate_types``: emit ``z.preprocess(null → {})``
    # around these fields' Zod schemas so the Node SDK mirrors the
    # ``_parse_handle_payloads`` model_validator below.
    __zod_preprocess_null_coerce__: ClassVar[Set[str]] = {"handle_inputs", "handle_outputs"}

    # Identity
    run_id: str = Field(default="", description="Parent workflow run ID")
    organization_id: str = Field(default="", description="Organization that owns this")

    # Core fields
    block_id: str = Field(..., description="Logical ID of the block")
    step_id: str = Field(default="", description="Full step ID with iteration context")
    block_type: BlockType = Field(..., description="Type of the block")
    block_label: str = Field(..., description="Label of the block")
    status: StepExecutionStatus = Field(..., description="Current status")
    started_at: Optional[datetime.datetime] = Field(default=None, description="When the step started")
    completed_at: Optional[datetime.datetime] = Field(default=None, description="When the step completed")
    created_at: Optional[datetime.datetime] = Field(default=None, description="When the step document was created")
    updated_at: Optional[datetime.datetime] = Field(default=None, description="When the step document was last updated")
    duration_ms: Optional[int] = Field(default=None, description="Duration in milliseconds")
    error: Optional[str] = Field(default=None, description="Error message if failed")
    error_stage: Optional[str] = Field(default=None, description="Which execution stage failed")
    error_category: Optional[str] = Field(default=None, description="Category of error for retry decisions")
    error_details: Optional[Dict[str, Any]] = Field(default=None, description="Detailed error context including stack trace")

    # Cost and token tracking
    model: Optional[str] = Field(default=None, description="LLM model used")
    cost: Optional[Dict[str, Any]] = Field(default=None, description="Cost breakdown for this step")
    tokens: Optional[Dict[str, Any]] = Field(default=None, description="Token usage for this step")
    trace_spans: Optional[List[Dict[str, Any]]] = Field(default=None, description="Nested execution trace spans")

    # Flat execution payload.
    handle_inputs: Dict[str, HandlePayload] = Field(
        default_factory=dict,
        description="Handle input payloads consumed by this step",
    )
    handle_outputs: Dict[str, HandlePayload] = Field(
        default_factory=dict,
        description="Handle output payloads produced by this step",
    )
    artifact: Optional[StepArtifact] = Field(
        default=None,
        description=(
            "Canonical result of executing this step: a persisted-ref pointer "
            "(operation + id) into the backing collection. None for steps "
            "that produce no canonical result (start/end/note/merge/sentinels)."
        ),
    )

    # Lifecycle reasons (top-level, not block-result):
    skip_reason: Optional[str] = Field(
        default=None,
        description="Reason this step was skipped (set when status='skipped')",
    )
    cancel_reason: Optional[str] = Field(
        default=None,
        description="Reason this step was cancelled (set when status='cancelled')",
    )

    # Retry tracking
    retry_count: Optional[int] = Field(default=None, description="Number of retry attempts for this step execution")

    # Loop iteration tracking
    loop_id: Optional[str] = Field(default=None, description="ID of the containing while_loop (if inside a loop)")
    iteration: Optional[int] = Field(default=None, description="Iteration number (0-based) if inside a loop")
    iteration_context: Optional[IterationContextData] = Field(
        default=None,
        description="Structured container hierarchy for this step",
    )

    @model_validator(mode="before")
    @classmethod
    def _parse_handle_payloads(cls, data: Any) -> Any:
        if isinstance(data, dict):
            for field in ("handle_outputs", "handle_inputs"):
                raw = data.get(field, ...)
                if raw is None:
                    # Backend uses non-optional dicts; coerce a literal None to an empty dict.
                    data[field] = {}
                elif isinstance(raw, dict):
                    parsed: Dict[str, Any] = {}
                    for k, v in raw.items():
                        if isinstance(v, dict):
                            try:
                                parsed[k] = HandlePayload.model_validate(v)
                            except Exception:
                                parsed[k] = v
                        else:
                            parsed[k] = v
                    data[field] = parsed
        return data

    @property
    def extracted_data(self) -> Optional[dict]:
        """Get the JSON output data from the default handle (``output-json-0``)."""
        if not self.handle_outputs:
            return None
        payload = self.handle_outputs.get("output-json-0")
        if isinstance(payload, HandlePayload) and payload.type == "json":
            return payload.data
        return None


class WorkflowRunError(Exception):
    """Raised by ``WorkflowRun.raise_for_status()`` when the run has failed."""

    def __init__(self, run: "WorkflowRun") -> None:
        self.run = run
        if run.status == "cancelled":
            msg = f"Workflow run {run.id} was cancelled"
        else:
            msg = f"Workflow run {run.id} failed"
        if run.error:
            msg += f": {run.error}"
        super().__init__(msg)


class WorkflowRun(BaseModel):
    """A stored workflow run record.

    The embedded ``steps`` list carries only :class:`StepStatusSummary` rows.
    Full step details (handle payloads, the ``artifact`` pointer) are stored
    separately in the persisted step collection — fetch them via
    ``client.workflows.runs.steps.get(...)``.
    """
    model_config = ConfigDict(extra="ignore")

    id: str = Field(..., description="Unique ID for this run")
    workflow_id: str = Field(..., description="ID of the workflow that was run")
    workflow_name: str = Field(..., description="Name of the workflow at time of execution")
    organization_id: str = Field(..., description="Organization that owns this run")
    status: Literal["pending", "running", "completed", "error", "waiting_for_human", "cancelled"] = Field(default="pending", description="Overall status")
    started_at: datetime.datetime = Field(..., description="When the workflow started")
    completed_at: Optional[datetime.datetime] = Field(default=None, description="When the workflow completed")
    duration_ms: Optional[int] = Field(default=None, description="Total duration in milliseconds")
    steps: List[StepStatusSummary] = Field(default_factory=list, description="Lightweight summary of each step")
    input_documents: Optional[Dict[str, FileRef]] = Field(default=None, description="Start block ID -> input document reference")
    final_outputs: Optional[dict] = Field(default=None, description="Final outputs from terminal blocks")
    error: Optional[str] = Field(default=None, description="Error message if workflow failed")
    created_at: datetime.datetime = Field(..., description="When the run was created")
    updated_at: datetime.datetime = Field(..., description="When the run was last updated")
    waiting_for_block_ids: List[str] = Field(default_factory=list, description="Block IDs that are waiting for human review")
    pending_block_outputs: Optional[dict] = Field(default=None, description="Serialized block outputs to resume from")
    config_snapshot_id: Optional[str] = Field(default=None, description="ID of the config snapshot used for this run")
    trigger_type: Optional[str] = Field(default=None, description="How the run was triggered (manual, api, schedule, webhook, email, restart)")
    trigger_email: Optional[str] = Field(default=None, description="Email address that triggered the run (for email triggers)")
    execution_phase: Optional[str] = Field(default=None, description="Current execution phase (created, dispatching, started, running)")
    input_json_data: Optional[Dict[str, dict]] = Field(default=None, description="Start JSON block ID -> input JSON data")
    error_details: Optional[dict] = Field(default=None, description="Detailed error information including stack trace and context")
    cost_summary: Optional[dict] = Field(default=None, description="Aggregate cost and token usage for the run")
    human_waiting_duration_ms: int = Field(default=0, description="Total time spent waiting for human review in milliseconds")

    def raise_for_status(self) -> None:
        """Raise :class:`WorkflowRunError` if the run did not succeed.

        Modelled after ``httpx.Response.raise_for_status()``.
        """
        if self.status in {"error", "cancelled"}:
            raise WorkflowRunError(self)

    @property
    def is_terminal(self) -> bool:
        """Whether the run has reached a terminal state (``completed``, ``error``, ``cancelled``)."""
        return self.status in TERMINAL_WORKFLOW_RUN_STATUSES



class Workflow(BaseModel):
    """A stored workflow record."""
    model_config = ConfigDict(extra="ignore")

    class Published(BaseModel):
        snapshot_id: Optional[str] = Field(default=None, description="Published snapshot ID")
        published_at: Optional[datetime.datetime] = Field(default=None, description="When the workflow was last published")

    class EmailTrigger(BaseModel):
        allowed_senders: List[str] = Field(default_factory=list, description="Allowed sender email addresses")
        allowed_domains: List[str] = Field(default_factory=list, description="Allowed sender email domains")

    id: str = Field(..., description="Unique ID for this workflow")
    name: str = Field(default="Untitled Workflow", description="Workflow name")
    description: str = Field(default="", description="Workflow description")
    organization_id: Optional[str] = Field(default=None, description="Organization that owns this workflow")
    published: Optional[Published] = Field(default=None, description="Published workflow metadata")
    email_trigger: EmailTrigger = Field(default_factory=EmailTrigger, description="Email trigger allowlist policy")
    created_at: datetime.datetime = Field(..., description="When the workflow was created")
    updated_at: datetime.datetime = Field(..., description="When the workflow was last updated")

    @property
    def published_snapshot_id(self) -> Optional[str]:
        return self.published.snapshot_id if self.published is not None else None

    @property
    def published_at(self) -> Optional[datetime.datetime]:
        return self.published.published_at if self.published is not None else None

    @property
    def email_senders_whitelist(self) -> List[str]:
        return list(self.email_trigger.allowed_senders)

    @property
    def email_domains_whitelist(self) -> List[str]:
        return list(self.email_trigger.allowed_domains)


# ---------------------------------------------------------------------------
# Type aliases
# ---------------------------------------------------------------------------

WorkflowRunStatus = Literal["pending", "running", "completed", "error", "waiting_for_human", "cancelled"]
WorkflowRunTriggerType = Literal["manual", "api", "schedule", "webhook", "email", "restart"]

TERMINAL_WORKFLOW_RUN_STATUSES: tuple[str, ...] = ("completed", "error", "cancelled")


class StepExecutionResponse(BaseModel):
    """Step status and handle data for a specific step in a workflow run."""
    model_config = ConfigDict(extra="ignore")

    # Hint to ``retab.generate_types``: emit ``z.preprocess(null → {})``
    # around these fields so the Node SDK mirrors the
    # ``_parse_handle_payloads`` model_validator below.
    __zod_preprocess_null_coerce__: ClassVar[Set[str]] = {"handle_inputs", "handle_outputs"}

    block_id: str = Field(..., description="ID of the block")
    block_type: str = Field(..., description="Type of the block")
    block_label: str = Field(..., description="Label of the block")
    status: str = Field(..., description="Step status")
    error: Optional[str] = Field(default=None, description="Error message if failed")
    artifact: Optional[StepArtifact] = Field(
        default=None,
        description=(
            "Canonical persisted-ref pointer (operation + id) for this step. "
            "Fetch the typed backing record via the matching resource client "
            "(e.g. ``client.extractions.get`` for ``operation == 'extraction'``)."
        ),
    )
    skip_reason: Optional[str] = Field(
        default=None,
        description="Reason this step was skipped (set when status='skipped')",
    )
    cancel_reason: Optional[str] = Field(
        default=None,
        description="Reason this step was cancelled (set when status='cancelled')",
    )
    handle_outputs: Dict[str, HandlePayload] = Field(
        default_factory=dict,
        description="Handle outputs keyed by handle ID",
    )
    handle_inputs: Dict[str, HandlePayload] = Field(
        default_factory=dict,
        description="Handle inputs keyed by handle ID (what this block received)",
    )

    @model_validator(mode="before")
    @classmethod
    def _parse_handle_payloads(cls, data: Any) -> Any:
        if isinstance(data, dict):
            for field in ("handle_outputs", "handle_inputs"):
                raw = data.get(field, ...)
                if raw is None:
                    # Backend uses non-optional dicts; coerce a literal None to an empty dict.
                    data[field] = {}
                elif isinstance(raw, dict):
                    parsed: Dict[str, Any] = {}
                    for k, v in raw.items():
                        if isinstance(v, dict):
                            try:
                                parsed[k] = HandlePayload.model_validate(v)
                            except Exception:
                                parsed[k] = v
                        else:
                            parsed[k] = v
                    data[field] = parsed
        return data

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


class WorkflowRunStep(BaseModel):
    """Persisted public step document returned by list workflow run steps.

    Mirrors the backend ``StepStatus`` shape — ``handle_inputs``,
    ``handle_outputs`` and the singular :attr:`artifact` pointer are flat
    fields. There is no ``artifacts`` list, no ``metadata`` block and no
    ``requires_human_review`` / ``human_reviewed_at`` /
    ``human_review_approved`` flags. To inspect HIL or evaluation state,
    fetch the artifact's backing record (e.g. :class:`HilEvaluation`,
    :class:`ConditionalEvaluation`) instead.
    """
    model_config = ConfigDict(extra="ignore")

    # Hint to ``retab.generate_types``: emit ``z.preprocess(null → {})``
    # around these fields so the Node SDK mirrors the
    # ``_parse_handle_payloads`` model_validator below.
    __zod_preprocess_null_coerce__: ClassVar[Set[str]] = {"handle_inputs", "handle_outputs"}

    run_id: str = Field(..., description="Parent workflow run ID")
    organization_id: str = Field(..., description="Organization that owns this run")
    block_id: str = Field(..., description="Logical ID of the block")
    step_id: str = Field(..., description="Stored step ID")
    block_type: str = Field(..., description="Type of the block")
    block_label: str = Field(..., description="Label of the block")
    status: str = Field(..., description="Step status")
    artifact: Optional[StepArtifact] = Field(
        default=None,
        description=(
            "Canonical persisted-ref pointer (operation + id) for this step. "
            "None for steps that produce no canonical result (start/end/note/"
            "merge/sentinels)."
        ),
    )
    skip_reason: Optional[str] = Field(
        default=None,
        description="Reason this step was skipped (set when status='skipped')",
    )
    cancel_reason: Optional[str] = Field(
        default=None,
        description="Reason this step was cancelled (set when status='cancelled')",
    )
    started_at: Optional[datetime.datetime] = Field(default=None, description="When the step started")
    completed_at: Optional[datetime.datetime] = Field(default=None, description="When the step completed")
    duration_ms: Optional[int] = Field(default=None, description="Duration in milliseconds")
    error: Optional[str] = Field(default=None, description="Error message if failed")
    error_stage: Optional[str] = Field(default=None, description="Which execution stage failed")
    error_category: Optional[str] = Field(default=None, description="Category of error for retry decisions")
    error_details: Optional[Dict[str, Any]] = Field(default=None, description="Detailed error context")
    handle_outputs: Dict[str, HandlePayload] = Field(
        default_factory=dict,
        description="Handle outputs keyed by handle ID",
    )
    handle_inputs: Dict[str, HandlePayload] = Field(
        default_factory=dict,
        description="Handle inputs keyed by handle ID",
    )
    model: Optional[str] = Field(default=None, description="LLM model used")
    cost: Optional[Dict[str, Any]] = Field(default=None, description="Cost breakdown for this step")
    tokens: Optional[Dict[str, Any]] = Field(default=None, description="Token usage for this step")
    trace_spans: Optional[List[Dict[str, Any]]] = Field(default=None, description="Nested execution trace spans")
    retry_count: Optional[int] = Field(default=None, description="Retry count for this step")
    loop_id: Optional[str] = Field(default=None, description="Containing while_loop ID")
    iteration: Optional[int] = Field(default=None, description="Loop iteration number")
    iteration_context: Optional[IterationContextData] = Field(
        default=None,
        description="Structured container hierarchy for this step",
    )
    created_at: Optional[datetime.datetime] = Field(default=None, description="When the step document was created")
    updated_at: Optional[datetime.datetime] = Field(default=None, description="When the step document was last updated")

    @model_validator(mode="before")
    @classmethod
    def _parse_handle_payloads(cls, data: Any) -> Any:
        if isinstance(data, dict):
            for field in ("handle_outputs", "handle_inputs"):
                raw = data.get(field, ...)
                if raw is None:
                    # Backend uses non-optional dicts; coerce a literal None to an empty dict.
                    data[field] = {}
                elif isinstance(raw, dict):
                    parsed: Dict[str, Any] = {}
                    for k, v in raw.items():
                        if isinstance(v, dict):
                            try:
                                parsed[k] = HandlePayload.model_validate(v)
                            except Exception:
                                parsed[k] = v
                        else:
                            parsed[k] = v
                    data[field] = parsed
        return data

    @property
    def extracted_data(self) -> Optional[dict]:
        """Get the JSON output data from the default handle (``output-json-0``)."""
        if not self.handle_outputs:
            return None
        payload = self.handle_outputs.get("output-json-0")
        if isinstance(payload, HandlePayload) and payload.type == "json":
            return payload.data
        return None


class StepExecutionsBatchResponse(BaseModel):
    """Response for batch step execution retrieval, keyed by block ID."""
    executions: Dict[str, StepExecutionResponse] = Field(
        default_factory=dict,
        description="Step execution records keyed by block ID (missing steps are omitted)",
    )


# ---------------------------------------------------------------------------
# Cancel / Restart / HIL decision response types
# ---------------------------------------------------------------------------

class CancelWorkflowResponse(BaseModel):
    """Response from cancelling a workflow run."""
    run: WorkflowRun
    cancellation_status: Literal["cancelled", "cancellation_requested", "cancellation_failed"] = Field(
        default="cancellation_requested",
        description="Cancellation delivery state",
    )


class HILDecisionResource(BaseModel):
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


class SubmitHILDecisionResponse(BaseModel):
    """Response from submitting a HIL decision."""
    submission_status: Literal["accepted", "already_received", "already_applied"] = Field(
        ...,
        description="Decision submission lifecycle status",
    )
    decision: HILDecisionResource = Field(
        ...,
        description="Temporal-owned HIL decision state for the block",
    )


class RunCountsResponse(BaseModel):
    """Run counts grouped by status."""
    total: int = 0
    completed: int = 0
    running: int = 0
    error: int = 0
    pending: int = 0
    waiting_for_human: int = 0
    cancelled: int = 0


class ExecutionOrderResponse(BaseModel):
    """DAG-ordered step IDs for a workflow run."""
    run_id: str = Field(..., description="Workflow run ID")
    ordered_step_ids: List[str] = Field(default_factory=list, description="Step IDs in DAG execution order")


class DocumentSignedUrlResponse(BaseModel):
    """Signed URL for downloading a document from a run step."""
    signed_url: str = Field(..., description="Signed download URL")
    filename: str = Field(..., description="Original filename")
    mime_type: Optional[str] = Field(default=None, description="MIME type")


class ExportResponse(BaseModel):
    """Export payload containing CSV data."""
    csv_data: str = Field(..., description="CSV content as string")
    rows: int = Field(..., description="Number of data rows")
    columns: int = Field(..., description="Number of columns")


# ---------------------------------------------------------------------------
# Workflow graph request models
# ---------------------------------------------------------------------------


class WorkflowBlockCreateRequest(BaseModel):
    """Typed request payload for creating a workflow block."""
    model_config = ConfigDict(extra="forbid")

    id: str
    type: str
    label: str = ""
    position_x: float = 0
    position_y: float = 0
    width: Optional[float] = None
    height: Optional[float] = None
    config: Optional[dict] = None
    parent_id: Optional[str] = None


class WorkflowBlockUpdateRequest(BaseModel):
    """Typed request payload for updating a workflow block."""
    model_config = ConfigDict(extra="forbid")

    block_id: str
    label: Optional[str] = None
    position_x: Optional[float] = None
    position_y: Optional[float] = None
    width: Optional[float] = None
    height: Optional[float] = None
    config: Optional[dict] = None
    parent_id: Optional[str] = None


class WorkflowEdgeCreateRequest(BaseModel):
    """Typed request payload for creating a workflow edge."""
    model_config = ConfigDict(extra="forbid")

    id: str
    source_block: str
    target_block: str
    source_handle: Optional[str] = None
    target_handle: Optional[str] = None


# ---------------------------------------------------------------------------
# Workflow graph response models (blocks, edges, subflows)
# ---------------------------------------------------------------------------


class ResolvedSchemas(BaseModel):
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


class WorkflowBlock(BaseModel):
    """A block in a workflow graph."""
    model_config = ConfigDict(extra="ignore")

    id: str = Field(..., description="Block ID")
    workflow_id: str = Field(..., description="Parent workflow ID")
    organization_id: str = Field(..., description="Organization ID")
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


class WorkflowEdgeDoc(BaseModel):
    """A persisted edge document connecting two blocks in a workflow graph."""
    model_config = ConfigDict(extra="ignore")

    id: str = Field(..., description="Edge ID")
    workflow_id: str = Field(..., description="Parent workflow ID")
    organization_id: str = Field(..., description="Organization ID")
    draft_version: Optional[str] = Field(default=None, description="Draft version for the live entity")
    source_block: str = Field(..., description="Source block ID")
    target_block: str = Field(..., description="Target block ID")
    source_handle: Optional[str] = Field(default=None, description="Output handle on source block")
    target_handle: Optional[str] = Field(default=None, description="Input handle on target block")
    updated_at: Optional[datetime.datetime] = Field(default=None, description="Last updated timestamp")


class WorkflowSubflow(BaseModel):
    """A subflow container (loop or parallel execution group)."""
    model_config = ConfigDict(extra="ignore")

    id: str = Field(..., description="Subflow ID")
    workflow_id: str = Field(..., description="Parent workflow ID")
    organization_id: str = Field(..., description="Organization ID")
    draft_version: Optional[str] = Field(default=None, description="Draft version for the live entity")
    type: str = Field(..., description="Subflow type (while, parallel)")
    label: str = Field(default="", description="Display label")
    position_x: float = Field(default=0, description="X position on canvas")
    position_y: float = Field(default=0, description="Y position on canvas")
    width: float = Field(default=400, description="Container width")
    height: float = Field(default=300, description="Container height")
    config: Optional[dict] = Field(default=None, description="Subflow configuration")
    child_block_ids: List[str] = Field(default_factory=list, description="Block IDs inside this subflow")
    updated_at: Optional[datetime.datetime] = Field(default=None, description="Last updated timestamp")


class WorkflowWithEntities(BaseModel):
    """Complete workflow with its graph structure (blocks, edges, subflows)."""
    model_config = ConfigDict(extra="ignore")

    workflow: Workflow
    blocks: List[WorkflowBlock] = Field(default_factory=list)
    edges: List[WorkflowEdgeDoc] = Field(default_factory=list)
    subflows: List[WorkflowSubflow] = Field(default_factory=list)

    @property
    def start_blocks(self) -> List[WorkflowBlock]:
        """Document input start blocks."""
        return [b for b in self.blocks if b.type == "start"]

    @property
    def start_json_blocks(self) -> List[WorkflowBlock]:
        """JSON input start blocks."""
        return [b for b in self.blocks if b.type == "start_json"]
