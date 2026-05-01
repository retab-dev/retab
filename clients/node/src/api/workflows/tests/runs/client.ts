import { CompositionClient, RequestOptions } from "../../../../client.js";
import {
    BlockTestRunListResponse,
    WorkflowTestRunRecord,
    ZBlockTestRunListResponse,
    ZWorkflowTestRunRecord,
} from "../types.js";

/**
 * Read-only access to a block test's run-record history.
 */
export default class APIWorkflowTestRuns extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    /**
     * List the most recent run records for a block test, newest first.
     *
     * @example
     * ```typescript
     * const runs = await client.workflows.tests.runs.list({
     *     workflowId: "wf_abc123",
     *     testId: "wfnodetest_abc",
     *     limit: 10,
     * });
     * ```
     */
    async list(
        {
            workflowId,
            testId,
            limit = 20,
        }: {
            workflowId: string;
            testId: string;
            limit?: number;
        },
        options?: RequestOptions
    ): Promise<BlockTestRunListResponse> {
        return this._fetchJson(ZBlockTestRunListResponse, {
            url: `/workflows/${workflowId}/block-tests/${testId}/runs`,
            method: "GET",
            // `options.params` wins so callers can override the typed
            // `limit` arg via the escape hatch — matches sibling
            // `APIWorkflowRuns.list` precedence (runs/client.ts).
            params: { limit, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    /**
     * Fetch a single immutable run record.
     */
    async get(
        {
            workflowId,
            testId,
            runId,
        }: {
            workflowId: string;
            testId: string;
            runId: string;
        },
        options?: RequestOptions
    ): Promise<WorkflowTestRunRecord> {
        return this._fetchJson(ZWorkflowTestRunRecord, {
            url: `/workflows/${workflowId}/block-tests/${testId}/runs/${runId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }
}
