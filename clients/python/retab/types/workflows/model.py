import datetime
from typing import Any, Dict, List, Literal, Optional
from typing import TypeAlias

from pydantic import BaseModel, Field, ConfigDict, model_validator

from retab.types.mime import BaseMIMEData
from retab.types.documents.split import SplitResponse


class HandlePayload(BaseModel):
    """
    Payload for a single output handle.

    Each output handle on a node produces a typed payload that can be:
    - file: A document reference (PDF, image, etc.)
    - json: Structured JSON data (extracted data, etc.)
    - text: Plain text content
    """
    type: Literal["file", "json", "text"] = Field(..., description="Type of payload")
    document: Optional[BaseMIMEData] = Field(default=None, description="For file handles: document reference")
    data: Optional[dict] = Field(default=None, description="For JSON handles: structured data")
    text: Optional[str] = Field(default=None, description="For text payloads: text content")


# Workflow run payloads can contain newer backend node types before the SDK is
# regenerated. Keep runtime validation permissive so informational step metadata
# does not break run parsing.
NodeType = str
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


class StepStatus(BaseModel):
    """Status of a single step in workflow execution"""
    node_id: str = Field(..., description="ID of the node")
    node_type: NodeType = Field(..., description="Type of the node")
    node_label: str = Field(..., description="Label of the node")
    status: StepExecutionStatus = Field(..., description="Current status")
    started_at: Optional[datetime.datetime] = Field(default=None, description="When the step started")
    completed_at: Optional[datetime.datetime] = Field(default=None, description="When the step completed")
    duration_ms: Optional[int] = Field(default=None, description="Duration in milliseconds")
    error: Optional[str] = Field(default=None, description="Error message if failed")
    handle_outputs: Optional[Dict[str, HandlePayload]] = Field(
        default=None,
        description="Output payloads keyed by handle ID (e.g., 'output-file-0', 'output-json-0')"
    )
    input_document: Optional[BaseMIMEData] = Field(default=None, description="Reference to input document")
    output_document: Optional[BaseMIMEData] = Field(default=None, description="Reference to output document")
    split_documents: Optional[Dict[str, BaseMIMEData]] = Field(default=None, description="For split nodes: subdocument -> document reference")
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
    input_documents: Optional[Dict[str, BaseMIMEData]] = Field(default=None, description="Start node ID -> input document reference")
    final_outputs: Optional[dict] = Field(default=None, description="Final outputs from end nodes")
    error: Optional[str] = Field(default=None, description="Error message if workflow failed")
    created_at: datetime.datetime = Field(..., description="When the run was created")
    updated_at: datetime.datetime = Field(..., description="When the run was last updated")
    waiting_for_node_ids: List[str] = Field(default_factory=list, description="Node IDs that are waiting for human review")
    pending_node_outputs: Optional[dict] = Field(default=None, description="Serialized node outputs to resume from")
    config_snapshot_id: Optional[str] = Field(default=None, description="ID of the config snapshot used for this run")
    trigger_type: Optional[str] = Field(default=None, description="How the run was triggered (manual, api, schedule, webhook, email, restart)")
    trigger_email: Optional[str] = Field(default=None, description="Email address that triggered the run (for email triggers)")
    execution_phase: Optional[str] = Field(default=None, description="Current execution phase (created, dispatching, started, running)")
    input_json_data: Optional[Dict[str, dict]] = Field(default=None, description="Start JSON node ID -> input JSON data")
    error_details: Optional[dict] = Field(default=None, description="Detailed error information including stack trace and context")
    cost_summary: Optional[dict] = Field(default=None, description="Aggregate cost and token usage for the run")
    human_waiting_duration_ms: int = Field(default=0, description="Total time spent waiting for human review in milliseconds")

    def raise_for_status(self) -> None:
        """Raise :class:`WorkflowRunError` if the run did not succeed.

        Modelled after ``httpx.Response.raise_for_status()``.
        """
        if self.status in {"error", "cancelled"}:
            raise WorkflowRunError(self)



class Workflow(BaseModel):
    """A stored workflow record."""
    model_config = ConfigDict(extra="ignore")

    id: str = Field(..., description="Unique ID for this workflow")
    name: str = Field(default="Untitled Workflow", description="Workflow name")
    description: str = Field(default="", description="Workflow description")
    is_published: bool = Field(default=False, description="Whether the workflow has a published snapshot")
    published_snapshot_id: Optional[str] = Field(default=None, description="Published snapshot ID")
    published_at: Optional[datetime.datetime] = Field(default=None, description="When the workflow was last published")
    draft_version: Optional[str] = Field(default=None, description="Current draft version")
    organization_id: Optional[str] = Field(default=None, description="Organization that owns this workflow")
    email_senders_whitelist: List[str] = Field(default_factory=list, description="Allowed sender email addresses")
    email_domains_whitelist: List[str] = Field(default_factory=list, description="Allowed sender email domains")
    created_at: datetime.datetime = Field(..., description="When the workflow was created")
    updated_at: datetime.datetime = Field(..., description="When the workflow was last updated")


# ---------------------------------------------------------------------------
# Type aliases
# ---------------------------------------------------------------------------

WorkflowRunStatus = Literal["pending", "running", "completed", "error", "waiting_for_human", "cancelled"]
WorkflowRunTriggerType = Literal["manual", "api", "schedule", "webhook", "email", "restart"]

TERMINAL_WORKFLOW_RUN_STATUSES: tuple[str, ...] = ("completed", "error", "cancelled")


# ---------------------------------------------------------------------------
# Step output types (from GET /v1/workflows/runs/{run_id}/steps/{node_id})
# ---------------------------------------------------------------------------


class StartDocumentStepOutput(BaseModel):
    """Metadata persisted for document-based start nodes."""

    filename: str = Field(..., description="Original filename")
    mime_type: str = Field(..., description="Document MIME type")
    id: Optional[str] = Field(default=None, description="Stored file identifier")


class SkippedStepOutput(BaseModel):
    skipped: bool = Field(default=True, description="Whether this node was skipped")
    reason: str = Field(..., description="Reason why the node was skipped")
    missing_input: str = Field(..., description="The type of input that was missing")


class ExtractStepOutput(BaseModel):
    extracted_data: Dict[str, Any] = Field(..., description="The extracted structured data")
    likelihoods: Optional[Dict[str, Any]] = Field(default=None, description="Field confidence scores")
    extraction_id: Optional[str] = Field(default=None, description="Extraction ID")
    json_schema: Optional[Dict[str, Any]] = Field(default=None, description="JSON schema used for extraction")


class ParseStepOutput(BaseModel):
    pages: List[str] = Field(..., description="Text content for each parsed page")


class MergePdfStepOutput(BaseModel):
    message: str = Field(..., description="Status message about the merge operation")
    document_count: int = Field(..., description="Number of documents merged")


class MergeDictsStepOutput(BaseModel):
    message: str = Field(..., description="Status message about the merge operation")
    field_count: int = Field(..., description="Number of fields in the merged output")
    fields: List[str] = Field(default_factory=list, description="Names of the merged fields")
    extracted_data: Dict[str, Any] = Field(default_factory=dict, description="Merged JSON data")
    merged_schema: Optional[Dict[str, Any]] = Field(default=None, description="Combined schema from merged inputs")


class ClassifierStepOutput(BaseModel):
    category: str = Field(..., description="The classified category")
    reasoning: Optional[str] = Field(default=None, description="Model reasoning")
    all_categories: List[str] = Field(default_factory=list, description="All available categories")
    confidence: Optional[float] = Field(default=None, description="Classification confidence")


class FormulaStepOutput(BaseModel):
    message: str = Field(..., description="Status message about the computation")
    computations: List[str] = Field(default_factory=list, description="List of computed field names")
    extracted_data: Dict[str, Any] = Field(default_factory=dict, description="Enriched JSON data")
    extraction_id: Optional[str] = Field(default=None, description="Associated extraction ID")
    computation_spec: Optional[Dict[str, Any]] = Field(default=None, description="Computation spec used")
    json_schema: Optional[Dict[str, Any]] = Field(default=None, description="Output schema")


class HILStepOutput(BaseModel):
    message: str = Field(..., description="Status message about the HIL node")
    requires_review: bool = Field(default=True, description="Whether human review is required")
    extracted_data: Optional[Dict[str, Any]] = Field(default=None, description="Data awaiting or completed review")
    extraction_id: Optional[str] = Field(default=None, description="Associated extraction ID")
    computation_spec: Optional[Dict[str, Any]] = Field(default=None, description="Computation spec")
    human_modified: bool = Field(default=False, description="Whether the data was modified by human review")
    json_schema: Optional[Dict[str, Any]] = Field(default=None, description="Schema for review rendering")


class ConditionalStepOutput(BaseModel):
    message: str = Field(..., description="Status message about the conditional evaluation")
    selected_branch: Optional[str] = Field(default=None, description="Selected branch name")
    matched_condition_id: Optional[str] = Field(default=None, description="Matched condition ID")
    branch_outputs: Dict[str, Any] = Field(default_factory=dict, description="Data routed to each branch")
    evaluations: List[Dict[str, Any]] = Field(default_factory=list, description="Individual evaluation results")
    extracted_data: Optional[Dict[str, Any]] = Field(default=None, description="Evaluated input data")
    extraction_id: Optional[str] = Field(default=None, description="Associated extraction ID")
    computation_spec: Optional[Dict[str, Any]] = Field(default=None, description="Passed-through computation spec")


class ConditionalCheckStepOutput(BaseModel):
    message: str = Field(..., description="Status message about the conditional check")
    requires_review: bool = Field(..., description="Whether review is required")
    matched_conditions: List[str] = Field(default_factory=list, description="Matched condition IDs")
    evaluations: List[Dict[str, Any]] = Field(default_factory=list, description="Individual evaluation results")
    extracted_data: Optional[Dict[str, Any]] = Field(default=None, description="Transmitted data")
    conditional_data: Optional[Dict[str, Any]] = Field(default=None, description="Conditional data used for evaluation")
    extraction_id: Optional[str] = Field(default=None, description="Associated extraction ID")
    computation_spec: Optional[Dict[str, Any]] = Field(default=None, description="Passed-through computation spec")
    human_modified: bool = Field(default=False, description="Whether data was modified by human review")
    json_schema: Optional[Dict[str, Any]] = Field(default=None, description="Schema for review rendering")


class APICallStepOutput(BaseModel):
    message: str = Field(..., description="Status message about the API call")
    url: str = Field(..., description="URL that was called")
    method: str = Field(..., description="HTTP method used")
    status_code: int = Field(..., description="HTTP response status code")
    response_data: Optional[Dict[str, Any]] = Field(default=None, description="Parsed JSON response data")
    response_text: Optional[str] = Field(default=None, description="Raw response text")
    request_body: Optional[str] = Field(default=None, description="Request body that was sent")
    json_schema: Optional[Dict[str, Any]] = Field(default=None, description="Output schema for downstream nodes")
    error: Optional[str] = Field(default=None, description="Error message if the request failed")


class FunctionStepOutput(BaseModel):
    message: str = Field(..., description="Status message about the function execution")
    execution_time_ms: Optional[float] = Field(default=None, description="Execution time in milliseconds")
    stdout: Optional[str] = Field(default=None, description="Standard output from the sandbox")
    stderr: Optional[str] = Field(default=None, description="Standard error from the sandbox")
    error: Optional[str] = Field(default=None, description="Error message if execution failed")
    traceback_str: Optional[str] = Field(default=None, description="Python traceback if execution failed")
    json_schema: Optional[Dict[str, Any]] = Field(default=None, description="Output schema for downstream nodes")


class EditStepOutput(BaseModel):
    mode: Literal["template", "document"] = Field(..., description="Edit execution mode")
    form_fields: List[Dict[str, Any]] = Field(default_factory=list, description="Form fields produced by the edit workflow")
    filled_count: int = Field(..., description="Number of filled fields")
    total_fields: int = Field(..., description="Total number of fields")
    template_id: Optional[str] = Field(default=None, description="Template ID used in template mode")


class EndStepOutput(BaseModel):
    message: str = Field(..., description="Status message for the end node")
    webhook_sent: bool = Field(default=False, description="Whether a webhook was attempted")
    webhook_status_code: Optional[int] = Field(default=None, description="HTTP status code from webhook response")
    webhook_response_text: Optional[str] = Field(default=None, description="Response body text from the webhook")
    webhook_response_headers: Optional[Dict[str, str]] = Field(default=None, description="Webhook response headers")
    webhook_request_body: Optional[Dict[str, Any]] = Field(default=None, description="Payload sent to the webhook")
    webhook_duration_ms: Optional[float] = Field(default=None, description="Webhook round-trip time in milliseconds")
    webhook_error: Optional[str] = Field(default=None, description="Error message if the webhook failed")


class LoopContextStepOutput(BaseModel):
    iteration: int = Field(..., description="Current iteration number")
    condition_info: Dict[str, Any] = Field(default_factory=dict, description="Condition evaluation info from the previous iteration")
    previous_output: Optional[Dict[str, Any]] = Field(default=None, description="Termination data from the previous iteration")


class ForEachSentinelStartStepOutput(BaseModel):
    message: str = Field(..., description="Status message about the sentinel start")
    mr_phase: str = Field(default="processing", description="Current for-each phase")
    mr_id: str = Field(..., description="ID of the for-each container")
    current_index: int = Field(..., description="Current item index")
    total_items: int = Field(..., description="Total number of items to process")
    max_iterations: int = Field(..., description="Maximum allowed iterations")
    is_first_iteration: bool = Field(..., description="Whether this is the first iteration")
    map_method: Optional[str] = Field(default=None, description="Map method used for splitting")
    current_item_key: Optional[str] = Field(default=None, description="Partition key for the current item")
    all_item_keys: Optional[List[str]] = Field(default=None, description="All item keys")
    all_iteration_context_texts: Optional[List[str]] = Field(default=None, description="All iteration context texts")


class ForEachSentinelEndStepOutput(BaseModel):
    message: str = Field(..., description="Status message about the sentinel end")
    mr_phase: str = Field(default="processing", description="Current for-each phase")
    mr_id: str = Field(..., description="ID of the for-each container")
    current_index: int = Field(..., description="Current item index")
    total_items: int = Field(..., description="Total number of items processed")
    max_iterations: int = Field(..., description="Maximum allowed iterations")
    should_continue: bool = Field(..., description="Whether more items remain")
    is_reduce_phase: bool = Field(..., description="Whether this is the reduce phase")
    all_item_keys: Optional[List[str]] = Field(default=None, description="All item keys")
    output: Optional[Dict[str, Any]] = Field(default=None, description="Reduced output payload")


class WhileLoopSentinelStartStepOutput(BaseModel):
    message: str = Field(..., description="Status message about the sentinel start")
    loop_id: str = Field(..., description="ID of the while-loop container")
    iteration: int = Field(..., description="Current iteration number")
    max_iterations: int = Field(..., description="Maximum allowed iterations")
    passed_through: bool = Field(default=True, description="Whether data was passed through")
    iteration_context_text: Optional[str] = Field(default=None, description="Iteration context text")


class WhileLoopSentinelEndStepOutput(BaseModel):
    message: str = Field(..., description="Status message about the sentinel end")
    loop_id: str = Field(..., description="ID of the while-loop container")
    iteration: int = Field(..., description="Current iteration number")
    max_iterations: int = Field(..., description="Maximum allowed iterations")
    should_continue: bool = Field(..., description="Whether the loop should continue")
    termination_reason: Optional[str] = Field(default=None, description="Reason for termination if the loop exits")
    evaluations: List[Dict[str, Any]] = Field(default_factory=list, description="Condition evaluation results")


WorkflowStepOutputData: TypeAlias = (
    StartDocumentStepOutput
    | SkippedStepOutput
    | ExtractStepOutput
    | ParseStepOutput
    | MergePdfStepOutput
    | MergeDictsStepOutput
    | SplitResponse
    | ClassifierStepOutput
    | FormulaStepOutput
    | HILStepOutput
    | ConditionalStepOutput
    | ConditionalCheckStepOutput
    | APICallStepOutput
    | FunctionStepOutput
    | EditStepOutput
    | EndStepOutput
    | LoopContextStepOutput
    | ForEachSentinelStartStepOutput
    | ForEachSentinelEndStepOutput
    | WhileLoopSentinelStartStepOutput
    | WhileLoopSentinelEndStepOutput
    | Dict[str, Any]
)


_STEP_OUTPUT_MODEL_BY_NODE_TYPE: Dict[str, type[BaseModel]] = {
    "start": StartDocumentStepOutput,
    "parse": ParseStepOutput,
    "edit": EditStepOutput,
    "extract": ExtractStepOutput,
    "split": SplitResponse,
    "classifier": ClassifierStepOutput,
    "conditional": ConditionalStepOutput,
    "hil": HILStepOutput,
    "api_call": APICallStepOutput,
    "function": FunctionStepOutput,
    "formula": FormulaStepOutput,
    "merge_pdf": MergePdfStepOutput,
    "merge_dicts": MergeDictsStepOutput,
    "end": EndStepOutput,
    "for_each_sentinel_start": ForEachSentinelStartStepOutput,
    "for_each_sentinel_end": ForEachSentinelEndStepOutput,
    "while_loop_sentinel_start": WhileLoopSentinelStartStepOutput,
    "while_loop_sentinel_end": WhileLoopSentinelEndStepOutput,
    "loop_context": LoopContextStepOutput,
}


def parse_workflow_step_output(
    node_type: Optional[str],
    output: Any,
) -> Optional[WorkflowStepOutputData]:
    if output is None:
        return None
    if not isinstance(output, dict):
        return output
    if node_type is None:
        return output

    model_cls = _STEP_OUTPUT_MODEL_BY_NODE_TYPE.get(node_type)
    if model_cls is None:
        return output

    try:
        return model_cls.model_validate(output)
    except Exception:
        return output

class StepOutputResponse(BaseModel):
    """Step status and handle data for a specific step in a workflow run."""
    node_id: str = Field(..., description="ID of the node")
    node_type: str = Field(..., description="Type of the node")
    node_label: str = Field(..., description="Label of the node")
    status: str = Field(..., description="Step status")
    handle_outputs: Optional[Dict[str, HandlePayload]] = Field(default=None, description="Handle outputs keyed by handle ID")
    handle_inputs: Optional[Dict[str, HandlePayload]] = Field(default=None, description="Handle inputs keyed by handle ID (what this node received)")

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

        Most extract/formula nodes emit JSON on ``output-json-0``.
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
    node_id: str = Field(..., description="Logical ID of the node")
    step_id: str = Field(..., description="Stored step ID")
    node_type: str = Field(..., description="Type of the node")
    node_label: str = Field(..., description="Label of the node")
    status: str = Field(..., description="Step status")
    started_at: Optional[datetime.datetime] = Field(default=None, description="When the step started")
    completed_at: Optional[datetime.datetime] = Field(default=None, description="When the step completed")
    duration_ms: Optional[int] = Field(default=None, description="Duration in milliseconds")
    error: Optional[str] = Field(default=None, description="Error message if failed")
    handle_outputs: Optional[Dict[str, HandlePayload]] = Field(default=None, description="Handle outputs keyed by handle ID")
    handle_inputs: Optional[Dict[str, HandlePayload]] = Field(default=None, description="Handle inputs keyed by handle ID")
    input_document: Optional[BaseMIMEData] = Field(default=None, description="Reference to input document")
    output_document: Optional[BaseMIMEData] = Field(default=None, description="Reference to output document")
    split_documents: Optional[Dict[str, BaseMIMEData]] = Field(default=None, description="Split node document outputs")
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


class StepOutputsBatchResponse(BaseModel):
    """Response for batch step output retrieval, keyed by node ID."""
    outputs: Dict[str, StepOutputResponse] = Field(
        default_factory=dict,
        description="Step outputs keyed by node ID (missing steps are omitted)",
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
    """Temporal-owned decision state for a workflow HIL node."""
    run_id: str = Field(..., description="Workflow run ID")
    node_id: str = Field(..., description="HIL node ID")
    node_status: Optional[str] = Field(default=None, description="Current workflow node status")
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
        description="Temporal-owned HIL decision state for the node",
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
    subflow_id: Optional[str] = None
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
    subflow_id: Optional[str] = None
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


class WorkflowBlock(BaseModel):
    """A block in a workflow graph."""
    model_config = ConfigDict(extra="ignore")

    id: str = Field(..., description="Block ID")
    workflow_id: str = Field(..., description="Parent workflow ID")
    type: str = Field(..., description="Block type (start, extract, parse, classifier, etc.)")
    label: str = Field(default="", description="Display label")
    position_x: float = Field(default=0, description="X position on canvas")
    position_y: float = Field(default=0, description="Y position on canvas")
    width: Optional[float] = Field(default=None, description="Block width")
    height: Optional[float] = Field(default=None, description="Block height")
    config: Optional[dict] = Field(default=None, description="Block-specific configuration")
    subflow_id: Optional[str] = Field(default=None, description="Parent subflow ID if inside a subflow")
    parent_id: Optional[str] = Field(default=None, description="Parent container block ID (while_loop, for_each)")
    updated_at: Optional[datetime.datetime] = Field(default=None, description="Last updated timestamp")


class WorkflowEdge(BaseModel):
    """An edge connecting two blocks in a workflow graph."""
    model_config = ConfigDict(extra="ignore")

    id: str = Field(..., description="Edge ID")
    workflow_id: str = Field(..., description="Parent workflow ID")
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
    type: str = Field(..., description="Subflow type (while, parallel)")
    label: str = Field(default="", description="Display label")
    position_x: float = Field(default=0, description="X position on canvas")
    position_y: float = Field(default=0, description="Y position on canvas")
    width: float = Field(default=400, description="Container width")
    height: float = Field(default=300, description="Container height")
    config: Optional[dict] = Field(default=None, description="Subflow configuration")
    child_block_ids: List[str] = Field(default_factory=list, description="Block IDs inside this subflow")


class WorkflowWithEntities(BaseModel):
    """Complete workflow with its graph structure (blocks, edges, subflows)."""
    model_config = ConfigDict(extra="ignore")

    workflow: Workflow
    blocks: List[WorkflowBlock] = Field(default_factory=list)
    edges: List[WorkflowEdge] = Field(default_factory=list)
    subflows: List[WorkflowSubflow] = Field(default_factory=list)

    @property
    def start_nodes(self) -> List[WorkflowBlock]:
        """Document input start nodes."""
        return [b for b in self.blocks if b.type == "start"]

    @property
    def start_json_nodes(self) -> List[WorkflowBlock]:
        """JSON input start nodes."""
        return [b for b in self.blocks if b.type == "start_json"]
