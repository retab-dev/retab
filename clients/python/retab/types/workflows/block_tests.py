"""Pydantic models mirroring `backend/main_server/main_server/services/v1/workflows/block_tests/models.py`.

Models stay in sync with the wire format the backend sends back so SDK
callers get typed responses for `client.workflows.tests.*` calls.
"""

from __future__ import annotations

import datetime
from typing import Annotated, Any, Literal

from pydantic import BaseModel, ConfigDict, Field

# ---------------------------------------------------------------------------
# Status enums
# ---------------------------------------------------------------------------

#: Outcome of evaluating ONE assertion against a block's output.
AssertionResultStatus = Literal["passed", "failed", "blocked", "error"]

#: Status of a TEST RUN — aggregates the assertion result with execution-side
#: state. 7 values total; transient (``queued``, ``running``) only appear on
#: in-flight records, terminal states appear on completed records and on
#: ``latest_run_summary``. See `TerminalBlockTestRunStatus` for the
#: terminal-only subset.
BlockTestRunStatus = Literal[
    "queued",
    "running",
    "passed",
    "failed",
    "blocked",
    "error",
    "cancelled",
]

#: Strict subset of `BlockTestRunStatus` — values that can appear on a
#: completed run record's `latest_run_summary`. Mirrors the backend's
#: `TerminalBlockTestRunStatus` literal.
TerminalBlockTestRunStatus = Literal[
    "passed", "failed", "blocked", "error", "cancelled"
]


# ---------------------------------------------------------------------------
# Discriminated unions: `target` (what's tested) and `source` (where inputs come from)
# ---------------------------------------------------------------------------


class WorkflowTestBlockTarget(BaseModel):
    """Run the test against a single block in the workflow.

    The shape is a discriminated union by `type` so workflow-level targets
    (`{ type: "workflow" }`) can be added later without renaming the field
    at every callsite. Today `block` is the only variant.
    """

    model_config = ConfigDict(extra="forbid")

    type: Literal["block"] = "block"
    block_id: str


class ManualWorkflowTestSource(BaseModel):
    """Hand-written inputs. Use for synthetic test cases."""

    model_config = ConfigDict(extra="forbid")

    type: Literal["manual"] = "manual"
    handle_inputs: dict[str, Any] = Field(default_factory=dict)


class RunStepWorkflowTestSource(BaseModel):
    """Replay the inputs the block received during a previous workflow run.

    `step_id` is required for blocks executed inside a `for_each` (each
    iteration is its own step).
    """

    model_config = ConfigDict(extra="forbid")

    type: Literal["run_step"] = "run_step"
    run_id: str
    step_id: str | None = None


WorkflowTestSource = Annotated[
    ManualWorkflowTestSource | RunStepWorkflowTestSource,
    Field(discriminator="type"),
]


# ---------------------------------------------------------------------------
# Assertion shapes
# ---------------------------------------------------------------------------


class AssertionTarget(BaseModel):
    """Names a declared output handle and an optional dotted path inside
    that handle's payload (e.g. ``output-json-0``, path ``vendor.name``).
    """

    model_config = ConfigDict(extra="allow")

    output_handle_id: str
    path: str = ""


class AssertionSpec(BaseModel):
    """One assertion against one declared output handle.

    Block tests intentionally normalize to one assertion per test —
    multiple small tests beat one broad assertion when something breaks.

    `condition` is intentionally typed as ``dict[str, Any]`` because the
    operator-specific shape varies widely (``equals``, ``compare``,
    ``matches_regex``, ``similarity_gte``, ``llm_judged_as``, ``split_iou_gte``,
    etc.). See `workflows/block-tests.mdx` in the docs for the catalog.

    `extra="allow"` (vs `WorkflowTestBlockTarget`'s strict `extra="forbid"`)
    because `AssertionSpec` is used in BOTH directions — request body on
    create/update AND response body on `WorkflowTest.assertion`. Forbidding
    extras would crash on a backend that adds a new optional field before
    the SDK ships an update.
    """

    model_config = ConfigDict(extra="allow")

    id: str | None = None
    target: AssertionTarget
    condition: dict[str, Any]
    label: str | None = None


class AssertionFailure(BaseModel):
    model_config = ConfigDict(extra="allow")

    code: str
    message: str
    details: dict[str, Any] = Field(default_factory=dict)


class AssertionResult(BaseModel):
    model_config = ConfigDict(extra="allow")

    assertion_id: str
    condition_kind: str
    status: AssertionResultStatus
    actual_value: Any = None
    expected_value: Any = None
    score: float | None = None
    threshold: float | None = None
    metric_kind: str | None = None
    assertion_label: str | None = None
    failure: AssertionFailure | None = None


class VerdictSummary(BaseModel):
    model_config = ConfigDict(extra="allow")

    result: bool
    assertions_passed: int = 0
    assertions_failed: int = 0
    blocked_assertions: int = 0
    failed_assertion_ids: list[str] = Field(default_factory=list)


class LatestBlockTestRunSummary(BaseModel):
    """Compact summary attached to a `WorkflowTest` so list / get responses
    can show the most recent run state without a second fetch.
    """

    model_config = ConfigDict(extra="allow")

    run_record_id: str
    # Typed as the wider `BlockTestRunStatus` for back-compat with stored
    # docs. In practice only `TerminalBlockTestRunStatus` values are
    # populated (the runner only writes summaries on terminal-state
    # transitions).
    status: BlockTestRunStatus
    started_at: datetime.datetime
    completed_at: datetime.datetime | None = None
    duration_ms: int | None = None
    workflow_draft_fingerprint: str = ""
    block_config_fingerprint: str = ""
    assertions_passed: int = 0
    assertions_failed: int = 0
    blocked_assertions: int = 0


# ---------------------------------------------------------------------------
# Drift / validation auxiliary shapes
# ---------------------------------------------------------------------------


class AssertionSchemaDep(BaseModel):
    model_config = ConfigDict(extra="allow")

    schema_path: str
    subtree_hash: str
    depends_on_root: bool = False


AssertionDriftStatus = Literal["fresh", "drifted", "unknown"]
SchemaDriftStatus = Literal["fresh", "partial", "drifted", "unknown"]


# ---------------------------------------------------------------------------
# Top-level shapes returned by the API
# ---------------------------------------------------------------------------


class WorkflowTest(BaseModel):
    """Public response shape for a single test."""

    model_config = ConfigDict(extra="allow")

    id: str
    workflow_id: str
    organization_id: str
    target: WorkflowTestBlockTarget
    source: WorkflowTestSource
    name: str | None = None
    # Optional because pre-rewrite tests in storage may have ``assertion=None``.
    # Create / Update API still REQUIRE assertion via `CreateBlockTestRequest`;
    # this field reflects what comes back on a list/get for a legacy test that
    # hasn't been re-saved yet.
    assertion: AssertionSpec | None = None
    assertion_schema_dep: AssertionSchemaDep | None = None
    assertion_drift_status: AssertionDriftStatus | None = None
    schema_drift: SchemaDriftStatus = "unknown"
    schema_drift_detail: str | None = None
    validation_status: str = "valid"
    validation_issues: list[Any] = Field(default_factory=list)
    latest_run_summary: LatestBlockTestRunSummary | None = None
    latest_passing_run_summary: LatestBlockTestRunSummary | None = None
    latest_failing_run_summary: LatestBlockTestRunSummary | None = None
    created_at: datetime.datetime
    updated_at: datetime.datetime


class WorkflowTestRunRecord(BaseModel):
    """Public response shape for a single run record."""

    model_config = ConfigDict(extra="allow")

    id: str
    test_id: str
    status: BlockTestRunStatus
    workflow_id: str
    organization_id: str
    target: WorkflowTestBlockTarget
    execution_fingerprint: str = ""
    handle_inputs_fingerprint: str = ""
    workflow_draft_fingerprint: str = ""
    block_config_fingerprint: str = ""
    started_at: datetime.datetime
    completed_at: datetime.datetime | None = None
    duration_ms: int | None = None
    source: WorkflowTestSource
    outputs: dict[str, Any] | None = None
    routing_decision: list[str] | None = None
    warnings: list[str] = Field(default_factory=list)
    error: str | None = None
    skipped: bool = False
    assertion_result: AssertionResult | None = None
    verdict_summary: VerdictSummary | None = None


class BlockTestBatchExecutionCounts(BaseModel):
    """One bucket per `BlockTestRunStatus` value. Today only terminal
    buckets are populated by the runner; transient ones are declared for
    forward-compat with any future code path that persists transient state.
    """

    model_config = ConfigDict(extra="allow")

    queued: int = 0
    running: int = 0
    passed: int = 0
    failed: int = 0
    blocked: int = 0
    error: int = 0
    cancelled: int = 0


class BlockTestBatchExecutionItem(BaseModel):
    model_config = ConfigDict(extra="allow")

    test_id: str
    run_record_id: str
    status: BlockTestRunStatus
    workflow_id: str
    target: WorkflowTestBlockTarget
    duration_ms: int | None = None


class BlockTestBatchExecutionResult(BaseModel):
    """The payload that lands on `Job.result` after `tests.execute(...)`."""

    model_config = ConfigDict(extra="allow")

    workflow_id: str
    target: WorkflowTestBlockTarget | None = None
    counts: BlockTestBatchExecutionCounts = Field(
        default_factory=BlockTestBatchExecutionCounts
    )
    results: list[BlockTestBatchExecutionItem] = Field(default_factory=list)


class ExecuteBlockTestsResponse(BaseModel):
    """Synchronous response from `tests.execute(...)`. Poll
    `client.jobs.retrieve(job_id)` until terminal to fetch the
    `BlockTestBatchExecutionResult`.
    """

    model_config = ConfigDict(extra="allow")

    batch_id: str
    job_id: str
    status: Literal["queued"] = "queued"
    workflow_id: str
    target: WorkflowTestBlockTarget | None = None
    test_id: str | None = None
    total_tests: int


class BlockTestListResponse(BaseModel):
    """Response shape for `tests.list(...)`."""

    model_config = ConfigDict(extra="allow")

    tests: list[WorkflowTest] = Field(default_factory=list)


class BlockTestRunListResponse(BaseModel):
    """Response shape for `tests.runs.list(...)`."""

    model_config = ConfigDict(extra="allow")

    runs: list[WorkflowTestRunRecord] = Field(default_factory=list)


__all__ = [
    "AssertionDriftStatus",
    "AssertionFailure",
    "AssertionResult",
    "AssertionResultStatus",
    "AssertionSchemaDep",
    "AssertionSpec",
    "AssertionTarget",
    "BlockTestBatchExecutionCounts",
    "BlockTestBatchExecutionItem",
    "BlockTestBatchExecutionResult",
    "BlockTestListResponse",
    "BlockTestRunListResponse",
    "BlockTestRunStatus",
    "ExecuteBlockTestsResponse",
    "LatestBlockTestRunSummary",
    "ManualWorkflowTestSource",
    "RunStepWorkflowTestSource",
    "SchemaDriftStatus",
    "TerminalBlockTestRunStatus",
    "VerdictSummary",
    "WorkflowTest",
    "WorkflowTestBlockTarget",
    "WorkflowTestRunRecord",
    "WorkflowTestSource",
]
