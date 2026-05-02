import datetime
from typing import Any, Dict, List, Literal, Optional

from pydantic import BaseModel, Field, ConfigDict, model_validator

from retab.types.mime import FileRef

# Schemas are accessed via ``workflows.blocks.get(block_id).resolved_schemas``, not
# via step artifact view data. Step executions only carry data/payload; user-declared block
# config schemas (``start_json`` / ``extract`` / ``function`` / ``api_call``) live
# on the block itself, and every other block's input/output schema is inferred and
# exposed under ``resolved_schemas.input_schemas`` / ``resolved_schemas.output_schemas``.


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
WorkflowArtifactOperation = str


class StepArtifactRef(BaseModel):
    """Canonical persisted resource produced by a workflow step."""

    operation: WorkflowArtifactOperation = Field(
        ...,
        description="Persisted resource operation to use for lookup",
    )
    id: str = Field(..., description="Persisted resource identifier")


class StepArtifactView(BaseModel):
    """Block-specific artifact view data for rendering step results."""

    block_type: str = Field(..., description="Workflow block type that produced the view")
    artifact: Optional[StepArtifactRef] = Field(
        default=None,
        description="Primary artifact backing this view",
    )
    artifacts: List[StepArtifactRef] = Field(
        default_factory=list,
        description="All artifacts backing this view, primary first",
    )
    data: Optional[Any] = Field(
        default=None,
        description="Block-specific render data resolved from the step artifact",
    )
    source_handle_id: Optional[str] = Field(
        default=None,
        description="Handle that supplied the render data when available",
    )


class StepStatus(BaseModel):
    """Status of a single step in workflow execution"""
    block_id: str = Field(..., description="ID of the block")
    block_type: BlockType = Field(..., description="Type of the block")
    block_label: str = Field(..., description="Label of the block")
    status: StepExecutionStatus = Field(..., description="Current status")
    started_at: Optional[datetime.datetime] = Field(default=None, description="When the step started")
    completed_at: Optional[datetime.datetime] = Field(default=None, description="When the step completed")
    duration_ms: Optional[int] = Field(default=None, description="Duration in milliseconds")
    error: Optional[str] = Field(default=None, description="Error message if failed")
    artifact: Optional[StepArtifactRef] = Field(
        default=None,
        description="Canonical persisted resource produced by this step, if any",
    )
    artifacts: List[StepArtifactRef] = Field(
        default_factory=list,
        description="All persisted resources produced by this step, primary first",
    )
    artifact_view: Optional[StepArtifactView] = Field(
        default=None,
        description="Block-specific artifact view model for result rendering",
    )
    handle_outputs: Optional[Dict[str, HandlePayload]] = Field(
        default=None,
        description="Output payloads keyed by handle ID (e.g., 'output-file-0', 'output-json-0')"
    )
    requires_human_review: Optional[bool] = Field(default=None, description="Whether this step requires human review")
    human_reviewed_at: Optional[datetime.datetime] = Field(default=None, description="When human review was completed")
    human_review_approved: Optional[bool] = Field(default=None, description="Whether human approved or rejected")

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
    """A stored workflow run record"""
    model_config = ConfigDict(extra="ignore")

    id: str = Field(..., description="Unique ID for this run")
    workflow_id: str = Field(..., description="ID of the workflow that was run")
    workflow_name: str = Field(..., description="Name of the workflow at time of execution")
    organization_id: str = Field(..., description="Organization that owns this run")
    status: Literal["pending", "running", "completed", "error", "waiting_for_human", "cancelled"] = Field(default="pending", description="Overall status")
    started_at: datetime.datetime = Field(..., description="When the workflow started")
    completed_at: Optional[datetime.datetime] = Field(default=None, description="When the workflow completed")
    duration_ms: Optional[int] = Field(default=None, description="Total duration in milliseconds")
    steps: List[StepStatus] = Field(default_factory=list, description="Status of each step")
    input_documents: Optional[Dict[str, FileRef]] = Field(default=None, description="Start block ID -> input document reference")
    final_outputs: Optional[dict] = Field(default=None, description="Final outputs from end blocks")
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
    block_id: str = Field(..., description="ID of the block")
    block_type: str = Field(..., description="Type of the block")
    block_label: str = Field(..., description="Label of the block")
    status: str = Field(..., description="Step status")
    error: Optional[str] = Field(default=None, description="Error message if failed")
    artifact: Optional[StepArtifactRef] = Field(
        default=None,
        description="Canonical persisted resource produced by this step, if any",
    )
    artifacts: List[StepArtifactRef] = Field(
        default_factory=list,
        description="All persisted resources produced by this step, primary first",
    )
    artifact_view: Optional[StepArtifactView] = Field(
        default=None,
        description="Block-specific artifact view model for result rendering",
    )
    handle_outputs: Optional[Dict[str, HandlePayload]] = Field(default=None, description="Handle outputs keyed by handle ID")
    handle_inputs: Optional[Dict[str, HandlePayload]] = Field(default=None, description="Handle inputs keyed by handle ID (what this block received)")
    metadata: Optional[Dict[str, Any]] = Field(
        default=None,
        description="Execution metadata for routing, review, container state, consensus, and diagnostics",
    )

    @model_validator(mode="before")
    @classmethod
    def _parse_handle_payloads(cls, data: Any) -> Any:
        if isinstance(data, dict):
            for field in ("handle_outputs", "handle_inputs"):
                raw = data.get(field)
                if isinstance(raw, dict):
                    parsed = {}
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

        Most extract/formula blocks emit JSON on ``output-json-0``.
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
    """Persisted public step document returned by list workflow run steps."""
    model_config = ConfigDict(extra="ignore")

    run_id: str = Field(..., description="Parent workflow run ID")
    organization_id: str = Field(..., description="Organization that owns this run")
    block_id: str = Field(..., description="Logical ID of the block")
    step_id: str = Field(..., description="Stored step ID")
    block_type: str = Field(..., description="Type of the block")
    block_label: str = Field(..., description="Label of the block")
    status: str = Field(..., description="Step status")
    artifact: Optional[StepArtifactRef] = Field(
        default=None,
        description="Canonical persisted resource produced by this step, if any",
    )
    artifacts: List[StepArtifactRef] = Field(
        default_factory=list,
        description="All persisted resources produced by this step, primary first",
    )
    artifact_view: Optional[StepArtifactView] = Field(
        default=None,
        description="Block-specific artifact view model for result rendering",
    )
    started_at: Optional[datetime.datetime] = Field(default=None, description="When the step started")
    completed_at: Optional[datetime.datetime] = Field(default=None, description="When the step completed")
    duration_ms: Optional[int] = Field(default=None, description="Duration in milliseconds")
    error: Optional[str] = Field(default=None, description="Error message if failed")
    handle_outputs: Optional[Dict[str, HandlePayload]] = Field(default=None, description="Handle outputs keyed by handle ID")
    handle_inputs: Optional[Dict[str, HandlePayload]] = Field(default=None, description="Handle inputs keyed by handle ID")
    requires_human_review: Optional[bool] = Field(default=None, description="Whether this step requires human review")
    human_reviewed_at: Optional[datetime.datetime] = Field(default=None, description="When human review completed")
    human_review_approved: Optional[bool] = Field(default=None, description="Whether human approved or rejected")
    retry_count: Optional[int] = Field(default=None, description="Retry count for this step")
    loop_id: Optional[str] = Field(default=None, description="Containing while_loop ID")
    iteration: Optional[int] = Field(default=None, description="Loop iteration number")
    created_at: Optional[datetime.datetime] = Field(default=None, description="When the step document was created")
    updated_at: Optional[datetime.datetime] = Field(default=None, description="When the step document was last updated")

    @model_validator(mode="before")
    @classmethod
    def _parse_handle_payloads(cls, data: Any) -> Any:
        if isinstance(data, dict):
            for field in ("handle_outputs", "handle_inputs"):
                raw = data.get(field)
                if isinstance(raw, dict):
                    parsed = {}
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
