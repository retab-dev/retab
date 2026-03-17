import * as z from "zod";
import { CompositionClient, RequestOptions } from "../../../client.js";
import { WorkflowBlock, ZWorkflowBlock } from "../../../types.js";

/**
 * Workflow Blocks API client for managing blocks (nodes) in a workflow graph.
 *
 * Usage: `client.workflows.blocks.list(workflowId)`
 */
export default class APIWorkflowBlocks extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    /**
     * List all blocks for a workflow.
     *
     * @param workflowId - The workflow ID
     * @param subflowId - Optional: filter by subflow ID
     */
    async list(
        workflowId: string,
        { subflowId }: { subflowId?: string } = {},
        options?: RequestOptions
    ): Promise<WorkflowBlock[]> {
        const params = Object.fromEntries(
            Object.entries({
                subflow_id: subflowId,
                ...(options?.params || {}),
            }).filter(([_, v]) => v !== undefined)
        );

        return this._fetchJson(z.array(ZWorkflowBlock), {
            url: `/workflows/${workflowId}/blocks`,
            method: "GET",
            params,
            headers: options?.headers,
        });
    }

    /**
     * Get a single block by ID.
     */
    async get(workflowId: string, blockId: string, options?: RequestOptions): Promise<WorkflowBlock> {
        return this._fetchJson(ZWorkflowBlock, {
            url: `/workflows/${workflowId}/blocks/${blockId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Create a new block in a workflow.
     *
     * @example
     * ```typescript
     * const block = await client.workflows.blocks.create("wf_abc123", {
     *     id: "extract-1",
     *     type: "extract",
     *     label: "Extract Invoice Data",
     *     config: { json_schema: { ... } },
     * });
     * ```
     */
    async create(
        workflowId: string,
        {
            id,
            type,
            label = "",
            positionX = 0,
            positionY = 0,
            width,
            height,
            config,
            subflowId,
            parentId,
        }: {
            id: string;
            type: string;
            label?: string;
            positionX?: number;
            positionY?: number;
            width?: number;
            height?: number;
            config?: Record<string, unknown>;
            subflowId?: string;
            parentId?: string;
        },
        options?: RequestOptions
    ): Promise<WorkflowBlock> {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const body: Record<string, any> = {
            id,
            type,
            label,
            position_x: positionX,
            position_y: positionY,
        };
        if (width !== undefined) body.width = width;
        if (height !== undefined) body.height = height;
        if (config !== undefined) body.config = config;
        if (subflowId !== undefined) body.subflow_id = subflowId;
        if (parentId !== undefined) body.parent_id = parentId;

        return this._fetchJson(ZWorkflowBlock, {
            url: `/workflows/${workflowId}/blocks`,
            method: "POST",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Create multiple blocks in a single request.
     *
     * @param workflowId - The workflow ID
     * @param blocks - Array of block definitions (each with at least `id` and `type`)
     */
    async createBatch(
        workflowId: string,
        blocks: Array<{
            id: string;
            type: string;
            label?: string;
            position_x?: number;
            position_y?: number;
            width?: number;
            height?: number;
            config?: Record<string, unknown>;
            subflow_id?: string;
            parent_id?: string;
        }>,
        options?: RequestOptions
    ): Promise<WorkflowBlock[]> {
        return this._fetchJson(z.array(ZWorkflowBlock), {
            url: `/workflows/${workflowId}/blocks/batch`,
            method: "POST",
            body: blocks,
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Update a block with partial data. Only provided fields are updated.
     */
    async update(
        workflowId: string,
        blockId: string,
        {
            label,
            positionX,
            positionY,
            width,
            height,
            config,
            subflowId,
            parentId,
        }: {
            label?: string;
            positionX?: number;
            positionY?: number;
            width?: number;
            height?: number;
            config?: Record<string, unknown>;
            subflowId?: string;
            parentId?: string;
        },
        options?: RequestOptions
    ): Promise<WorkflowBlock> {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const body: Record<string, any> = {};
        if (label !== undefined) body.label = label;
        if (positionX !== undefined) body.position_x = positionX;
        if (positionY !== undefined) body.position_y = positionY;
        if (width !== undefined) body.width = width;
        if (height !== undefined) body.height = height;
        if (config !== undefined) body.config = config;
        if (subflowId !== undefined) body.subflow_id = subflowId;
        if (parentId !== undefined) body.parent_id = parentId;

        return this._fetchJson(ZWorkflowBlock, {
            url: `/workflows/${workflowId}/blocks/${blockId}`,
            method: "PATCH",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Delete a block and any edges connected to it.
     */
    async delete(workflowId: string, blockId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/workflows/${workflowId}/blocks/${blockId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }
}
