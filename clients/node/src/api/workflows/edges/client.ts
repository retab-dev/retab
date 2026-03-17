import * as z from "zod";
import { CompositionClient, RequestOptions } from "../../../client.js";
import { WorkflowEdge, ZWorkflowEdge } from "../../../types.js";

/**
 * Workflow Edges API client for managing connections between blocks.
 *
 * Usage: `client.workflows.edges.list(workflowId)`
 */
export default class APIWorkflowEdges extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    /**
     * List all edges for a workflow.
     *
     * @param workflowId - The workflow ID
     * @param sourceBlock - Optional: filter by source block ID
     * @param targetBlock - Optional: filter by target block ID
     */
    async list(
        workflowId: string,
        { sourceBlock, targetBlock }: { sourceBlock?: string; targetBlock?: string } = {},
        options?: RequestOptions
    ): Promise<WorkflowEdge[]> {
        const params = Object.fromEntries(
            Object.entries({
                source_block: sourceBlock,
                target_block: targetBlock,
                ...(options?.params || {}),
            }).filter(([_, v]) => v !== undefined)
        );

        return this._fetchJson(z.array(ZWorkflowEdge), {
            url: `/workflows/${workflowId}/edges`,
            method: "GET",
            params,
            headers: options?.headers,
        });
    }

    /**
     * Get a single edge by ID.
     */
    async get(workflowId: string, edgeId: string, options?: RequestOptions): Promise<WorkflowEdge> {
        return this._fetchJson(ZWorkflowEdge, {
            url: `/workflows/${workflowId}/edges/${edgeId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Create a new edge connecting two blocks.
     *
     * @example
     * ```typescript
     * const edge = await client.workflows.edges.create("wf_abc123", {
     *     id: "edge-1",
     *     sourceBlock: "start-1",
     *     targetBlock: "extract-1",
     *     sourceHandle: "output-file-0",
     *     targetHandle: "input-file-0",
     * });
     * ```
     */
    async create(
        workflowId: string,
        {
            id,
            sourceBlock,
            targetBlock,
            sourceHandle,
            targetHandle,
        }: {
            id: string;
            sourceBlock: string;
            targetBlock: string;
            sourceHandle?: string;
            targetHandle?: string;
        },
        options?: RequestOptions
    ): Promise<WorkflowEdge> {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const body: Record<string, any> = {
            id,
            source_block: sourceBlock,
            target_block: targetBlock,
        };
        if (sourceHandle !== undefined) body.source_handle = sourceHandle;
        if (targetHandle !== undefined) body.target_handle = targetHandle;

        return this._fetchJson(ZWorkflowEdge, {
            url: `/workflows/${workflowId}/edges`,
            method: "POST",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Create multiple edges in a single request.
     *
     * @param workflowId - The workflow ID
     * @param edges - Array of edge definitions
     */
    async createBatch(
        workflowId: string,
        edges: Array<{
            id: string;
            source_block: string;
            target_block: string;
            source_handle?: string;
            target_handle?: string;
        }>,
        options?: RequestOptions
    ): Promise<WorkflowEdge[]> {
        return this._fetchJson(z.array(ZWorkflowEdge), {
            url: `/workflows/${workflowId}/edges/batch`,
            method: "POST",
            body: edges,
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Delete an edge.
     */
    async delete(workflowId: string, edgeId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/workflows/${workflowId}/edges/${edgeId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Delete all edges for a workflow.
     */
    async deleteAll(workflowId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/workflows/${workflowId}/edges`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }
}
