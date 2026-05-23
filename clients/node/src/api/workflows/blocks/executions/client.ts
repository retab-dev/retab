import { CompositionClient, RequestOptions } from '../../../../client.js';
import { PaginatedList } from '../../../_pagination.js';
import {
  StoredBlockExecution,
  BlockExecutionCreateRequest,
  ZStoredBlockExecution,
} from '../../../../types.js';

function serializeCreateRequest(request: BlockExecutionCreateRequest): Record<string, unknown> {
  const body: Record<string, unknown> = {
    run_id: request.runId,
    block_id: request.blockId,
  };
  if (request.stepId !== undefined) body.step_id = request.stepId;
  if (request.nConsensus !== undefined) body.n_consensus = request.nConsensus;
  return body;
}

/**
 * Workflow Blocks Executions API client for creating and listing block executions.
 */
export default class APIWorkflowBlockExecutions extends CompositionClient {
  constructor(client: CompositionClient) {
    super(client);
  }

  prepare_create(request: BlockExecutionCreateRequest): {
    url: string;
    method: string;
    body: Record<string, unknown>;
  } {
    return {
      url: '/workflows/blocks/executions',
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
      url: '/workflows/blocks/executions',
      method: 'GET',
      params: {
        run_id: runId,
        block_id: blockId,
        limit,
      },
    };
  }

  /**
   * Create and run a workflow block execution.
   */
  async create(
    request: BlockExecutionCreateRequest,
    options?: RequestOptions
  ): Promise<StoredBlockExecution> {
    const preparedRequest = this.prepare_create(request);
    return this._fetchJson(ZStoredBlockExecution, {
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
   * List block executions for a run/block pair.
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
  ): Promise<PaginatedList<StoredBlockExecution>> {
    const preparedRequest = this.prepare_list({ runId, blockId, limit });
    return this._fetchPage(ZStoredBlockExecution, {
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
