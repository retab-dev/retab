import * as z from 'zod';
import { CompositionClient, RequestOptions } from '../../../client.js';
import {
  BlockResolvedSchemasResponse,
  BlockSimulation,
  PaginatedList,
  WorkflowBlock,
  WorkflowBlockCreateRequest,
  WorkflowBlockUpdateRequest,
  ZBlockResolvedSchemasResponse,
  ZBlockSimulation,
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

  prepare_config_history(workflowId: string, blockId: string): { url: string; method: string } {
    return {
      url: `/workflows/blocks/${blockId}/config-history?workflow_id=${workflowId}`,
      method: 'GET',
    };
  }

  /**
   * Return the config-version timeline for a block, wrapped in the canonical
   * pagination envelope. Each `data` entry groups consecutive workflow
   * snapshots in which the block's config did not change. Cursor pagination
   * is not yet implemented for this endpoint.
   */
  async configHistory(
    workflowId: string,
    blockId: string,
    options?: RequestOptions
  ): Promise<PaginatedList> {
    return this._fetchJson(ZPaginatedList, {
      url: `/workflows/blocks/${blockId}/config-history?workflow_id=${workflowId}`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }

  async config_history(
    workflowId: string,
    blockId: string,
    options?: RequestOptions
  ): Promise<PaginatedList> {
    return this.configHistory(workflowId, blockId, options);
  }

  prepare_list_simulations(
    runId: string,
    blockId: string,
    limit?: number | null
  ): {
    url: string;
    method: string;
    params?: Record<string, unknown>;
  } {
    const params: Record<string, unknown> = {};
    if (limit !== undefined && limit !== null) params.limit = limit;
    return {
      url: `/workflows/steps/${blockId}/simulations?run_id=${runId}`,
      method: 'GET',
      ...(Object.keys(params).length > 0 ? { params } : {}),
    };
  }

  /**
   * List recent simulation results for a block in a workflow run.
   *
   * Returns the canonical pagination envelope. Cursor pagination is not yet
   * implemented; pass `limit` (default 20, max 100) to bound the page size.
   */
  async listSimulations(
    runId: string,
    blockId: string,
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
      url: `/workflows/steps/${blockId}/simulations?run_id=${runId}`,
      method: 'GET',
      params,
      headers: options?.headers,
    });
  }

  async list_simulations(
    runId: string,
    blockId: string,
    opts: { limit?: number } = {},
    options?: RequestOptions
  ): Promise<PaginatedList> {
    return this.listSimulations(runId, blockId, opts, options);
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

  prepare_get_resolved_schemas(
    workflowId: string,
    blockId: string
  ): { url: string; method: string } {
    return {
      url: `/workflows/blocks/${blockId}/resolved-schemas?workflow_id=${workflowId}`,
      method: 'GET',
    };
  }

  /**
   * Get graph-derived schemas for one current-draft block.
   */
  async getResolvedSchemas(
    workflowId: string,
    blockId: string,
    options?: RequestOptions
  ): Promise<BlockResolvedSchemasResponse> {
    return this._fetchJson(ZBlockResolvedSchemasResponse, {
      url: `/workflows/blocks/${blockId}/resolved-schemas?workflow_id=${workflowId}`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }

  async get_resolved_schemas(
    workflowId: string,
    blockId: string,
    options?: RequestOptions
  ): Promise<BlockResolvedSchemasResponse> {
    return this.getResolvedSchemas(workflowId, blockId, options);
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
   * Create multiple blocks in a single request.
   *
   * @param workflowId - The workflow ID
   * @param blocks - Array of block definitions (each with at least `id` and `type`)
   */
  async createBatch(
    workflowId: string,
    blocks: WorkflowBlockCreateRequest[],
    options?: RequestOptions
  ): Promise<WorkflowBlock[]> {
    return this._fetchJson(z.array(ZWorkflowBlock), {
      url: `/workflows/blocks/batch?workflow_id=${workflowId}`,
      method: 'POST',
      body: blocks.map((block) => serializeBlockCreateRequest(block)),
      params: options?.params,
      headers: options?.headers,
    });
  }

  async create_batch(
    workflowId: string,
    blocks: WorkflowBlockCreateRequest[],
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

  prepare_simulate(
    runId: string,
    blockId: string,
    nConsensus?: number | null,
    stepId?: string | null,
    checkEligibility = true
  ): {
    url: string;
    method: string;
    data?: Record<string, unknown>;
  } {
    const data: Record<string, unknown> = {
      run_id: runId,
      step_id: stepId !== undefined && stepId !== null ? stepId : blockId,
    };
    if (stepId !== undefined && stepId !== null && stepId !== blockId) {
      data.source_step_id = stepId;
    }
    if (nConsensus !== undefined && nConsensus !== null) data.n_consensus = nConsensus;
    if (!checkEligibility) data.check_eligibility = false;

    return {
      url: '/workflows/simulations',
      method: 'POST',
      data,
    };
  }

  /**
   * Replay one block using inputs from a previous run + the current
   * draft config.
   *
   * Backend route: `POST /v1/workflows/simulations` — simulations are a
   * top-level resource, with `step_id` and `run_id` carried in the body.
   *
   * @example
   * ```typescript
   * const sim = await client.workflows.blocks.simulate({
   *     runId: "run_abc123",
   *     blockId: "extract-1",
   *     nConsensus: 5,
   * });
   * console.log(sim.success, sim.handle_outputs);
   * ```
   */
  async simulate(
    {
      runId,
      blockId,
      nConsensus,
      stepId,
      checkEligibility = true,
    }: {
      runId: string;
      blockId: string;
      nConsensus?: 3 | 5 | 7;
      stepId?: string;
      checkEligibility?: boolean;
    },
    options?: RequestOptions
  ): Promise<BlockSimulation> {
    const request = this.prepare_simulate(runId, blockId, nConsensus, stepId, checkEligibility);

    return this._fetchJson(ZBlockSimulation, {
      url: request.url,
      method: request.method,
      body: { ...request.data, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }
}
