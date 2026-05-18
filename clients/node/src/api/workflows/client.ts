import { CompositionClient, RequestOptions } from "../../client.js";
import {
    PaginatedList,
    Workflow,
    WorkflowDiagnosisResponse,
    WorkflowResolvedSchemasResponse,
    WorkflowWithEntities,
    ZPaginatedList,
    ZWorkflow,
    ZWorkflowDiagnosisResponse,
    ZWorkflowResolvedSchemasResponse,
    ZWorkflowWithEntities,
} from "../../types.js";
import APIWorkflowRuns from "./runs/client.js";
import APIWorkflowReviews from "./reviews/client.js";
import APIWorkflowBlocks from "./blocks/client.js";
import APIWorkflowEdges from "./edges/client.js";
import APIWorkflowArtifacts from "./artifacts/client.js";
import APIWorkflowSpecs from "./specs/client.js";
import APIWorkflowTests from "./tests/client.js";
import APIWorkflowExperiments from "./experiments/client.js";

/**
 * Workflows API client for workflow operations.
 *
 * Sub-clients:
 * - runs: Workflow run operations
 * - reviews: HIL review overlay operations (list, get, approve, reject, escalate, edit, claim, release, waitFor)
 * - blocks: Workflow block CRUD + simulate
 * - edges: Workflow edge CRUD
 * - artifacts: Workflow artifact dereference operations
 * - specs: Declarative workflow YAML validation, planning, apply, and export
 * - tests: Workflow block-tests CRUD + execution + run history
 * - experiments: Consensus-experiments CRUD + per-run + metrics
 */
export default class APIWorkflows extends CompositionClient {
    public runs: APIWorkflowRuns;
    public reviews: APIWorkflowReviews;
    public blocks: APIWorkflowBlocks;
    public edges: APIWorkflowEdges;
    public artifacts: APIWorkflowArtifacts;
    public specs: APIWorkflowSpecs;
    public tests: APIWorkflowTests;
    public experiments: APIWorkflowExperiments;

    constructor(client: CompositionClient) {
        super(client);
        this.runs = new APIWorkflowRuns(this);
        this.reviews = new APIWorkflowReviews(this);
        this.blocks = new APIWorkflowBlocks(this);
        this.edges = new APIWorkflowEdges(this);
        this.artifacts = new APIWorkflowArtifacts(this);
        this.specs = new APIWorkflowSpecs(this);
        this.tests = new APIWorkflowTests(this);
        this.experiments = new APIWorkflowExperiments(this);
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
            emailTrigger,
        }: {
            name?: string;
            description?: string;
            emailTrigger?: {
                allowedSenders?: string[];
                allowedDomains?: string[];
            };
        },
        options?: RequestOptions
    ): Promise<Workflow> {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const body: Record<string, any> = {};
        if (name !== undefined) body.name = name;
        if (description !== undefined) body.description = description;
        if (emailTrigger !== undefined) {
            body.email_trigger = {};
            if (emailTrigger.allowedSenders !== undefined) body.email_trigger.allowed_senders = emailTrigger.allowedSenders;
            if (emailTrigger.allowedDomains !== undefined) body.email_trigger.allowed_domains = emailTrigger.allowedDomains;
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
     * console.log(wf.published?.version_id);
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
     * Get a workflow with all its entities (blocks and edges).
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

    prepare_get_resolved_schemas(workflowId: string): { url: string; method: string } {
        return {
            url: `/workflows/${workflowId}/resolved-schemas`,
            method: "GET",
        };
    }

    /**
     * Get graph-derived schemas for all current-draft blocks in a workflow.
     */
    async getResolvedSchemas(
        workflowId: string,
        options?: RequestOptions
    ): Promise<WorkflowResolvedSchemasResponse> {
        return this._fetchJson(ZWorkflowResolvedSchemasResponse, {
            url: `/workflows/${workflowId}/resolved-schemas`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async get_resolved_schemas(
        workflowId: string,
        options?: RequestOptions
    ): Promise<WorkflowResolvedSchemasResponse> {
        return this.getResolvedSchemas(workflowId, options);
    }

    prepare_list_snapshots(
        workflowId: string,
        limit?: number | null
    ): {
        url: string;
        method: string;
        params?: Record<string, unknown>;
    } {
        const params: Record<string, unknown> = {};
        if (limit !== undefined && limit !== null) params.limit = limit;
        return {
            url: `/workflows/${workflowId}/snapshots`,
            method: 'GET',
            ...(Object.keys(params).length > 0 ? { params } : {}),
        };
    }

    /**
     * List published snapshots for a workflow (newest first).
     *
     * Returns the canonical
     * `{"data": [...], "list_metadata": {"before": null, "after": null}}`
     * pagination envelope. `limit` bounds the page size (server default: 50,
     * max 100). ID pagination is not yet implemented for this endpoint.
     */
    async listSnapshots(
        workflowId: string,
        { limit }: { limit?: number } = {},
        options?: RequestOptions
    ): Promise<PaginatedList> {
        const params = Object.fromEntries(
            Object.entries({
                limit,
                ...(options?.params || {}),
            }).filter(([, value]) => value !== undefined)
        );
        return this._fetchJson(ZPaginatedList, {
            url: `/workflows/${workflowId}/snapshots`,
            method: "GET",
            params,
            headers: options?.headers,
        });
    }

    async list_snapshots(
        workflowId: string,
        opts: { limit?: number } = {},
        options?: RequestOptions
    ): Promise<PaginatedList> {
        return this.listSnapshots(workflowId, opts, options);
    }

    prepare_diagnose(
        workflowId: string,
        blocks: Array<Record<string, unknown>>,
        edges: Array<Record<string, unknown>>,
        rePropagate = true
    ): {
        url: string;
        method: string;
        body: {
            blocks: Array<Record<string, unknown>>;
            edges: Array<Record<string, unknown>>;
            re_propagate: boolean;
        };
    } {
        return {
            url: `/workflows/${workflowId}/diagnose-graph`,
            method: "POST",
            body: {
                blocks,
                edges,
                re_propagate: rePropagate,
            },
        };
    }

    /**
     * Diagnose the workflow's draft graph for structural issues.
     *
     * Fetches the persisted draft entities first (via `getEntities`) and
     * POSTs them to `/v1/workflows/{workflow_id}/diagnose-graph`. Returns
     * a list of `issues` (errors must be fixed before publish; warnings
     * are advisory) and `stats`.
     *
     * To diagnose an in-memory editor graph that hasn't been saved yet,
     * call `diagnoseGraph` directly with your own `blocks` / `edges`.
     */
    async diagnose(
        workflowId: string,
        { rePropagate = true }: { rePropagate?: boolean } = {},
        options?: RequestOptions
    ): Promise<WorkflowDiagnosisResponse> {
        const entities = await this.getEntities(workflowId, options);
        const blocks = entities.blocks.map((block) => ({
            id: block.id,
            type: block.type,
            label: block.label,
            config: block.config,
            position: { x: block.position_x, y: block.position_y },
            width: block.width,
            height: block.height,
            parent_id: block.parent_id,
        }));
        const edges = entities.edges.map((edge) => ({
            id: edge.id,
            source: edge.source_block,
            target: edge.target_block,
            source_handle: edge.source_handle,
            target_handle: edge.target_handle,
        }));
        return this.diagnoseGraph(workflowId, { blocks, edges, rePropagate }, options);
    }

    /**
     * Lower-level: POST blocks + edges directly to the diagnose-graph
     * endpoint. Use this when you have an in-memory editor graph that
     * hasn't been saved yet.
     */
    async diagnoseGraph(
        workflowId: string,
        {
            blocks,
            edges,
            rePropagate = true,
        }: {
            blocks: Array<Record<string, unknown>>;
            edges: Array<Record<string, unknown>>;
            rePropagate?: boolean;
        },
        options?: RequestOptions
    ): Promise<WorkflowDiagnosisResponse> {
        return this._fetchJson(ZWorkflowDiagnosisResponse, {
            url: `/workflows/${workflowId}/diagnose-graph`,
            method: "POST",
            body: {
                blocks,
                edges,
                re_propagate: rePropagate,
                ...((options?.body as Record<string, unknown>) || {}),
            },
            params: options?.params,
            headers: options?.headers,
        });
    }
}
