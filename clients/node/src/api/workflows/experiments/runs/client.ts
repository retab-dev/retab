import * as z from "zod";
import { CompositionClient, RequestOptions } from "../../../../client.js";
import { ZJob } from "../../../../types.js";
import {
    ExperimentJobResponse,
    ExperimentRunListResponse,
    NConsensusValue,
    RunExperimentResponse,
    ZExperimentJobResponse,
    ZExperimentRunListResponse,
    ZRunExperimentResponse,
} from "../types.js";

/**
 * Job statuses that indicate the runner is done. Matches the Python SDK's
 * `_TERMINAL_JOB_STATUSES`. Failed/cancelled/expired make `waitForCompletion`
 * throw; only `completed` returns the underlying job.
 */
const TERMINAL_JOB_STATUSES: ReadonlySet<string> = new Set([
    "completed",
    "failed",
    "cancelled",
    "expired",
]);

type Job = z.infer<typeof ZJob>;

/**
 * Per-experiment run history and per-document job inspection.
 */
export default class APIWorkflowExperimentRuns extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    /**
     * Trigger an experiment run with the current draft block config.
     *
     * Async — returns a `job_id` immediately. Poll the job with
     * `client.jobs.retrieve(jobId)` or call {@link waitForCompletion}.
     *
     * @example
     * ```typescript
     * const run = await client.workflows.experiments.runs.create({
     *     workflowId: "wf_abc123",
     *     experimentId: "exp_xyz",
     *     nConsensus: 5,
     * });
     * await client.workflows.experiments.runs.waitForCompletion(run.job_id);
     * ```
     */
    async create(
        {
            workflowId,
            experimentId,
            nConsensus,
            retryFailedOnly = false,
        }: {
            workflowId: string;
            experimentId: string;
            nConsensus?: NConsensusValue;
            retryFailedOnly?: boolean;
        },
        options?: RequestOptions
    ): Promise<RunExperimentResponse> {
        const body: Record<string, unknown> = {};
        if (nConsensus !== undefined) body.n_consensus = nConsensus;
        if (retryFailedOnly) body.retry_failed_only = true;

        return this._fetchJson(ZRunExperimentResponse, {
            url: `/workflows/${workflowId}/experiments/${experimentId}/run`,
            method: "POST",
            body: { ...body, ...((options?.body as Record<string, unknown>) || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * List historical runs for one experiment, newest first.
     */
    async list(
        {
            workflowId,
            experimentId,
        }: {
            workflowId: string;
            experimentId: string;
        },
        options?: RequestOptions
    ): Promise<ExperimentRunListResponse> {
        return this._fetchJson(ZExperimentRunListResponse, {
            url: `/workflows/${workflowId}/experiments/${experimentId}/runs`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Get one document's job within a specific run.
     */
    async getJob(
        {
            workflowId,
            experimentId,
            runId,
            documentId,
        }: {
            workflowId: string;
            experimentId: string;
            runId: string;
            documentId: string;
        },
        options?: RequestOptions
    ): Promise<ExperimentJobResponse> {
        return this._fetchJson(ZExperimentJobResponse, {
            url: `/workflows/${workflowId}/experiments/${experimentId}/runs/${runId}/documents/${documentId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Poll an experiment-run job until it reaches a terminal state.
     *
     * Returns the underlying `Job`. Throws if the job ends in
     * `failed` / `cancelled` / `expired`, or if it doesn't terminate within
     * `timeoutMs`. Same structure as `APIWorkflowTests.waitForCompletion`.
     */
    async waitForCompletion(
        jobId: string,
        {
            pollIntervalMs = 2000,
            timeoutMs = 1_800_000,
        }: {
            pollIntervalMs?: number;
            timeoutMs?: number;
        } = {},
        options?: RequestOptions
    ): Promise<Job> {
        if (pollIntervalMs <= 0) throw new Error("pollIntervalMs must be > 0");
        if (timeoutMs <= 0) throw new Error("timeoutMs must be > 0");

        const deadlineMs = Date.now() + timeoutMs;
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
                        `Experiment run job ${jobId} ended in status '${job.status}': ${JSON.stringify(job.error)}`,
                    );
                }
                return job;
            }
            const nowMs = Date.now();
            if (nowMs >= deadlineMs) {
                throw new Error(
                    `Experiment run job ${jobId} did not complete within ${timeoutMs}ms`,
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
    ): Promise<Job> {
        return this.waitForCompletion(jobId, opts, options);
    }
}
