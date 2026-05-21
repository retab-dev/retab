import { CompositionClient, RequestOptions } from '../../../client.js';
import {
  BlockSimulation,
  BlockSimulationCreateRequest,
  PaginatedList,
  ZBlockSimulation,
  ZPaginatedList,
} from '../../../types.js';

function serializeCreateRequest(request: BlockSimulationCreateRequest): Record<string, unknown> {
  const body: Record<string, unknown> = {
    run_id: request.runId,
    block_id: request.blockId,
  };
  if (request.stepId !== undefined) body.step_id = request.stepId;
  if (request.nConsensus !== undefined) body.n_consensus = request.nConsensus;
  return body;
}

/**
 * Workflow Simulations API client for creating and listing block simulations.
 */
export default class APIWorkflowSimulations extends CompositionClient {
  constructor(client: CompositionClient) {
    super(client);
  }

  prepare_create(request: BlockSimulationCreateRequest): {
    url: string;
    method: string;
    body: Record<string, unknown>;
  } {
    return {
      url: '/workflows/simulations',
      method: 'POST',
      body: serializeCreateRequest(request),
    };
  }

  prepare_list({
    runId,
    blockId,
    limit = 20,
  }: {
    runId: string;
    blockId: string;
    limit?: number;
  }): {
    url: string;
    method: string;
    params: Record<string, unknown>;
  } {
    return {
      url: '/workflows/simulations',
      method: 'GET',
      params: {
        run_id: runId,
        block_id: blockId,
        limit,
      },
    };
  }

  /**
   * Create and run a workflow block simulation.
   */
  async create(
    request: BlockSimulationCreateRequest,
    options?: RequestOptions
  ): Promise<BlockSimulation> {
    const preparedRequest = this.prepare_create(request);
    return this._fetchJson(ZBlockSimulation, {
      url: preparedRequest.url,
      method: preparedRequest.method,
      body: {
        ...preparedRequest.body,
        ...((options?.body as Record<string, unknown>) || {}),
      },
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * List simulations for a run/block pair.
   */
  async list(
    {
      runId,
      blockId,
      limit = 20,
    }: {
      runId: string;
      blockId: string;
      limit?: number;
    },
    options?: RequestOptions
  ): Promise<PaginatedList> {
    const preparedRequest = this.prepare_list({ runId, blockId, limit });
    return this._fetchJson(ZPaginatedList, {
      url: preparedRequest.url,
      method: preparedRequest.method,
      params: {
        ...preparedRequest.params,
        ...(options?.params || {}),
      },
      headers: options?.headers,
    });
  }
}
