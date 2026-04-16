import { CompositionClient, RequestOptions } from "../../client.js";
import { PaginatedList, Workflow, WorkflowWithEntities, ZPaginatedList, ZWorkflow, ZWorkflowWithEntities } from "../../types.js";
import APIWorkflowRuns from "./runs/client.js";
import APIWorkflowBlocks from "./blocks/client.js";
import APIWorkflowEdges from "./edges/client.js";

/**
 * Workflows API client for workflow operations.
 *
 * Sub-clients:
 * - runs: Workflow run operations
 * - blocks: Workflow block CRUD
 * - edges: Workflow edge CRUD
 */
export default class APIWorkflows extends CompositionClient {
    public runs: APIWorkflowRuns;
    public blocks: APIWorkflowBlocks;
    public edges: APIWorkflowEdges;

    constructor(client: CompositionClient) {
        super(client);
        this.runs = new APIWorkflowRuns(this);
        this.blocks = new APIWorkflowBlocks(this);
        this.edges = new APIWorkflowEdges(this);
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

    /**
     * Create a new workflow.
     *
     * @param name - Workflow name (default: "Untitled Workflow")
     * @param description - Workflow description (default: "")
     * @returns The created workflow (unpublished, with a default start block)
     *
     * @example
     * ```typescript
     * const wf = await client.workflows.create({ name: "Invoice Extractor" });
     * ```
     */
    async create(
        {
            name = "Untitled Workflow",
            description = "",
        }: {
            name?: string;
            description?: string;
        } = {},
        options?: RequestOptions
    ): Promise<Workflow> {
        return this._fetchJson(ZWorkflow, {
            url: "/workflows",
            method: "POST",
            body: { name, description, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Update a workflow's metadata.
     *
     * @param workflowId - The workflow ID
     * @param fields - Fields to update (only provided fields are changed)
     *
     * @example
     * ```typescript
     * const wf = await client.workflows.update("wf_abc123", { name: "New Name" });
     * ```
     */
    async update(
        workflowId: string,
        {
            name,
            description,
            emailSendersWhitelist,
            emailDomainsWhitelist,
        }: {
            name?: string;
            description?: string;
            emailSendersWhitelist?: string[];
            emailDomainsWhitelist?: string[];
        },
        options?: RequestOptions
    ): Promise<Workflow> {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const body: Record<string, any> = {};
        if (name !== undefined) body.name = name;
        if (description !== undefined) body.description = description;
        if (emailSendersWhitelist !== undefined || emailDomainsWhitelist !== undefined) {
            body.email_trigger = {};
            if (emailSendersWhitelist !== undefined) body.email_trigger.allowed_senders = emailSendersWhitelist;
            if (emailDomainsWhitelist !== undefined) body.email_trigger.allowed_domains = emailDomainsWhitelist;
        }

        return this._fetchJson(ZWorkflow, {
            url: `/workflows/${workflowId}`,
            method: "PATCH",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Delete a workflow and all its associated entities (blocks, edges, snapshots).
     */
    async delete(workflowId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/workflows/${workflowId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Publish a workflow's current draft as a new version.
     *
     * @example
     * ```typescript
     * const wf = await client.workflows.publish("wf_abc123", { description: "v1" });
     * console.log(wf.published?.snapshot_id);
     * ```
     */
    async publish(
        workflowId: string,
        { description = "" }: { description?: string } = {},
        options?: RequestOptions
    ): Promise<Workflow> {
        return this._fetchJson(ZWorkflow, {
            url: `/workflows/${workflowId}/publish`,
            method: "POST",
            body: { description, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Duplicate a workflow with all its blocks and edges.
     *
     * @returns The new duplicated workflow (unpublished)
     */
    async duplicate(workflowId: string, options?: RequestOptions): Promise<Workflow> {
        return this._fetchJson(ZWorkflow, {
            url: `/workflows/${workflowId}/duplicate`,
            method: "POST",
            body: { ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Get a workflow with all its entities (blocks, edges, subflows).
     *
     * Use the returned `blocks` array to discover start blocks:
     * ```typescript
     * const entities = await client.workflows.getEntities("wf_abc123");
     * const startBlocks = entities.blocks.filter(b => b.type === "start");
     * ```
     */
    async getEntities(workflowId: string, options?: RequestOptions): Promise<WorkflowWithEntities> {
        return this._fetchJson(ZWorkflowWithEntities, {
            url: `/workflows/${workflowId}/entities`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async get_entities(workflowId: string, options?: RequestOptions): Promise<WorkflowWithEntities> {
        return this.getEntities(workflowId, options);
    }
}
