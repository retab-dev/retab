/**
 * Zod schemas + TS types for the workflow block-tests API.
 *
 * Mirrors `backend/main_server/main_server/services/v1/workflows/block_tests/models.py`.
 * Wire-format field names are snake_case (matching what the Pydantic
 * models serialize); SDK callers use these as plain object literals.
 */

import * as z from "zod";

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
export const ZBlockTestRunStatus = z.enum([
    "queued",
    "running",
    "passed",
    "failed",
    "blocked",
    "error",
    "cancelled",
]);
export type BlockTestRunStatus = z.infer<typeof ZBlockTestRunStatus>;

/**
 * Strict subset of `BlockTestRunStatus` — values that can appear on a
 * completed run record's `latest_run_summary`.
 */
export const ZTerminalBlockTestRunStatus = z.enum([
    "passed",
    "failed",
    "blocked",
    "error",
    "cancelled",
]);
export type TerminalBlockTestRunStatus = z.infer<typeof ZTerminalBlockTestRunStatus>;

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

export const ZLatestBlockTestRunSummary = z
    .object({
        run_record_id: z.string(),
        // Wider type for back-compat — in practice only terminal values appear.
        status: ZBlockTestRunStatus,
        started_at: z.string(),
        completed_at: z.string().nullable().optional(),
        duration_ms: z.number().nullable().optional(),
        workflow_draft_fingerprint: z.string().default(""),
        block_config_fingerprint: z.string().default(""),
        assertions_passed: z.number().default(0),
        assertions_failed: z.number().default(0),
        blocked_assertions: z.number().default(0),
    })
    .passthrough();
export type LatestBlockTestRunSummary = z.infer<typeof ZLatestBlockTestRunSummary>;

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
        organization_id: z.string(),
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
        latest_run_summary: ZLatestBlockTestRunSummary.nullable().optional(),
        latest_passing_run_summary: ZLatestBlockTestRunSummary.nullable().optional(),
        latest_failing_run_summary: ZLatestBlockTestRunSummary.nullable().optional(),
        created_at: z.string(),
        updated_at: z.string(),
    })
    .passthrough();
export type WorkflowTest = z.infer<typeof ZWorkflowTest>;

export const ZWorkflowTestRunRecord = z
    .object({
        id: z.string(),
        test_id: z.string(),
        status: ZBlockTestRunStatus,
        workflow_id: z.string(),
        organization_id: z.string(),
        target: ZWorkflowTestBlockTarget,
        execution_fingerprint: z.string().default(""),
        handle_inputs_fingerprint: z.string().default(""),
        workflow_draft_fingerprint: z.string().default(""),
        block_config_fingerprint: z.string().default(""),
        started_at: z.string(),
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
    })
    .passthrough();
export type WorkflowTestRunRecord = z.infer<typeof ZWorkflowTestRunRecord>;

/**
 * One bucket per `BlockTestRunStatus` value. Today only terminal buckets
 * are populated by the runner; transient ones (`queued` / `running`) are
 * declared for forward-compat with any future code path that persists
 * transient state.
 */
export const ZBlockTestBatchExecutionCounts = z
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
export type BlockTestBatchExecutionCounts = z.infer<typeof ZBlockTestBatchExecutionCounts>;

export const ZBlockTestBatchExecutionItem = z
    .object({
        test_id: z.string(),
        run_record_id: z.string(),
        status: ZBlockTestRunStatus,
        workflow_id: z.string(),
        target: ZWorkflowTestBlockTarget,
        duration_ms: z.number().nullable().optional(),
    })
    .passthrough();
export type BlockTestBatchExecutionItem = z.infer<typeof ZBlockTestBatchExecutionItem>;

export const ZBlockTestBatchExecutionResult = z
    .object({
        workflow_id: z.string(),
        target: ZWorkflowTestBlockTarget.nullable().optional(),
        counts: ZBlockTestBatchExecutionCounts.default(() => ({
            queued: 0,
            running: 0,
            passed: 0,
            failed: 0,
            blocked: 0,
            error: 0,
            cancelled: 0,
        })),
        results: z.array(ZBlockTestBatchExecutionItem).default([]),
    })
    .passthrough();
export type BlockTestBatchExecutionResult = z.infer<typeof ZBlockTestBatchExecutionResult>;

export const ZExecuteBlockTestsResponse = z
    .object({
        batch_id: z.string(),
        job_id: z.string(),
        status: z.literal("queued"),
        workflow_id: z.string(),
        target: ZWorkflowTestBlockTarget.nullable().optional(),
        test_id: z.string().nullable().optional(),
        total_tests: z.number(),
    })
    .passthrough();
export type ExecuteBlockTestsResponse = z.infer<typeof ZExecuteBlockTestsResponse>;

export const ZBlockTestListResponse = z
    .object({
        tests: z.array(ZWorkflowTest).default([]),
    })
    .passthrough();
export type BlockTestListResponse = z.infer<typeof ZBlockTestListResponse>;

export const ZBlockTestRunListResponse = z
    .object({
        runs: z.array(ZWorkflowTestRunRecord).default([]),
    })
    .passthrough();
export type BlockTestRunListResponse = z.infer<typeof ZBlockTestRunListResponse>;
