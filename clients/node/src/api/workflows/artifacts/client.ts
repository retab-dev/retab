import { CompositionClient, RequestOptions } from '../../../client.js';
import { PaginatedList, ZPaginatedList } from '../../../types.js';

/**
 * Workflow Artifacts API client for listing persisted run artifacts.
 */
export default class APIWorkflowArtifacts extends CompositionClient {
  constructor(client: CompositionClient) {
    super(client);
  }

  prepare_list(
    runId: string,
    operation?: string | null,
    blockId?: string | null
  ): {
    url: string;
    method: string;
    params: Record<string, string>;
  } {
    const params = Object.fromEntries(
      Object.entries({
        run_id: runId,
        operation,
        block_id: blockId,
      }).filter(([, value]) => value !== undefined && value !== null)
    ) as Record<string, string>;

    return {
      url: '/workflows/artifacts',
      method: 'GET',
      params,
    };
  }

  /**
   * List dereferenced artifacts produced by one workflow run.
   *
   * Returns the canonical
   * `{"data": [...], "list_metadata": {"before": null, "after": null}}`
   * pagination envelope. ID pagination is not yet implemented for this
   * endpoint; `list_metadata` is always `{before: null, after: null}`.
   */
  async list(
    {
      runId,
      operation,
      blockId,
    }: {
      runId: string;
      operation?: string;
      blockId?: string;
    },
    options?: RequestOptions
  ): Promise<PaginatedList> {
    const request = this.prepare_list(runId, operation, blockId);
    const params = {
      ...request.params,
      ...(options?.params || {}),
    };

    return this._fetchJson(ZPaginatedList, {
      url: request.url,
      method: request.method,
      params,
      headers: options?.headers,
    });
  }
}
