import { CompositionClient, RequestOptions } from "../../../client.js";
import { ZJob } from "../../../types.js";
import APIWorkflowTestRuns from "./runs/client.js";
import {
    AssertionSpec,
    BlockTestBatchExecutionResult,
    BlockTestListResponse,
    ExecuteBlockTestsResponse,
    WorkflowTest,
    WorkflowTestBlockTarget,
    WorkflowTestSource,
    ZBlockTestBatchExecutionResult,
    ZBlockTestListResponse,
    ZExecuteBlockTestsResponse,
    ZWorkflowTest,
} from "./types.js";

/**
 * Job statuses that indicate the runner is done. Mirrors the Python
 * `_TERMINAL_JOB_STATUSES` set in `retab/resources/workflows/tests/client.py`.
 * Failed/cancelled/expired make `waitForCompletion` throw; only `completed`
 * returns the parsed payload.
 */
const TERMINAL_JOB_STATUSES: ReadonlySet<string> = new Set([
    "completed",
    "failed",
    "cancelled",
    "expired",
]);

/**
 * Workflow block-tests API client. Mirrors the eight backend endpoints
 * under `/v1/workflows/{workflow_id}/block-tests` plus the nested run
 * endpoints.
 *
 * Sub-clients:
 * - runs: read-only access to per-test execution history
 *
 * @example
 * ```typescript
 * const test = await client.workflows.tests.create({
 *     workflowId: "wf_abc123",
 *     target: { type: "block", block_id: "block_extract" },
 *     source: { type: "manual", handle_inputs: {} },
 *     assertion: {
 *         target: { output_handle_id: "output-json-0", path: "total" },
 *         condition: { kind: "equals", expected: 1234.56 },
 *     },
 *     name: "Q1 invoice total",
 * });
 *
 * // Run all tests for the workflow asynchronously.
 * const batch = await client.workflows.tests.execute({
 *     workflowId: "wf_abc123",
 * });
 * console.log(batch.batch_id, batch.job_id);
 * ```
 */
export default class APIWorkflowTests extends CompositionClient {
    public runs: APIWorkflowTestRuns;

    constructor(client: CompositionClient) {
        super(client);
        this.runs = new APIWorkflowTestRuns(this);
    }

    /**
     * Create a new block test against a single block in the workflow.
     *
     * `assertion` is required. Pass `target` and `source` as discriminated
     * union literals — see types.ts for the variants.
     */
    async create(
        {
            workflowId,
            target,
            source,
            assertion,
            name,
        }: {
            workflowId: string;
            target: WorkflowTestBlockTarget;
            // `WorkflowTestSource` IS `Manual | RunStep` already — no need
            // to repeat the union members. Callers benefit from the named
            // alias showing up in IDE hover.
            source: WorkflowTestSource;
            assertion: AssertionSpec;
            // Narrowed to `string | undefined` so the typed contract matches
            // behavior — the SDK omits `name` from the body whether the
            // caller passes `undefined` or `null`, but the TS signature
            // shouldn't lie about supporting `null`.
            name?: string;
        },
        options?: RequestOptions
    ): Promise<WorkflowTest> {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const body: Record<string, any> = {
            target,
            source,
            assertion,
        };
        if (name !== undefined) {
            body.name = name;
        }
        return this._fetchJson(ZWorkflowTest, {
            url: `/workflows/${workflowId}/block-tests`,
            method: "POST",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Fetch a single block test by id (refreshes drift state).
     */
    async get(
        {
            workflowId,
            testId,
        }: {
            workflowId: string;
            testId: string;
        },
        options?: RequestOptions
    ): Promise<WorkflowTest> {
        return this._fetchJson(ZWorkflowTest, {
            url: `/workflows/${workflowId}/block-tests/${testId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * List all block tests for a workflow.
     */
    async list(
        {
            workflowId,
            targetBlockId,
            limit = 50,
        }: {
            workflowId: string;
            targetBlockId?: string;
            limit?: number;
        },
        options?: RequestOptions
    ): Promise<BlockTestListResponse> {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const baseParams: Record<string, any> = { limit };
        if (targetBlockId !== undefined) {
            baseParams.target_block_id = targetBlockId;
        }
        // `options.params` wins so callers can override the typed args via
        // the escape hatch — matches sibling `APIWorkflowRuns.list`
        // precedence (runs/client.ts).
        return this._fetchJson(ZBlockTestListResponse, {
            url: `/workflows/${workflowId}/block-tests`,
            method: "GET",
            params: { ...baseParams, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    /**
     * Patch the name, assertion, and/or source of a block test. Each
     * field is optional — fields you omit are left untouched. Setting
     * `assertion: null` is rejected by the backend; use `delete()` to
     * remove the test entirely.
     */
    async update(
        {
            workflowId,
            testId,
            name,
            assertion,
            source,
        }: {
            workflowId: string;
            testId: string;
            name?: string;
            assertion?: AssertionSpec;
            source?: WorkflowTestSource;
        },
        options?: RequestOptions
    ): Promise<WorkflowTest> {
        // Only include fields the caller actually passed — the backend
        // treats missing fields as `leave untouched` and explicitly rejects
        // an `assertion: null` payload.
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const body: Record<string, any> = {};
        if (name !== undefined) body.name = name;
        if (assertion !== undefined) body.assertion = assertion;
        if (source !== undefined) body.source = source;

        return this._fetchJson(ZWorkflowTest, {
            url: `/workflows/${workflowId}/block-tests/${testId}`,
            method: "PATCH",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Delete a block test. Returns void on success.
     */
    async delete(
        {
            workflowId,
            testId,
        }: {
            workflowId: string;
            testId: string;
        },
        options?: RequestOptions
    ): Promise<void> {
        return this._fetchJson({
            url: `/workflows/${workflowId}/block-tests/${testId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Run one block test, all tests for a single block, or every test in
     * a workflow.
     *
     * Provide EXACTLY ONE of:
     * - `testId` — run a single test by id
     * - `target` — run all tests for a single block
     * - neither — run every test in the workflow
     *
     * `nConsensus` is optional; allowed values are 3, 5, or 7.
     *
     * Execution is asynchronous: the response carries `batch_id` + `job_id`.
     * Poll `client.jobs.retrieve(jobId)` until terminal to fetch the
     * per-test results.
     */
    async execute(
        {
            workflowId,
            testId,
            target,
            nConsensus,
        }: {
            workflowId: string;
            testId?: string;
            target?: WorkflowTestBlockTarget;
            nConsensus?: number;
        },
        options?: RequestOptions
    ): Promise<ExecuteBlockTestsResponse> {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const body: Record<string, any> = {};
        if (testId !== undefined) body.test_id = testId;
        if (target !== undefined) body.target = target;
        if (nConsensus !== undefined) body.n_consensus = nConsensus;

        return this._fetchJson(ZExecuteBlockTestsResponse, {
            url: `/workflows/${workflowId}/block-tests/execute`,
            method: "POST",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Poll the test-batch job until it reaches a terminal state and
     * return the parsed result.
     *
     * Mirrors the Python SDK's `WorkflowTests.wait_for_completion`. Both
     * the Python helper and this one fetch `/jobs/{job_id}` directly
     * rather than going through a separate jobs sub-client, to keep the
     * tests resource self-contained (no cross-resource coupling).
     *
     * @param jobId - The `job_id` returned by `execute()`.
     * @param pollIntervalMs - Milliseconds between polls (default 2000).
     * @param timeoutMs - Maximum time to wait (default 600000 = 10 min).
     *
     * @throws if the job ends in `failed` / `cancelled` / `expired` (the
     *   error message includes the backend's error payload when available),
     *   or if the deadline is reached before the job terminates.
     *
     * @example
     * ```typescript
     * const batch = await client.workflows.tests.execute({
     *     workflowId: "wf_abc123",
     * });
     * const result = await client.workflows.tests.waitForCompletion(batch.job_id);
     * console.log(result.counts.passed, "tests passed");
     * ```
     */
    async waitForCompletion(
        jobId: string,
        {
            pollIntervalMs = 2000,
            timeoutMs = 600000,
        }: {
            pollIntervalMs?: number;
            timeoutMs?: number;
        } = {},
        options?: RequestOptions
    ): Promise<BlockTestBatchExecutionResult> {
        if (pollIntervalMs <= 0) throw new Error("pollIntervalMs must be > 0");
        if (timeoutMs <= 0) throw new Error("timeoutMs must be > 0");

        const deadlineMs = Date.now() + timeoutMs;
        // Loop body fetches the job, dispatches on terminal status. Same
        // structure as `APIJobs.waitForCompletion` (jobs/client.ts:142-180).
        // eslint-disable-next-line no-constant-condition
        while (true) {
            const job = await this._fetchJson(ZJob, {
                url: `/jobs/${jobId}`,
                method: "GET",
                params: options?.params,
                headers: options?.headers,
            });
            if (TERMINAL_JOB_STATUSES.has(job.status)) {
                if (job.status !== "completed") {
                    throw new Error(
                        `Test batch job ${jobId} ended in status '${job.status}': ${JSON.stringify(job.error)}`,
                    );
                }
                const payload = job.response?.body ?? {};
                return ZBlockTestBatchExecutionResult.parse(payload);
            }
            const nowMs = Date.now();
            if (nowMs >= deadlineMs) {
                throw new Error(
                    `Test batch job ${jobId} did not complete within ${timeoutMs}ms`,
                );
            }
            const sleepMs = Math.min(pollIntervalMs, Math.max(deadlineMs - nowMs, 0));
            await new Promise((resolve) => setTimeout(resolve, sleepMs));
        }
    }

    /** snake_case alias matching the Python SDK method name. */
    async wait_for_completion(
        jobId: string,
        opts?: { pollIntervalMs?: number; timeoutMs?: number },
        options?: RequestOptions
    ): Promise<BlockTestBatchExecutionResult> {
        return this.waitForCompletion(jobId, opts, options);
    }
}
