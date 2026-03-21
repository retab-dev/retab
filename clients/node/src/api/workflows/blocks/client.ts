import * as z from "zod";
import { CompositionClient, RequestOptions } from "../../../client.js";
import {
    WorkflowBlock,
    WorkflowBlockCreateRequest,
    WorkflowBlockUpdateRequest,
    ZWorkflowBlock,
} from "../../../types.js";

type LegacyWorkflowBlockCreateRequest = {
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
};

function serializeBlockCreateRequest(
    request: WorkflowBlockCreateRequest | LegacyWorkflowBlockCreateRequest
): Record<string, unknown> {
    const legacyRequest = request as Partial<LegacyWorkflowBlockCreateRequest>;
    return {
        id: request.id,
        type: request.type,
        label: request.label ?? "",
        position_x: "positionX" in request ? (request.positionX ?? 0) : (legacyRequest.position_x ?? 0),
        position_y: "positionY" in request ? (request.positionY ?? 0) : (legacyRequest.position_y ?? 0),
        width: request.width,
        height: request.height,
        config: request.config,
        subflow_id: "subflowId" in request ? request.subflowId : legacyRequest.subflow_id,
        parent_id: "parentId" in request ? request.parentId : legacyRequest.parent_id,
    };
}

function serializeBlockUpdateRequest(request: WorkflowBlockUpdateRequest): Record<string, unknown> {
    return {
        label: request.label,
        position_x: request.positionX,
        position_y: request.positionY,
        width: request.width,
        height: request.height,
        config: request.config,
        subflow_id: request.subflowId,
        parent_id: request.parentId,
    };
}

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
        request: WorkflowBlockCreateRequest,
        options?: RequestOptions
    ): Promise<WorkflowBlock> {
        return this._fetchJson(ZWorkflowBlock, {
            url: `/workflows/${workflowId}/blocks`,
            method: "POST",
            body: { ...serializeBlockCreateRequest(request), ...(options?.body as Record<string, unknown> || {}) },
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
        blocks: Array<WorkflowBlockCreateRequest | LegacyWorkflowBlockCreateRequest>,
        options?: RequestOptions
    ): Promise<WorkflowBlock[]> {
        return this._fetchJson(z.array(ZWorkflowBlock), {
            url: `/workflows/${workflowId}/blocks/batch`,
            method: "POST",
            body: blocks.map((block) => serializeBlockCreateRequest(block)),
            params: options?.params,
            headers: options?.headers,
        });
    }

    async create_batch(
        workflowId: string,
        blocks: Array<WorkflowBlockCreateRequest | LegacyWorkflowBlockCreateRequest>,
        options?: RequestOptions
    ): Promise<WorkflowBlock[]> {
        return this.createBatch(workflowId, blocks, options);
    }

    /**
     * Update a block with partial data. Only provided fields are updated.
     */
    async update(
        workflowId: string,
        blockId: string,
        request: WorkflowBlockUpdateRequest,
        options?: RequestOptions
    ): Promise<WorkflowBlock> {
        return this._fetchJson(ZWorkflowBlock, {
            url: `/workflows/${workflowId}/blocks/${blockId}`,
            method: "PATCH",
            body: { ...serializeBlockUpdateRequest(request), ...(options?.body as Record<string, unknown> || {}) },
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
