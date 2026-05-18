import { CompositionClient, RequestOptions } from "../../../../client.js";
import {
    ExperimentMetricView,
    ExperimentMetricsResponse,
    ExperimentResult,
    ExperimentResultListResponse,
    ExperimentRun,
    ExperimentRunListResponse,
    ZExperimentMetricsResponse,
    ZExperimentResult,
    ZExperimentResultListResponse,
    ZExperimentRun,
    ZExperimentRunListResponse,
} from "../types.js";

/**
 * Experiment run lifecycle, child results, and metrics.
 */
export default class APIWorkflowExperimentRuns extends CompositionClient {
    public results: APIWorkflowExperimentRunResults;
    public metrics: APIWorkflowExperimentRunMetrics;

    constructor(client: CompositionClient) {
        super(client);
        this.results = new APIWorkflowExperimentRunResults(this);
        this.metrics = new APIWorkflowExperimentRunMetrics(this);
    }

    prepare_create(
        workflowId: string,
        experimentId: string
    ): { url: string; method: string; body: Record<string, unknown> } {
        return {
            url: `/workflows/${workflowId}/experiments/${experimentId}/runs`,
            method: "POST",
            body: {},
        };
    }

    prepare_list({
        workflowId,
        experimentId,
        blockId,
        limit = 20,
    }: {
        workflowId?: string;
        experimentId?: string;
        blockId?: string;
        limit?: number;
    } = {}): { url: string; method: string; params: Record<string, unknown> } {
        const params = Object.fromEntries(
            Object.entries({
                workflow_id: workflowId,
                experiment_id: experimentId,
                block_id: blockId,
                limit,
            }).filter(([, value]) => value !== undefined)
        );
        return {
            url: "/workflows/experiments/runs",
            method: "GET",
            params,
        };
    }

    prepare_get(runId: string): { url: string; method: string } {
        return {
            url: `/workflows/experiments/runs/${runId}`,
            method: "GET",
        };
    }

    prepare_cancel(runId: string): { url: string; method: string; body: Record<string, never> } {
        return {
            url: `/workflows/experiments/runs/${runId}/cancel`,
            method: "POST",
            body: {},
        };
    }

    /**
     * Ensure an experiment has fresh results for the current draft block config.
     *
     * Async when work is needed — inspect the returned `id` through
     * the experiment runs surface.
     *
     * @example
     * ```typescript
     * const run = await client.workflows.experiments.runs.create({
     *     workflowId: "wf_abc123",
     *     experimentId: "exp_xyz",
     * });
     * console.log(run.id, run.lifecycle.status);
     * ```
     */
    async create(
        {
            workflowId,
            experimentId,
        }: {
            workflowId: string;
            experimentId: string;
        },
        options?: RequestOptions
    ): Promise<ExperimentRun> {
        const request = this.prepare_create(workflowId, experimentId);

        return this._fetchJson(ZExperimentRun, {
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
            blockId,
            limit = 20,
        }: {
            workflowId?: string;
            experimentId?: string;
            blockId?: string;
            limit?: number;
        },
        options?: RequestOptions
    ): Promise<ExperimentRunListResponse> {
        const request = this.prepare_list({ workflowId, experimentId, blockId, limit });
        return this._fetchJson(ZExperimentRunListResponse, {
            url: request.url,
            method: request.method,
            params: { ...request.params, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async get(
        {
            runId,
        }: {
            runId: string;
        },
        options?: RequestOptions
    ): Promise<ExperimentRun> {
        const request = this.prepare_get(runId);
        return this._fetchJson(ZExperimentRun, {
            url: request.url,
            method: request.method,
            params: options?.params,
            headers: options?.headers,
        });
    }

    async cancel(
        { runId }: { runId: string },
        options?: RequestOptions
    ): Promise<ExperimentRun> {
        const request = this.prepare_cancel(runId);
        return this._fetchJson(ZExperimentRun, {
            url: request.url,
            method: request.method,
            body: { ...request.body, ...((options?.body as Record<string, unknown>) || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
}

export class APIWorkflowExperimentRunResults extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    async list(
        { runId, limit = 20 }: { runId: string; limit?: number },
        options?: RequestOptions
    ): Promise<ExperimentResultListResponse> {
        return this._fetchJson(ZExperimentResultListResponse, {
            url: `/workflows/experiments/runs/${runId}/results`,
            method: "GET",
            params: { limit, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async get(
        { runId, documentId }: { runId: string; documentId: string },
        options?: RequestOptions
    ): Promise<ExperimentResult> {
        return this._fetchJson(ZExperimentResult, {
            url: `/workflows/experiments/runs/${runId}/results/${documentId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }
}

export class APIWorkflowExperimentRunMetrics extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    async get(
        {
            runId,
            view = "summary",
            documentId,
            targetPath,
            includePrior = true,
            priorRunId,
        }: {
            runId: string;
            view?: ExperimentMetricView;
            documentId?: string;
            targetPath?: string;
            includePrior?: boolean;
            priorRunId?: string;
        },
        options?: RequestOptions
    ): Promise<ExperimentMetricsResponse> {
        const params: Record<string, unknown> = { view, include_prior: includePrior };
        if (documentId !== undefined) params.document_id = documentId;
        if (targetPath !== undefined) params.target_path = targetPath;
        if (priorRunId !== undefined) params.prior_run_id = priorRunId;
        return this._fetchJson(ZExperimentMetricsResponse, {
            url: `/workflows/experiments/runs/${runId}/metrics`,
            method: "GET",
            params: { ...params, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }
}
