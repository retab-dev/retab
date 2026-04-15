import { CompositionClient, RequestOptions } from "../../../../client.js";
import {
    StepOutputResponse,
    ZStepOutputResponse,
    StepOutputsBatchResponse,
    ZStepOutputsBatchResponse,
    WorkflowRunStep,
    ZWorkflowRunStep,
    WorkflowRun,
    ZWorkflowRun,
} from "../../../../types.js";
import * as z from "zod";

/**
 * Workflow Run Steps API client for accessing step-level outputs.
 *
 * Usage: `client.workflows.runs.steps.get(runId, blockId)` or `client.workflows.runs.steps.getMany(runId, blockIds)`
 */
export default class APIWorkflowRunSteps extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    /**
     * Get step status and handle data for a specific step in a workflow run.
     *
     * @example
     * ```typescript
     * const step = await client.workflows.runs.steps.get("run_abc123", "extract-node-1");
     * const data = step.handle_outputs?.["output-json-0"]?.data;
     * ```
     */
    async get(
        runId: string,
        blockId: string,
        options?: RequestOptions
    ): Promise<StepOutputResponse> {
        return this._fetchJson(ZStepOutputResponse, {
            url: `/workflows/runs/${runId}/steps/${blockId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * List all persisted step documents for a workflow run.
     *
     * @example
     * ```typescript
     * const steps = await client.workflows.runs.steps.list("run_abc123");
     * for (const step of steps) {
     *     console.log(`${step.block_id}: ${step.status}`);
     * }
     * ```
     */
    async list(
        runId: string,
        options?: RequestOptions
    ): Promise<WorkflowRunStep[]> {
        return this._fetchJson(z.array(ZWorkflowRunStep), {
            url: `/workflows/runs/${runId}/steps`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Batch-get step outputs for multiple blocks in a single request.
     *
     * @param runId - The workflow run ID
     * @param blockIds - List of block IDs to fetch outputs for
     * @returns Step outputs keyed by block ID
     *
     * @example
     * ```typescript
     * const batch = await client.workflows.runs.steps.getMany("run_abc123", ["extract-1", "classifier-1"]);
     * console.log(batch.outputs["extract-1"]?.handle_outputs);
     * ```
     */
    async getMany(
        runId: string,
        blockIds: string[],
        options?: RequestOptions
    ): Promise<StepOutputsBatchResponse> {
        return this._fetchJson(ZStepOutputsBatchResponse, {
            url: `/workflows/runs/${runId}/steps/batch`,
            method: "POST",
            body: { block_ids: blockIds, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async get_many(
        runId: string,
        blockIds: string[],
        options?: RequestOptions
    ): Promise<StepOutputsBatchResponse> {
        return this.getMany(runId, blockIds, options);
    }

    /**
     * Fetch outputs for all steps in a workflow run in one call.
     *
     * Internally fetches the run to discover step block IDs, then batch-fetches all outputs.
     *
     * @param runId - The workflow run ID
     * @returns Step outputs keyed by block ID
     *
     * @example
     * ```typescript
     * const allOutputs = await client.workflows.runs.steps.getAll("run_abc123");
     * console.log(allOutputs.outputs["extract-1"]?.handle_outputs);
     * ```
     */
    async getAll(
        run: WorkflowRun | string,
        options?: RequestOptions
    ): Promise<StepOutputsBatchResponse> {
        const workflowRun = typeof run === "string"
            ? await this._fetchJson(ZWorkflowRun, {
                url: `/workflows/runs/${run}`,
                method: "GET",
                params: options?.params,
                headers: options?.headers,
            })
            : run;

        const blockIds = workflowRun.steps.map((s) => s.block_id);
        if (blockIds.length === 0) {
            return { outputs: {} };
        }

        return this.getMany(workflowRun.id, blockIds, options);
    }

    async get_all(
        run: WorkflowRun | string,
        options?: RequestOptions
    ): Promise<StepOutputsBatchResponse> {
        return this.getAll(run, options);
    }
}
