"""Pydantic models mirroring `backend/main_server/main_server/services/v1/workflows/tests/models.py`.

Models stay in sync with the wire format the backend sends back so SDK
callers get typed responses for `client.workflows.tests.*` calls.
"""

from __future__ import annotations

import datetime
from typing import Annotated, Any, Literal

from pydantic import ConfigDict, Field
from retab.types.base import RetabBaseModel

# ---------------------------------------------------------------------------
# Status enums
# ---------------------------------------------------------------------------

#: Outcome of evaluating ONE assertion against a block's output.
AssertionResultStatus = Literal["passed", "failed", "blocked", "error"]

#: Status of a TEST RUN — aggregates the assertion result with execution-side
#: state. 7 values total; transient (``queued``, ``running``) only appear on
#: in-flight records, terminal states appear on completed records and on
#: ``latest_run_summary``. See `TerminalWorkflowTestRunStatus` for the
#: terminal-only subset.
WorkflowTestRunStatus = Literal[
    "queued",
    "running",
    "passed",
    "failed",
    "blocked",
    "error",
    "cancelled",
]

#: Strict subset of `WorkflowTestRunStatus` — values that can appear on a
#: completed run record's `latest_run_summary`. Mirrors the backend's
#: `TerminalWorkflowTestRunStatus` literal.
TerminalWorkflowTestRunStatus = Literal[
    "passed", "failed", "blocked", "error", "cancelled"
]

WorkflowTestExecutionRunStatus = Literal["queued", "running", "completed", "failed"]


# ---------------------------------------------------------------------------
# Discriminated unions: `target` (what's tested) and `source` (where inputs come from)
# ---------------------------------------------------------------------------


class WorkflowTestBlockTarget(RetabBaseModel):
    """Run the test against a single block in the workflow.

    The shape is a discriminated union by `type` so workflow-level targets
    (`{ type: "workflow" }`) can be added later without renaming the field
    at every callsite. Today `block` is the only variant.
    """

    model_config = ConfigDict(extra="ignore")

    type: Literal["block"] = "block"
    block_id: str


class ManualWorkflowTestSource(RetabBaseModel):
    """Hand-written inputs. Use for synthetic test cases."""

    model_config = ConfigDict(extra="ignore")

    type: Literal["manual"] = "manual"
    handle_inputs: dict[str, Any] = Field(default_factory=dict)


class RunStepWorkflowTestSource(RetabBaseModel):
    """Replay the inputs the block received during a previous workflow run.

    `step_id` is required for blocks executed inside a `for_each` (each
    iteration is its own step).
    """

    model_config = ConfigDict(extra="ignore")

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


class AssertionTarget(RetabBaseModel):
    """Names a declared output handle and an optional dotted path inside
    that handle's payload (e.g. ``output-json-0``, path ``vendor.name``).
    """

    model_config = ConfigDict(extra="ignore")

    output_handle_id: str
    path: str = ""


class AssertionSpec(RetabBaseModel):
    """One assertion against one declared output handle.

    Workflow tests intentionally normalize to one assertion per test —
    multiple small tests beat one broad assertion when something breaks.

    `condition` is intentionally typed as ``dict[str, Any]`` because the
    operator-specific shape varies widely (``equals``, ``compare``,
    ``matches_regex``, ``similarity_gte``, ``llm_judged_as``, ``split_iou_gte``,
    etc.). See `workflows/tests.mdx` in the docs for the catalog.

    `extra="ignore"` because `AssertionSpec` is used in BOTH directions —
    request body on create/update AND response body on `WorkflowTest.assertion`.
    Forbidding extras would crash on a backend that adds a new optional field
    before the SDK ships an update.
    """

    model_config = ConfigDict(extra="ignore")

    id: str | None = None
    target: AssertionTarget
    condition: dict[str, Any]
    label: str | None = None


class AssertionFailure(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    code: str
    message: str
    details: dict[str, Any] = Field(default_factory=dict)


class AssertionResult(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

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


class VerdictSummary(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    result: bool
    assertions_passed: int = 0
    assertions_failed: int = 0
    blocked_assertions: int = 0
    failed_assertion_ids: list[str] = Field(default_factory=list)


class LatestWorkflowTestRunSummary(RetabBaseModel):
    """Compact summary attached to a `WorkflowTest` so list / get responses
    can show the most recent run state without a second fetch.
    """

    model_config = ConfigDict(extra="ignore")

    run_record_id: str
    # Typed as the wider `WorkflowTestRunStatus` for back-compat with stored
    # docs. In practice only `TerminalWorkflowTestRunStatus` values are
    # populated (the runner only writes summaries on terminal-state
    # transitions).
    status: WorkflowTestRunStatus
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


class AssertionSchemaDep(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    schema_path: str
    subtree_hash: str
    depends_on_root: bool = False


AssertionDriftStatus = Literal["fresh", "drifted", "unknown"]
SchemaDriftStatus = Literal["fresh", "partial", "drifted", "unknown"]


# ---------------------------------------------------------------------------
# Top-level shapes returned by the API
# ---------------------------------------------------------------------------


class WorkflowTest(RetabBaseModel):
    """Public response shape for a single test."""

    model_config = ConfigDict(extra="ignore")

    id: str
    workflow_id: str
    target: WorkflowTestBlockTarget
    source: WorkflowTestSource
    name: str | None = None
    # Optional because pre-rewrite tests in storage may have ``assertion=None``.
    # Create / Update API still REQUIRE assertion via `CreateWorkflowTestRequest`;
    # this field reflects what comes back on a list/get for a legacy test that
    # hasn't been re-saved yet.
    assertion: AssertionSpec | None = None
    assertion_schema_dep: AssertionSchemaDep | None = None
    assertion_drift_status: AssertionDriftStatus | None = None
    schema_drift: SchemaDriftStatus = "unknown"
    schema_drift_detail: str | None = None
    validation_status: str = "valid"
    validation_issues: list[Any] = Field(default_factory=list)
    latest_run_summary: LatestWorkflowTestRunSummary | None = None
    latest_passing_run_summary: LatestWorkflowTestRunSummary | None = None
    latest_failing_run_summary: LatestWorkflowTestRunSummary | None = None
    created_at: datetime.datetime
    updated_at: datetime.datetime


class WorkflowTestRunRecord(RetabBaseModel):
    """Public response shape for a single run record."""

    model_config = ConfigDict(extra="ignore")

    id: str
    test_id: str
    status: WorkflowTestRunStatus
    workflow_id: str
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


class WorkflowTestBatchExecutionCounts(RetabBaseModel):
    """One bucket per `WorkflowTestRunStatus` value. Today only terminal
    buckets are populated by the runner; transient ones are declared for
    forward-compat with any future code path that persists transient state.
    """

    model_config = ConfigDict(extra="ignore")

    queued: int = 0
    running: int = 0
    passed: int = 0
    failed: int = 0
    blocked: int = 0
    error: int = 0
    cancelled: int = 0


class WorkflowTestBatchExecutionItem(RetabBaseModel):
    model_config = ConfigDict(extra="ignore")

    test_id: str
    run_record_id: str
    status: WorkflowTestRunStatus
    workflow_id: str
    target: WorkflowTestBlockTarget
    duration_ms: int | None = None


class WorkflowTestBatchExecutionResult(RetabBaseModel):
    """The payload that lands on `Job.result` after `tests.execute(...)`."""

    model_config = ConfigDict(extra="ignore")

    workflow_id: str
    target: WorkflowTestBlockTarget | None = None
    counts: WorkflowTestBatchExecutionCounts = Field(
        default_factory=WorkflowTestBatchExecutionCounts
    )
    results: list[WorkflowTestBatchExecutionItem] = Field(default_factory=list)


class ExecuteWorkflowTestsResponse(RetabBaseModel):
    """Synchronous response from `tests.execute(...)`.

    Poll the workflow-test run by `run_id` until terminal, then fetch
    run results from the workflow-test runs surface.
    """

    model_config = ConfigDict(extra="ignore")

    run_id: str
    status: WorkflowTestExecutionRunStatus = "queued"
    workflow_id: str
    target: WorkflowTestBlockTarget | None = None
    test_id: str | None = None
    total_tests: int
    counts: WorkflowTestBatchExecutionCounts = Field(
        default_factory=WorkflowTestBatchExecutionCounts
    )
    created_at: datetime.datetime
    started_at: datetime.datetime | None = None
    completed_at: datetime.datetime | None = None
    duration_ms: int | None = None
    error: str | None = None


class WorkflowTestExecutionRunResults(RetabBaseModel):
    """Results for a parent workflow-test execution run."""

    model_config = ConfigDict(extra="ignore")

    run_id: str
    status: WorkflowTestExecutionRunStatus
    workflow_id: str
    target: WorkflowTestBlockTarget | None = None
    test_id: str | None = None
    total_tests: int
    counts: WorkflowTestBatchExecutionCounts = Field(
        default_factory=WorkflowTestBatchExecutionCounts
    )
    results: list[WorkflowTestRunRecord] = Field(default_factory=list)


__all__ = [
    "AssertionDriftStatus",
    "AssertionFailure",
    "AssertionResult",
    "AssertionResultStatus",
    "AssertionSchemaDep",
    "AssertionSpec",
    "AssertionTarget",
    "WorkflowTestBatchExecutionCounts",
    "WorkflowTestBatchExecutionItem",
    "WorkflowTestBatchExecutionResult",
    "WorkflowTestRunStatus",
    "ExecuteWorkflowTestsResponse",
    "LatestWorkflowTestRunSummary",
    "ManualWorkflowTestSource",
    "RunStepWorkflowTestSource",
    "SchemaDriftStatus",
    "TerminalWorkflowTestRunStatus",
    "VerdictSummary",
    "WorkflowTest",
    "WorkflowTestBlockTarget",
    "WorkflowTestExecutionRunStatus",
    "WorkflowTestExecutionRunResults",
    "WorkflowTestRunRecord",
    "WorkflowTestSource",
]
