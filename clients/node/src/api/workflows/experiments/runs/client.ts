import { CompositionClient, RequestOptions } from "../../../../client.js";
import {
    ExperimentContentResponse,
    ExperimentRunListResponse,
    NConsensusValue,
    RunExperimentResponse,
    ZExperimentContentResponse,
    ZExperimentRunListResponse,
    ZRunExperimentResponse,
} from "../types.js";

/**
 * Per-experiment run history and run content inspection.
 */
export default class APIWorkflowExperimentRuns extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    prepare_create(
        workflowId: string,
        experimentId: string,
        { nConsensus, retryFailedOnly = false }: { nConsensus?: NConsensusValue; retryFailedOnly?: boolean } = {}
    ): { url: string; method: string; body: Record<string, unknown> } {
        const body: Record<string, unknown> = {};
        if (nConsensus !== undefined) body.n_consensus = nConsensus;
        if (retryFailedOnly) body.retry_failed_only = true;
        return {
            url: `/workflows/${workflowId}/experiments/${experimentId}/run`,
            method: "POST",
            body,
        };
    }

    prepare_list(workflowId: string, experimentId: string): { url: string; method: string } {
        return {
            url: `/workflows/${workflowId}/experiments/${experimentId}/runs`,
            method: "GET",
        };
    }

    prepare_get(
        workflowId: string,
        experimentId: string,
        { runId }: { runId?: string } = {}
    ): { url: string; method: string; params?: Record<string, unknown> } {
        const params: Record<string, unknown> = {};
        if (runId !== undefined) params.run_id = runId;
        return {
            url: `/workflows/${workflowId}/experiments/${experimentId}/content`,
            method: "GET",
            ...(Object.keys(params).length > 0 ? { params } : {}),
        };
    }

    /**
     * Trigger an experiment run with the current draft block config.
     *
     * Async — returns a `job_id` immediately. Poll the job with `client.jobs`.
     *
     * @example
     * ```typescript
     * const run = await client.workflows.experiments.runs.create({
     *     workflowId: "wf_abc123",
     *     experimentId: "exp_xyz",
     *     nConsensus: 5,
     * });
     * await client.jobs.waitForCompletion(run.job_id);
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
        const request = this.prepare_create(workflowId, experimentId, { nConsensus, retryFailedOnly });

        return this._fetchJson(ZRunExperimentResponse, {
            url: request.url,
            method: request.method,
            body: { ...request.body, ...((options?.body as Record<string, unknown>) || {}) },
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
        const request = this.prepare_list(workflowId, experimentId);
        return this._fetchJson(ZExperimentRunListResponse, {
            url: request.url,
            method: request.method,
            params: options?.params,
            headers: options?.headers,
        });
    }

    async get(
        {
            workflowId,
            experimentId,
            runId,
        }: {
            workflowId: string;
            experimentId: string;
            runId?: string;
        },
        options?: RequestOptions
    ): Promise<ExperimentContentResponse> {
        const request = this.prepare_get(workflowId, experimentId, { runId });
        return this._fetchJson(ZExperimentContentResponse, {
            url: request.url,
            method: request.method,
            params: { ...(request.params || {}), ...(options?.params || {}) },
            headers: options?.headers,
        });
    }
}
