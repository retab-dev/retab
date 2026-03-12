import { CompositionClient, RequestOptions } from "../../client.js";
import { PaginatedList, Workflow, ZPaginatedList, ZWorkflow } from "../../types.js";
import APIWorkflowRuns from "./runs/client.js";

/**
 * Workflows API client for workflow operations.
 *
 * Sub-clients:
 * - runs: Workflow run operations
 */
export default class APIWorkflows extends CompositionClient {
    public runs: APIWorkflowRuns;

    constructor(client: CompositionClient) {
        super(client);
        this.runs = new APIWorkflowRuns(this);
    }

    /**
     * Get a workflow by ID.
     */
    async get(workflowId: string, options?: RequestOptions): Promise<Workflow> {
        return this._fetchJson(ZWorkflow, {
            url: `/workflows/${workflowId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * List workflows with pagination.
     */
    async list(
        {
            before,
            after,
            limit = 10,
            order = "desc",
            sortBy,
            fields,
        }: {
            before?: string;
            after?: string;
            limit?: number;
            order?: "asc" | "desc";
            sortBy?: string;
            fields?: string;
        } = {},
        options?: RequestOptions
    ): Promise<PaginatedList> {
        const params = Object.fromEntries(
            Object.entries({
                before,
                after,
                limit,
                order,
                sort_by: sortBy,
                fields,
                ...(options?.params || {}),
            }).filter(([_, value]) => value !== undefined)
        );

        return this._fetchJson(ZPaginatedList, {
            url: "/workflows",
            method: "GET",
            params,
            headers: options?.headers,
        });
    }
}
