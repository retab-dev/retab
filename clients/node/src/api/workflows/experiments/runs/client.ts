import { CompositionClient, RequestOptions } from "../../../../client.js";
import {
    CancelExperimentResponse,
    ExperimentContentResponse,
    ExperimentRunListResponse,
    RunExperimentResponse,
    ZCancelExperimentResponse,
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
        experimentId: string
    ): { url: string; method: string; body: Record<string, unknown> } {
        return {
            url: `/workflows/${workflowId}/experiments/${experimentId}/runs`,
            method: "POST",
            body: {},
        };
    }

    prepare_list(workflowId: string, experimentId: string): { url: string; method: string } {
        return {
            url: `/workflows/${workflowId}/experiments/${experimentId}/runs`,
            method: "GET",
        };
    }

    prepare_cancel_document(
        workflowId: string,
        experimentId: string,
        documentId: string
    ): { url: string; method: string; body: Record<string, never> } {
        return {
            url: `/workflows/${workflowId}/experiments/${experimentId}/documents/${documentId}/cancel`,
            method: "POST",
            body: {},
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
     * Ensure an experiment has fresh results for the current draft block config.
     *
     * Async when work is needed — inspect the returned `run_id` through
     * the experiment runs surface unless `noop` is true.
     *
     * @example
     * ```typescript
     * const run = await client.workflows.experiments.runs.create({
     *     workflowId: "wf_abc123",
     *     experimentId: "exp_xyz",
     * });
     * console.log(run.run_id, run.status);
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
    ): Promise<RunExperimentResponse> {
        const request = this.prepare_create(workflowId, experimentId);

        return this._fetchJson(ZRunExperimentResponse, {
            url: request.url,
            method: request.method,
            body: { ...request.body, ...((options?.body as Record<string, unknown>) || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async cancel_document(
        {
            workflowId,
            experimentId,
            documentId,
        }: {
            workflowId: string;
            experimentId: string;
            documentId: string;
        },
        options?: RequestOptions
    ): Promise<CancelExperimentResponse> {
        const request = this.prepare_cancel_document(workflowId, experimentId, documentId);
        return this._fetchJson(ZCancelExperimentResponse, {
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
