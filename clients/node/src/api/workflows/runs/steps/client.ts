import { CompositionClient, RequestOptions } from "../../../../client.js";
import {
    PaginatedList,
    StepExecutionResponse,
    ZPaginatedList,
    ZStepExecutionResponse,
} from "../../../../types.js";

/**
 * Workflow Run Steps API client for accessing step-level execution artifacts.
 *
 * Usage: `client.workflows.runs.steps.get(runId, blockId)`
 */
export default class APIWorkflowRunSteps extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    /**
     * Get full step execution records with handle inputs and outputs.
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
    ): Promise<StepExecutionResponse> {
        if (typeof blockId !== "string" || blockId.length === 0) {
            throw new TypeError("blockId is required");
        }
        return this._fetchJson(ZStepExecutionResponse, {
            url: `/workflows/runs/${runId}/steps/${blockId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * List all persisted step documents for a workflow run.
     *
     * Returns the canonical
     * `{"data": [...], "list_metadata": {"before": null, "after": null}}`
     * pagination envelope. Cursor pagination is not yet implemented; iterate
     * `result.data` directly.
     *
     * @example
     * ```typescript
     * const steps = await client.workflows.runs.steps.list("run_abc123");
     * for (const step of steps.data) {
     *     console.log(`${step.block_id}: ${step.lifecycle.status}`);
     * }
     * ```
     */
    async list(
        runId: string,
        options?: RequestOptions
    ): Promise<PaginatedList> {
        return this._fetchJson(ZPaginatedList, {
            url: `/workflows/runs/${runId}/steps`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

}
