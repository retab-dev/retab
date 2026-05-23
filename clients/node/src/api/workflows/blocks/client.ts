import { CompositionClient, RequestOptions } from '../../../client.js';
import { PaginatedList } from '../../_pagination.js';
import {
  WorkflowBlock,
  WorkflowBlockCreateRequest,
  UpdateWorkflowBlockRequest,
  ZWorkflowBlock,
} from '../../../types.js';
import APIWorkflowBlockExecutions from './executions/client.js';

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

function serializeBlockUpdateRequest(request: UpdateWorkflowBlockRequest): Record<string, unknown> {
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
 *
 * Sub-clients:
 * - executions: Per-block execution operations
 *   (`client.workflows.blocks.executions.create / .list`)
 */
export default class APIWorkflowBlocks extends CompositionClient {
  public executions: APIWorkflowBlockExecutions;

  constructor(client: CompositionClient) {
    super(client);
    this.executions = new APIWorkflowBlockExecutions(this);
  }

  /**
   * List all blocks for a workflow.
   *
   * Returns the canonical
   * `{"data": [...], "list_metadata": {"before": null, "after": null}}`
   * pagination envelope. ID pagination is not yet implemented for this
   * endpoint; `list_metadata` is always `{before: null, after: null}`.
   */
  async list(workflowId: string, options?: RequestOptions): Promise<PaginatedList<WorkflowBlock>> {
    return this._fetchPage(ZWorkflowBlock, {
      url: `/workflows/blocks?workflow_id=${workflowId}`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Get a single block by ID.
   */
  async get(blockId: string, options?: RequestOptions): Promise<WorkflowBlock> {
    return this._fetchJson(ZWorkflowBlock, {
      url: `/workflows/blocks/${blockId}`,
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
      url: '/workflows/blocks',
      method: 'POST',
      body: {
        workflow_id: workflowId,
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
    blockId: string,
    request: UpdateWorkflowBlockRequest,
    options?: RequestOptions
  ): Promise<WorkflowBlock> {
    return this._fetchJson(ZWorkflowBlock, {
      url: `/workflows/blocks/${blockId}`,
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
  async delete(blockId: string, options?: RequestOptions): Promise<void> {
    return this._fetchJson({
      url: `/workflows/blocks/${blockId}`,
      method: 'DELETE',
      params: options?.params,
      headers: options?.headers,
    });
  }
}
