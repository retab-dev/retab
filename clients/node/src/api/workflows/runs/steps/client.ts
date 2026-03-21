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
 * Usage: `client.workflows.runs.steps.get(runId, nodeId)` or `client.workflows.runs.steps.getMany(runId, nodeIds)`
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
        nodeId: string,
        options?: RequestOptions
    ): Promise<StepOutputResponse> {
        return this._fetchJson(ZStepOutputResponse, {
            url: `/workflows/runs/${runId}/steps/${nodeId}`,
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
     *     console.log(`${step.node_id}: ${step.status}`);
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
     * Batch-get step outputs for multiple nodes in a single request.
     *
     * @param runId - The workflow run ID
     * @param nodeIds - List of node IDs to fetch outputs for
     * @returns Step outputs keyed by node ID
     *
     * @example
     * ```typescript
     * const batch = await client.workflows.runs.steps.getMany("run_abc123", ["extract-1", "classifier-1"]);
     * console.log(batch.outputs["extract-1"]?.handle_outputs);
     * ```
     */
    async getMany(
        runId: string,
        nodeIds: string[],
        options?: RequestOptions
    ): Promise<StepOutputsBatchResponse> {
        return this._fetchJson(ZStepOutputsBatchResponse, {
            url: `/workflows/runs/${runId}/steps/batch`,
            method: "POST",
            body: { node_ids: nodeIds, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async get_many(
        runId: string,
        nodeIds: string[],
        options?: RequestOptions
    ): Promise<StepOutputsBatchResponse> {
        return this.getMany(runId, nodeIds, options);
    }

    /**
     * Fetch outputs for all steps in a workflow run in one call.
     *
     * Internally fetches the run to discover step node IDs, then batch-fetches all outputs.
     *
     * @param runId - The workflow run ID
     * @returns Step outputs keyed by node ID
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

        const nodeIds = workflowRun.steps.map((s) => s.node_id);
        if (nodeIds.length === 0) {
            return { outputs: {} };
        }

        return this.getMany(workflowRun.id, nodeIds, options);
    }

    async get_all(
        run: WorkflowRun | string,
        options?: RequestOptions
    ): Promise<StepOutputsBatchResponse> {
        return this.getAll(run, options);
    }
}
