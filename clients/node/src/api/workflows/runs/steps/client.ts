import { CompositionClient, RequestOptions } from "../../../../client.js";
import {
    StepOutputResponse,
    ZStepOutputResponse,
    StepOutputsBatchResponse,
    ZStepOutputsBatchResponse,
} from "../../../../types.js";

/**
 * Workflow Run Steps API client for accessing step-level outputs.
 *
 * Usage: `client.workflows.runs.steps.get(runId, nodeId)`
 */
export default class APIWorkflowRunSteps extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    /**
     * Get the full output data for a specific step in a workflow run.
     *
     * @param runId - The ID of the workflow run
     * @param nodeId - The ID of the node/step to get output for
     * @param options - Optional request options
     * @returns The step output with handle outputs and inputs
     *
     * @example
     * ```typescript
     * const step = await client.workflows.runs.steps.get("run_abc123", "extract-node-1");
     * console.log(`Step status: ${step.status}, output:`, step.output);
     * ```
     */
    async get(
        runId: string,
        nodeId: string,
        options?: RequestOptions
    ): Promise<StepOutputResponse> {
        return this._fetchJson(ZStepOutputResponse, {
            url: `/v1/workflows/runs/${runId}/steps/${nodeId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Get outputs for multiple steps in a single request.
     *
     * @param runId - The ID of the workflow run
     * @param nodeIds - List of node IDs to fetch outputs for (max 1000)
     * @param options - Optional request options
     * @returns Step outputs keyed by node ID
     *
     * @example
     * ```typescript
     * const batch = await client.workflows.runs.steps.getBatch(
     *     "run_abc123",
     *     ["extract-1", "extract-2", "split-1"]
     * );
     * for (const [nodeId, step] of Object.entries(batch.outputs)) {
     *     console.log(`${nodeId}: ${step.status}`);
     * }
     * ```
     */
    async getBatch(
        runId: string,
        nodeIds: string[],
        options?: RequestOptions
    ): Promise<StepOutputsBatchResponse> {
        return this._fetchJson(ZStepOutputsBatchResponse, {
            url: `/v1/workflows/runs/${runId}/steps/batch`,
            method: "POST",
            body: { node_ids: nodeIds, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
}
