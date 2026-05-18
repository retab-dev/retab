import { CompositionClient, RequestOptions } from "../../../../client.js";
import {
    ExecuteWorkflowTestsResponse,
    WorkflowTestExecutionRunResults,
    WorkflowTestRunListResponse,
    WorkflowTestRunRecord,
    ZExecuteWorkflowTestsResponse,
    ZWorkflowTestExecutionRunResults,
    ZWorkflowTestRunListResponse,
    ZWorkflowTestRunRecord,
} from "../types.js";

/**
 * Read-only access to a workflow test's run-record history.
 */
export default class APIWorkflowTestRuns extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    /**
     * List the most recent run records for a workflow test, newest first.
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
    ): Promise<WorkflowTestRunListResponse> {
        return this._fetchJson(ZWorkflowTestRunListResponse, {
            url: `/workflows/${workflowId}/tests/${testId}/runs`,
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
            url: `/workflows/${workflowId}/tests/${testId}/runs/${runId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Fetch parent execution-run status returned by `tests.execute`.
     */
    async getExecution(
        {
            workflowId,
            runId,
        }: {
            workflowId: string;
            runId: string;
        },
        options?: RequestOptions
    ): Promise<ExecuteWorkflowTestsResponse> {
        return this._fetchJson(ZExecuteWorkflowTestsResponse, {
            url: `/workflows/${workflowId}/tests/runs/${runId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async get_execution(
        args: {
            workflowId: string;
            runId: string;
        },
        options?: RequestOptions
    ): Promise<ExecuteWorkflowTestsResponse> {
        return this.getExecution(args, options);
    }

    /**
     * Fetch per-test results for a parent execution run.
     */
    async results(
        {
            workflowId,
            runId,
        }: {
            workflowId: string;
            runId: string;
        },
        options?: RequestOptions
    ): Promise<WorkflowTestExecutionRunResults> {
        return this._fetchJson(ZWorkflowTestExecutionRunResults, {
            url: `/workflows/${workflowId}/tests/runs/${runId}/results`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }
}
