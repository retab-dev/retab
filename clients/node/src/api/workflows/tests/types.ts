/**
 * Zod schemas + TS types for the workflow tests API.
 *
 * Mirrors `backend/main_server/main_server/services/v1/workflows/tests/models.py`.
 * Wire-format field names are snake_case (matching what the Pydantic
 * models serialize); SDK callers use these as plain object literals.
 */

import * as z from "zod";
import {
    ZRunLifecycle,
    ZRunTiming,
    ZTrigger,
    ZWorkflowSnapshotRef,
} from "../../../generated_types.js";

// ---------------------------------------------------------------------------
// Status enums
// ---------------------------------------------------------------------------

/**
 * Outcome of evaluating ONE assertion against a block's output.
 */
export const ZAssertionResultStatus = z.enum([
    "passed",
    "failed",
    "blocked",
    "error",
]);
export type AssertionResultStatus = z.infer<typeof ZAssertionResultStatus>;

/**
 * Status of a TEST RUN — aggregates the assertion result with execution-side
 * state. 7 values total. Transient (`queued`, `running`) only appear on
 * in-flight records; terminal states appear on completed records and on
 * `latest_run_summary`.
 */
export const ZWorkflowTestRunStatus = z.enum([
    "queued",
    "running",
    "passed",
    "failed",
    "blocked",
    "error",
    "cancelled",
]);
export type WorkflowTestRunStatus = z.infer<typeof ZWorkflowTestRunStatus>;

/**
 * Strict subset of `WorkflowTestRunStatus` — values that can appear on a
 * completed run record's `latest_run_summary`.
 */
export const ZTerminalWorkflowTestRunStatus = z.enum([
    "passed",
    "failed",
    "blocked",
    "error",
    "cancelled",
]);
export type TerminalWorkflowTestRunStatus = z.infer<typeof ZTerminalWorkflowTestRunStatus>;

// ---------------------------------------------------------------------------
// Discriminated unions: target (what's tested) and source (where inputs come from)
// ---------------------------------------------------------------------------

export const ZWorkflowTestBlockTarget = z
    .object({
        type: z.literal("block"),
        block_id: z.string(),
    })
    .strict();
export type WorkflowTestBlockTarget = z.infer<typeof ZWorkflowTestBlockTarget>;

export const ZManualWorkflowTestSource = z
    .object({
        type: z.literal("manual"),
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        handle_inputs: z.record(z.string(), z.any()).default({}),
    })
    .strict();
export type ManualWorkflowTestSource = z.infer<typeof ZManualWorkflowTestSource>;

export const ZRunStepWorkflowTestSource = z
    .object({
        type: z.literal("run_step"),
        run_id: z.string(),
        step_id: z.string().nullable().optional(),
    })
    .strict();
export type RunStepWorkflowTestSource = z.infer<typeof ZRunStepWorkflowTestSource>;

export const ZWorkflowTestSource = z.discriminatedUnion("type", [
    ZManualWorkflowTestSource,
    ZRunStepWorkflowTestSource,
]);
export type WorkflowTestSource = z.infer<typeof ZWorkflowTestSource>;

// ---------------------------------------------------------------------------
// Assertion shapes
// ---------------------------------------------------------------------------

export const ZAssertionTarget = z
    .object({
        output_handle_id: z.string(),
        path: z.string().default(""),
    })
    .passthrough();
export type AssertionTarget = z.infer<typeof ZAssertionTarget>;

export const ZAssertionSpec = z
    .object({
        id: z.string().nullable().optional(),
        target: ZAssertionTarget,
        // The condition shape varies per operator (equals, compare, similarity_gte,
        // llm_judged_as, ...). The SDK passes it through as a plain object so
        // callers don't have to import a discriminated union of every operator.
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        condition: z.record(z.string(), z.any()),
        label: z.string().nullable().optional(),
    })
    .passthrough();
export type AssertionSpec = z.infer<typeof ZAssertionSpec>;

export const ZAssertionFailure = z
    .object({
        code: z.string(),
        message: z.string(),
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        details: z.record(z.string(), z.any()).default({}),
    })
    .passthrough();
export type AssertionFailure = z.infer<typeof ZAssertionFailure>;

export const ZAssertionResult = z
    .object({
        assertion_id: z.string(),
        condition_kind: z.string(),
        status: ZAssertionResultStatus,
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        actual_value: z.any().nullable().optional(),
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        expected_value: z.any().nullable().optional(),
        score: z.number().nullable().optional(),
        threshold: z.number().nullable().optional(),
        metric_kind: z.string().nullable().optional(),
        assertion_label: z.string().nullable().optional(),
        failure: ZAssertionFailure.nullable().optional(),
    })
    .passthrough();
export type AssertionResult = z.infer<typeof ZAssertionResult>;

export const ZVerdictSummary = z
    .object({
        result: z.boolean(),
        assertions_passed: z.number().default(0),
        assertions_failed: z.number().default(0),
        blocked_assertions: z.number().default(0),
        failed_assertion_ids: z.array(z.string()).default([]),
    })
    .passthrough();
export type VerdictSummary = z.infer<typeof ZVerdictSummary>;

export const ZLatestWorkflowTestRunSummary = z
    .object({
        run_record_id: z.string(),
        // Wider type for back-compat — in practice only terminal values appear.
        status: ZWorkflowTestRunStatus,
        started_at: z.string().nullable().optional(),
        completed_at: z.string().nullable().optional(),
        duration_ms: z.number().nullable().optional(),
        workflow_draft_fingerprint: z.string().default(""),
        block_config_fingerprint: z.string().default(""),
        assertions_passed: z.number().default(0),
        assertions_failed: z.number().default(0),
        blocked_assertions: z.number().default(0),
    })
    .passthrough();
export type LatestWorkflowTestRunSummary = z.infer<typeof ZLatestWorkflowTestRunSummary>;

// ---------------------------------------------------------------------------
// Drift / validation
// ---------------------------------------------------------------------------

export const ZAssertionSchemaDep = z
    .object({
        schema_path: z.string(),
        subtree_hash: z.string(),
        depends_on_root: z.boolean().default(false),
    })
    .passthrough();
export type AssertionSchemaDep = z.infer<typeof ZAssertionSchemaDep>;

export const ZAssertionDriftStatus = z.enum(["fresh", "drifted", "unknown"]);
export type AssertionDriftStatus = z.infer<typeof ZAssertionDriftStatus>;

export const ZSchemaDriftStatus = z.enum(["fresh", "partial", "drifted", "unknown"]);
export type SchemaDriftStatus = z.infer<typeof ZSchemaDriftStatus>;

// ---------------------------------------------------------------------------
// Response models
// ---------------------------------------------------------------------------

export const ZWorkflowTest = z
    .object({
        id: z.string(),
        workflow_id: z.string(),
        target: ZWorkflowTestBlockTarget,
        source: ZWorkflowTestSource,
        name: z.string().nullable().optional(),
        // Nullable to mirror the backend — pre-rewrite tests in storage may
        // have `assertion=None`. Create / Update API still REQUIRE assertion;
        // this field only goes null for legacy tests that haven't been re-saved.
        assertion: ZAssertionSpec.nullable().optional(),
        assertion_schema_dep: ZAssertionSchemaDep.nullable().optional(),
        assertion_drift_status: ZAssertionDriftStatus.nullable().optional(),
        schema_drift: ZSchemaDriftStatus.default("unknown"),
        schema_drift_detail: z.string().nullable().optional(),
        validation_status: z.string().default("valid"),
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        validation_issues: z.array(z.any()).default([]),
        latest_run_summary: ZLatestWorkflowTestRunSummary.nullable().optional(),
        latest_passing_run_summary: ZLatestWorkflowTestRunSummary.nullable().optional(),
        latest_failing_run_summary: ZLatestWorkflowTestRunSummary.nullable().optional(),
        created_at: z.string(),
        updated_at: z.string(),
    })
    .passthrough();
export type WorkflowTest = z.infer<typeof ZWorkflowTest>;

export const ZWorkflowTestResult = z
    .object({
        id: z.string(),
        run_id: z.string(),
        test_id: z.string(),
        lifecycle: ZRunLifecycle,
        timing: ZRunTiming,
        target: ZWorkflowTestBlockTarget,
        execution_fingerprint: z.string().default(""),
        handle_inputs_fingerprint: z.string().default(""),
        workflow_draft_fingerprint: z.string().default(""),
        block_config_fingerprint: z.string().default(""),
        started_at: z.string().nullable().optional(),
        completed_at: z.string().nullable().optional(),
        duration_ms: z.number().nullable().optional(),
        source: ZWorkflowTestSource,
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        outputs: z.record(z.string(), z.any()).nullable().optional(),
        routing_decision: z.array(z.string()).nullable().optional(),
        warnings: z.array(z.string()).default([]),
        error: z.string().nullable().optional(),
        skipped: z.boolean().default(false),
        assertion_result: ZAssertionResult.nullable().optional(),
        verdict_summary: ZVerdictSummary.nullable().optional(),
        verdict: z.record(z.string(), z.any()).nullable().optional(),
    })
    .passthrough();
export type WorkflowTestResult = z.infer<typeof ZWorkflowTestResult>;

/**
 * One bucket per `WorkflowTestRunStatus` value. Today only terminal buckets
 * are populated by the runner; transient ones (`queued` / `running`) are
 * declared for forward-compat with any future code path that persists
 * transient state.
 */
export const ZWorkflowTestBatchExecutionCounts = z
    .object({
        queued: z.number().default(0),
        running: z.number().default(0),
        passed: z.number().default(0),
        failed: z.number().default(0),
        blocked: z.number().default(0),
        error: z.number().default(0),
        cancelled: z.number().default(0),
    })
    .passthrough();
export type WorkflowTestBatchExecutionCounts = z.infer<typeof ZWorkflowTestBatchExecutionCounts>;

export const ZWorkflowTestRun = z
    .object({
        id: z.string(),
        workflow: ZWorkflowSnapshotRef,
        trigger: ZTrigger,
        lifecycle: ZRunLifecycle,
        timing: ZRunTiming,
        target: ZWorkflowTestBlockTarget.nullable().optional(),
        test_id: z.string().nullable().optional(),
        total_tests: z.number(),
        counts: ZWorkflowTestBatchExecutionCounts.default(() => ({
            queued: 0,
            running: 0,
            passed: 0,
            failed: 0,
            blocked: 0,
            error: 0,
            cancelled: 0,
        })),
    })
    .strip();
export type WorkflowTestRun = z.infer<typeof ZWorkflowTestRun>;

// Canonical PaginatedList envelope. The two list routes used to return
// `{"tests": [...]}` and `{"runs": [...]}` respectively — the migration to
// `{data, list_metadata}` brings them in line with every other Retab list
// endpoint (workflows list, files list, extractions list, etc.).
export const ZListMetadata = z
    .object({
        before: z.string().nullable(),
        after: z.string().nullable(),
    })
    .passthrough();
export type ListMetadata = z.infer<typeof ZListMetadata>;

export const ZWorkflowTestListResponse = z
    .object({
        data: z.array(ZWorkflowTest).default([]),
        list_metadata: ZListMetadata,
    })
    .passthrough();
export type WorkflowTestListResponse = z.infer<typeof ZWorkflowTestListResponse>;

export const ZWorkflowTestRunListResponse = z
    .object({
        data: z.array(ZWorkflowTestRun).default([]),
        list_metadata: ZListMetadata,
    })
    .passthrough();
export type WorkflowTestRunListResponse = z.infer<typeof ZWorkflowTestRunListResponse>;

export const ZWorkflowTestResultListResponse = z
    .object({
        data: z.array(ZWorkflowTestResult).default([]),
        list_metadata: ZListMetadata,
    })
    .passthrough();
export type WorkflowTestResultListResponse = z.infer<typeof ZWorkflowTestResultListResponse>;
