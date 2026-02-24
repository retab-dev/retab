import datetime
from typing import Any, Dict, List, Literal, Optional

from pydantic import BaseModel, Field, ConfigDict

from retab.types.mime import BaseMIMEData


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
    text: Optional[str] = Field(default=None, description="For text handles: text content")


NodeType = Literal["start", "extract", "split", "end", "hil"]


class StepStatus(BaseModel):
    """Status of a single step in workflow execution"""
    node_id: str = Field(..., description="ID of the node")
    node_type: NodeType = Field(..., description="Type of the node")
    node_label: str = Field(..., description="Label of the node")
    status: Literal["pending", "running", "completed", "error", "waiting_for_human", "cancelled"] = Field(..., description="Current status")
    started_at: Optional[datetime.datetime] = Field(default=None, description="When the step started")
    completed_at: Optional[datetime.datetime] = Field(default=None, description="When the step completed")
    duration_ms: Optional[int] = Field(default=None, description="Duration in milliseconds")
    error: Optional[str] = Field(default=None, description="Error message if failed")
    output: Optional[dict] = Field(default=None, description="Output data from the step")
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


# ---------------------------------------------------------------------------
# Type aliases
# ---------------------------------------------------------------------------

WorkflowRunStatus = Literal["pending", "running", "completed", "error", "waiting_for_human", "cancelled"]

TERMINAL_WORKFLOW_RUN_STATUSES: tuple[str, ...] = ("completed", "error", "cancelled")


# ---------------------------------------------------------------------------
# Step output types (from GET /v1/workflows/runs/{run_id}/steps/{node_id})
# ---------------------------------------------------------------------------

class StepOutputResponse(BaseModel):
    """Full output data for a specific step in a workflow run."""
    node_id: str = Field(..., description="ID of the node")
    node_type: str = Field(..., description="Type of the node")
    node_label: str = Field(..., description="Label of the node")
    status: str = Field(..., description="Step status")
    output: Optional[dict] = Field(default=None, description="Step output data")
    handle_outputs: Optional[Dict[str, Any]] = Field(default=None, description="Handle outputs keyed by handle ID")
    handle_inputs: Optional[Dict[str, Any]] = Field(default=None, description="Handle inputs keyed by handle ID (what this node received)")


class StepOutputsBatchResponse(BaseModel):
    """Response for batch step output retrieval, keyed by node ID."""
    outputs: Dict[str, StepOutputResponse] = Field(
        default_factory=dict,
        description="Step outputs keyed by node ID (missing steps are omitted)",
    )


# ---------------------------------------------------------------------------
# Cancel / Restart / Resume response types
# ---------------------------------------------------------------------------

class CancelWorkflowResponse(BaseModel):
    """Response from cancelling a workflow run."""
    run: WorkflowRun
    cancellation_status: Literal["cancelled", "cancellation_requested", "cancellation_failed"] = Field(
        default="cancellation_requested",
        description="Cancellation delivery state",
    )


class ResumeWorkflowResponse(BaseModel):
    """Response from resuming a workflow run after HIL review."""
    run: WorkflowRun
    resume_status: Literal["processing", "queued", "already_processed"] = Field(
        ..., description="Status of the resume operation"
    )
    queue_position: Optional[int] = Field(default=None, description="Position in queue if queued")
    queue_item_id: str = Field(..., description="ID of the queue item for tracking")
