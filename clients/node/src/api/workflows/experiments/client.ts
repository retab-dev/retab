import { CompositionClient, RequestOptions } from "../../../client.js";
import APIWorkflowExperimentRuns from "./runs/client.js";
import {
    EligibleBlockListResponse,
    ExperimentDocumentCaptureRequest,
    ExperimentList,
    ExperimentResponse,
    ExplicitExperimentDocumentRequest,
    NConsensusValue,
    ZEligibleBlockListResponse,
    ZExperimentList,
    ZExperimentResponse,
} from "./types.js";

/**
 * Workflow experiments API client. Mirrors the REST endpoints under
 * `/v1/workflows/{workflow_id}/experiments` and the Python SDK's
 * `client.workflows.experiments.*`.
 *
 * The MCP tools (`experiments_create`, `experiments_runs_create`,
 * `experiments_runs_metrics_get`, …) call the same routes — this resource is
 * the SDK-side equivalent.
 *
 * Sub-clients:
 * - runs: run lifecycle, per-document results, and metrics.
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
 * console.log(run.id, run.lifecycle.status);
 *
 * const metrics = await client.workflows.experiments.runs.metrics.get({
 *     runId: run.id,
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

    prepare_list_eligible_blocks(workflowId: string): { url: string; method: string } {
        return {
            url: `/workflows/${workflowId}/experiments/eligible-blocks`,
            method: "GET",
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

}
