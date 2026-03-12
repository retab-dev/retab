import { CompositionClient, RequestOptions } from "../../../../client.js";
import {
    StepOutputResponse,
    ZStepOutputResponse,
    WorkflowRunStep,
    ZWorkflowRunStep,
} from "../../../../types.js";
import * as z from "zod";

/**
 * Workflow Run Steps API client for accessing step-level outputs.
 *
 * Usage: `client.workflows.runs.steps.get(runId, nodeId)` or `client.workflows.runs.steps.list(runId)`
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
     * List all persisted step documents for a workflow run.
     *
     * @param runId - The ID of the workflow run
     * @param options - Optional request options
     * @returns All step documents for the run
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
            url: `/v1/workflows/runs/${runId}/steps`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }
}
