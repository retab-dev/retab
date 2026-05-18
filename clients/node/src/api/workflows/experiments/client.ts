import { CompositionClient, RequestOptions } from "../../../client.js";
import APIWorkflowExperimentRuns from "./runs/client.js";
import {
    CancelExperimentResponse,
    EligibleBlockListResponse,
    ExperimentDocumentCaptureRequest,
    ExperimentList,
    ExperimentMetricView,
    ExperimentMetricsResponse,
    ExperimentResponse,
    ExplicitExperimentDocumentRequest,
    NConsensusValue,
    RunBatchResponse,
    ZCancelExperimentResponse,
    ZEligibleBlockListResponse,
    ZExperimentList,
    ZExperimentMetricsResponse,
    ZExperimentResponse,
    ZRunBatchResponse,
} from "./types.js";

/**
 * Workflow experiments API client. Mirrors the REST endpoints under
 * `/v1/workflows/{workflow_id}/experiments` and the Python SDK's
 * `client.workflows.experiments.*`.
 *
 * The MCP tools (`experiments_create`, `experiments_runs_create`,
 * `experiments_get_metrics`, …) call the same routes — this resource is
 * the SDK-side equivalent.
 *
 * Sub-clients:
 * - runs: per-experiment run history + run content inspection.
 *
 * @example
 * ```typescript
 * const exp = await client.workflows.experiments.create({
 *     workflowId: "wf_abc123",
 *     blockId: "block_extract",
 *     name: "Q1 invoice totals",
 *     documentCaptures: [{ workflow_run_id: "run_xyz" }],
 *     nConsensus: 5,
 * });
 *
 * const run = await client.workflows.experiments.runs.create({
 *     workflowId: "wf_abc123",
 *     experimentId: exp.id,
 * });
 * console.log(run.run_id, run.status);
 *
 * const metrics = await client.workflows.experiments.getMetrics({
 *     workflowId: "wf_abc123",
 *     experimentId: exp.id,
 *     view: "summary",
 * });
 * ```
 */
export default class APIWorkflowExperiments extends CompositionClient {
    public runs: APIWorkflowExperimentRuns;

    constructor(client: CompositionClient) {
        super(client);
        this.runs = new APIWorkflowExperimentRuns(this);
    }

    prepare_create(
        workflowId: string,
        {
            blockId,
            name,
            documentCaptures,
            documents,
            nConsensus = 5,
        }: {
            blockId: string;
            name: string;
            documentCaptures?: ExperimentDocumentCaptureRequest[];
            documents?: ExplicitExperimentDocumentRequest[];
            nConsensus?: NConsensusValue;
        }
    ): { url: string; method: string; body: Record<string, unknown> } {
        const body: Record<string, unknown> = {
            block_id: blockId,
            name,
            n_consensus: nConsensus,
        };
        if (documentCaptures !== undefined) body.document_captures = documentCaptures;
        if (documents !== undefined) body.documents = documents;
        return {
            url: `/workflows/${workflowId}/experiments`,
            method: "POST",
            body,
        };
    }

    prepare_list(workflowId: string): { url: string; method: string } {
        return {
            url: `/workflows/${workflowId}/experiments`,
            method: "GET",
        };
    }

    prepare_get(workflowId: string, experimentId: string): { url: string; method: string } {
        return {
            url: `/workflows/${workflowId}/experiments/${experimentId}`,
            method: "GET",
        };
    }

    prepare_update(
        workflowId: string,
        experimentId: string,
        {
            name,
            documentCaptures,
            documents,
            nConsensus,
        }: {
            name?: string;
            documentCaptures?: ExperimentDocumentCaptureRequest[];
            documents?: ExplicitExperimentDocumentRequest[];
            nConsensus?: NConsensusValue;
        } = {}
    ): { url: string; method: string; body: Record<string, unknown> } {
        const body: Record<string, unknown> = {};
        if (name !== undefined) body.name = name;
        if (documentCaptures !== undefined) body.document_captures = documentCaptures;
        if (documents !== undefined) body.documents = documents;
        if (nConsensus !== undefined) body.n_consensus = nConsensus;
        return {
            url: `/workflows/${workflowId}/experiments/${experimentId}`,
            method: "PATCH",
            body,
        };
    }

    prepare_delete(workflowId: string, experimentId: string): { url: string; method: string } {
        return {
            url: `/workflows/${workflowId}/experiments/${experimentId}`,
            method: "DELETE",
        };
    }

    prepare_duplicate(workflowId: string, experimentId: string): { url: string; method: string; body: Record<string, never> } {
        return {
            url: `/workflows/${workflowId}/experiments/${experimentId}/duplicate`,
            method: "POST",
            body: {},
        };
    }

    prepare_cancel(workflowId: string, experimentId: string): { url: string; method: string; body: Record<string, never> } {
        return {
            url: `/workflows/${workflowId}/experiments/${experimentId}/cancel`,
            method: "POST",
            body: {},
        };
    }

    prepare_get_metrics(
        workflowId: string,
        experimentId: string,
        {
            view = "summary",
            runId,
            documentId,
            targetPath,
            includePrior = true,
            priorRunId,
        }: {
            view?: ExperimentMetricView;
            runId?: string;
            documentId?: string;
            targetPath?: string;
            includePrior?: boolean;
            priorRunId?: string;
        } = {}
    ): { url: string; method: string; params: Record<string, unknown> } {
        const params: Record<string, unknown> = {
            view,
            include_prior: includePrior,
        };
        if (runId !== undefined) params.run_id = runId;
        if (documentId !== undefined) params.document_id = documentId;
        if (targetPath !== undefined) params.target_path = targetPath;
        if (priorRunId !== undefined) params.prior_run_id = priorRunId;
        return {
            url: `/workflows/${workflowId}/experiments/${experimentId}/metrics`,
            method: "GET",
            params,
        };
    }

    prepare_list_eligible_blocks(workflowId: string): { url: string; method: string } {
        return {
            url: `/workflows/${workflowId}/experiments/eligible-blocks`,
            method: "GET",
        };
    }

    prepare_run_batch(
        workflowId: string,
        { blockId }: { blockId: string }
    ): { url: string; method: string; body: Record<string, unknown> } {
        const body: Record<string, unknown> = { block_id: blockId };
        return {
            url: `/workflows/${workflowId}/experiments/run-batch`,
            method: "POST",
            body,
        };
    }

    /**
     * Create a consensus experiment on a supported block.
     *
     * Supported block kinds: `extract`, `classifier`, `split`, and
     * `for_each` configured with `map_method="split_by_key"`.
     *
     * Provide documents via `documentCaptures` (provenance from a workflow
     * run) and/or `documents` (explicit `handle_inputs`). At least one
     * document is required.
     *
     * Note: this does NOT trigger a run — call
     * `client.workflows.experiments.runs.create(...)` next.
     */
    async create(
        {
            workflowId,
            blockId,
            name,
            documentCaptures,
            documents,
            nConsensus = 5,
        }: {
            workflowId: string;
            blockId: string;
            name: string;
            documentCaptures?: ExperimentDocumentCaptureRequest[];
            documents?: ExplicitExperimentDocumentRequest[];
            nConsensus?: NConsensusValue;
        },
        options?: RequestOptions
    ): Promise<ExperimentResponse> {
        const request = this.prepare_create(workflowId, {
            blockId,
            name,
            documentCaptures,
            documents,
            nConsensus,
        });

        return this._fetchJson(ZExperimentResponse, {
            url: request.url,
            method: request.method,
            body: { ...request.body, ...((options?.body as Record<string, unknown>) || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * List experiments for a workflow.
     */
    async list(workflowId: string, options?: RequestOptions): Promise<ExperimentList> {
        const request = this.prepare_list(workflowId);
        return this._fetchJson(ZExperimentList, {
            url: request.url,
            method: request.method,
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Get a single experiment by ID.
     */
    async get(
        workflowId: string,
        experimentId: string,
        options?: RequestOptions
    ): Promise<ExperimentResponse> {
        const request = this.prepare_get(workflowId, experimentId);
        return this._fetchJson(ZExperimentResponse, {
            url: request.url,
            method: request.method,
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Patch an experiment. Only the fields you supply are updated.
     *
     * Updating `documentCaptures` or `documents` replaces the document set
     * — re-run the experiment afterwards via `runs.create` to refresh
     * metrics.
     */
    async update(
        workflowId: string,
        experimentId: string,
        {
            name,
            documentCaptures,
            documents,
            nConsensus,
        }: {
            name?: string;
            documentCaptures?: ExperimentDocumentCaptureRequest[];
            documents?: ExplicitExperimentDocumentRequest[];
            nConsensus?: NConsensusValue;
        },
        options?: RequestOptions
    ): Promise<ExperimentResponse> {
        const request = this.prepare_update(workflowId, experimentId, {
            name,
            documentCaptures,
            documents,
            nConsensus,
        });

        return this._fetchJson(ZExperimentResponse, {
            url: request.url,
            method: request.method,
            body: { ...request.body, ...((options?.body as Record<string, unknown>) || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Delete an experiment and its runs.
     */
    async delete(
        workflowId: string,
        experimentId: string,
        options?: RequestOptions
    ): Promise<void> {
        const request = this.prepare_delete(workflowId, experimentId);
        return this._fetchJson({
            url: request.url,
            method: request.method,
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Duplicate an experiment with the same documents and configuration.
     * The duplicate has no runs.
     */
    async duplicate(
        workflowId: string,
        experimentId: string,
        options?: RequestOptions
    ): Promise<ExperimentResponse> {
        const request = this.prepare_duplicate(workflowId, experimentId);
        return this._fetchJson(ZExperimentResponse, {
            url: request.url,
            method: request.method,
            body: { ...request.body, ...((options?.body as Record<string, unknown>) || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Cancel the latest pending or running run for this experiment.
     *
     * Returns 404 if no in-flight run exists.
     */
    async cancel(
        workflowId: string,
        experimentId: string,
        options?: RequestOptions
    ): Promise<CancelExperimentResponse> {
        const request = this.prepare_cancel(workflowId, experimentId);
        return this._fetchJson(ZCancelExperimentResponse, {
            url: request.url,
            method: request.method,
            body: { ...request.body, ...((options?.body as Record<string, unknown>) || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Get experiment metrics — consensus likelihoods (0.0 = low confidence,
     * 1.0 = high confidence).
     *
     * The response shape depends on `view`. Branch on the `view` field of
     * the result, or check for `error: "stale_metrics" | "no_metrics"` to
     * detect the two error envelopes the backend returns instead of throwing.
     *
     * Views:
     * - `summary` (default): overall score + block-specific diagnostics.
     * - `by_document`: one document → all targets sorted ascending. Requires `documentId`.
     * - `by_target`: one target → scores across all documents. Requires `targetPath`.
     * - `votes`: one document + one target → consensus rows with flat voter values.
     *   Requires both `documentId` and `targetPath`.
     */
    async getMetrics(
        {
            workflowId,
            experimentId,
            view = "summary",
            runId,
            documentId,
            targetPath,
            includePrior = true,
            priorRunId,
        }: {
            workflowId: string;
            experimentId: string;
            view?: ExperimentMetricView;
            runId?: string;
            documentId?: string;
            targetPath?: string;
            includePrior?: boolean;
            priorRunId?: string;
        },
        options?: RequestOptions
    ): Promise<ExperimentMetricsResponse> {
        const request = this.prepare_get_metrics(workflowId, experimentId, {
            view,
            runId,
            documentId,
            targetPath,
            includePrior,
            priorRunId,
        });
        return this._fetchJson(ZExperimentMetricsResponse, {
            url: request.url,
            method: request.method,
            params: { ...request.params, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async get_metrics(
        workflowId: string,
        experimentId: string,
        args: {
            view?: ExperimentMetricView;
            runId?: string;
            documentId?: string;
            targetPath?: string;
            includePrior?: boolean;
            priorRunId?: string;
        } = {},
        options?: RequestOptions
    ): Promise<ExperimentMetricsResponse> {
        return this.getMetrics({ workflowId, experimentId, ...args }, options);
    }

    /**
     * List blocks in the workflow that support experiments, with rolled-up
     * counts of how many experiments are attached, drifted, and stale.
     */
    async listEligibleBlocks(
        workflowId: string,
        options?: RequestOptions
    ): Promise<EligibleBlockListResponse> {
        const request = this.prepare_list_eligible_blocks(workflowId);
        return this._fetchJson(ZEligibleBlockListResponse, {
            url: request.url,
            method: request.method,
            params: options?.params,
            headers: options?.headers,
        });
    }

    async list_eligible_blocks(
        workflowId: string,
        options?: RequestOptions
    ): Promise<EligibleBlockListResponse> {
        return this.listEligibleBlocks(workflowId, options);
    }

    /**
     * Trigger one new run for every experiment attached to `blockId`.
     *
     * Returns 404 if no experiments are attached to the block.
     */
    async runBatch(
        {
            workflowId,
            blockId,
        }: {
            workflowId: string;
            blockId: string;
        },
        options?: RequestOptions
    ): Promise<RunBatchResponse> {
        const request = this.prepare_run_batch(workflowId, { blockId });
        return this._fetchJson(ZRunBatchResponse, {
            url: request.url,
            method: request.method,
            body: { ...request.body, ...((options?.body as Record<string, unknown>) || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async run_batch(
        workflowId: string,
        { blockId }: { blockId: string },
        options?: RequestOptions
    ): Promise<RunBatchResponse> {
        return this.runBatch({ workflowId, blockId }, options);
    }
}
