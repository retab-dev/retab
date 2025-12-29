import datetime
from typing import Any, Dict, List, Literal, Optional

from pydantic import BaseModel, Field, ConfigDict


class StepIOReference(BaseModel):
    """Reference to step input/output stored in GCS"""
    file_id: Optional[str] = Field(default=None, description="File ID for document storage lookup")
    gcs_path: Optional[str] = Field(default=None, description="GCS path to the stored file")
    filename: Optional[str] = Field(default=None, description="Original filename")
    mime_type: Optional[str] = Field(default=None, description="MIME type of the file")


class HandlePayload(BaseModel):
    """
    Payload for a single output handle.
    
    Each output handle on a node produces a typed payload that can be:
    - file: A document reference (PDF, image, etc.)
    - json: Structured JSON data (extracted data, etc.)
    - text: Plain text content
    """
    type: Literal["file", "json", "text"] = Field(..., description="Type of payload")
    document: Optional[StepIOReference] = Field(default=None, description="For file handles: document reference")
    data: Optional[dict] = Field(default=None, description="For JSON handles: structured data")
    text: Optional[str] = Field(default=None, description="For text handles: text content")


NodeType = Literal["start", "extract", "split", "end", "hil"]


class StepStatus(BaseModel):
    """Status of a single step in workflow execution"""
    node_id: str = Field(..., description="ID of the node")
    node_type: NodeType = Field(..., description="Type of the node")
    node_label: str = Field(..., description="Label of the node")
    status: Literal["pending", "running", "completed", "error", "waiting_for_human"] = Field(..., description="Current status")
    started_at: Optional[datetime.datetime] = Field(default=None, description="When the step started")
    completed_at: Optional[datetime.datetime] = Field(default=None, description="When the step completed")
    duration_ms: Optional[int] = Field(default=None, description="Duration in milliseconds")
    error: Optional[str] = Field(default=None, description="Error message if failed")
    output: Optional[dict] = Field(default=None, description="Output data from the step")
    handle_outputs: Optional[Dict[str, HandlePayload]] = Field(
        default=None, 
        description="Output payloads keyed by handle ID (e.g., 'output-file-0', 'output-json-0')"
    )
    input_document: Optional[StepIOReference] = Field(default=None, description="Reference to input document")
    output_document: Optional[StepIOReference] = Field(default=None, description="Reference to output document")
    split_documents: Optional[Dict[str, StepIOReference]] = Field(default=None, description="For split nodes: category -> document reference")
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
    status: Literal["pending", "running", "completed", "error", "waiting_for_human"] = Field(default="pending", description="Overall status")
    started_at: datetime.datetime = Field(..., description="When the workflow started")
    completed_at: Optional[datetime.datetime] = Field(default=None, description="When the workflow completed")
    duration_ms: Optional[int] = Field(default=None, description="Total duration in milliseconds")
    steps: List[StepStatus] = Field(default_factory=list, description="Status of each step")
    input_documents: Optional[Dict[str, StepIOReference]] = Field(default=None, description="Start node ID -> input document reference")
    final_outputs: Optional[dict] = Field(default=None, description="Final outputs from end nodes")
    error: Optional[str] = Field(default=None, description="Error message if workflow failed")
    created_at: datetime.datetime = Field(..., description="When the run was created")
    updated_at: datetime.datetime = Field(..., description="When the run was last updated")
    waiting_for_node_ids: List[str] = Field(default_factory=list, description="Node IDs that are waiting for human review")
    pending_node_outputs: Optional[dict] = Field(default=None, description="Serialized node outputs to resume from")

