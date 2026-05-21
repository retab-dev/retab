import { CompositionClient, RequestOptions } from '../../../client.js';
import {
  PaginatedList,
  WorkflowBlock,
  WorkflowBlockCreateRequest,
  WorkflowBlockUpdateRequest,
  ZPaginatedList,
  ZWorkflowBlock,
} from '../../../types.js';

function serializeBlockCreateRequest(request: WorkflowBlockCreateRequest): Record<string, unknown> {
  return {
    id: request.id,
    type: request.type,
    label: request.label ?? '',
    position_x: request.positionX ?? 0,
    position_y: request.positionY ?? 0,
    width: request.width,
    height: request.height,
    config: request.config,
    parent_id: request.parentId,
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
    parent_id: request.parentId,
  };
}

/**
 * Workflow Blocks API client for managing blocks in a workflow graph.
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
   * Returns the canonical
   * `{"data": [...], "list_metadata": {"before": null, "after": null}}`
   * pagination envelope. ID pagination is not yet implemented for this
   * endpoint; `list_metadata` is always `{before: null, after: null}`.
   */
  async list(workflowId: string, options?: RequestOptions): Promise<PaginatedList> {
    return this._fetchJson(ZPaginatedList, {
      url: `/workflows/blocks?workflow_id=${workflowId}`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Get a single block by ID.
   */
  async get(workflowId: string, blockId: string, options?: RequestOptions): Promise<WorkflowBlock> {
    return this._fetchJson(ZWorkflowBlock, {
      url: `/workflows/blocks/${blockId}?workflow_id=${workflowId}`,
      method: 'GET',
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
      url: `/workflows/blocks?workflow_id=${workflowId}`,
      method: 'POST',
      body: {
        ...serializeBlockCreateRequest(request),
        ...((options?.body as Record<string, unknown>) || {}),
      },
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
    request: WorkflowBlockUpdateRequest,
    options?: RequestOptions
  ): Promise<WorkflowBlock> {
    return this._fetchJson(ZWorkflowBlock, {
      url: `/workflows/blocks/${blockId}?workflow_id=${workflowId}`,
      method: 'PATCH',
      body: {
        ...serializeBlockUpdateRequest(request),
        ...((options?.body as Record<string, unknown>) || {}),
      },
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Delete a block and any edges connected to it.
   */
  async delete(workflowId: string, blockId: string, options?: RequestOptions): Promise<void> {
    return this._fetchJson({
      url: `/workflows/blocks/${blockId}?workflow_id=${workflowId}`,
      method: 'DELETE',
      params: options?.params,
      headers: options?.headers,
    });
  }
}
