import { CompositionClient, RequestOptions } from '../../../client.js';
import {
  PaginatedList,
  StepQueryResult,
  StepsQueryRequest,
  ZPaginatedList,
  ZStepQueryResult,
} from '../../../types.js';

/**
 * Workflow Steps API client for accessing step-level execution artifacts.
 *
 * Usage: `client.workflows.steps.get(stepId)`
 */
export default class APIWorkflowSteps extends CompositionClient {
  constructor(client: CompositionClient) {
    super(client);
  }

  /**
   * Get one joined step row by step ID.
   *
   * @example
   * ```typescript
   * const step = await client.workflows.steps.get("step_abc123");
   * const output = step.handle_outputs?.["output-json-0"];
   * ```
   */
  async get(stepId: string, options?: RequestOptions): Promise<StepQueryResult> {
    if (typeof stepId !== 'string' || stepId.length === 0) {
      throw new TypeError('stepId is required');
    }
    if (options !== undefined && (typeof options !== 'object' || Array.isArray(options))) {
      throw new TypeError('options must be an object');
    }
    return this._fetchJson(ZStepQueryResult, {
      url: `/workflows/steps/${encodeURIComponent(stepId)}`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Query joined step rows for a workflow.
   */
  async query(request: StepsQueryRequest, options?: RequestOptions): Promise<StepQueryResult[]> {
    return this._fetchJson(ZStepQueryResult.array(), {
      url: '/workflows/steps/query',
      method: 'POST',
      body: { ...request, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * List all persisted step documents for a workflow run.
   *
   * Returns the canonical
   * `{"data": [...], "list_metadata": {"before": null, "after": null}}`
   * pagination envelope. ID pagination is not yet implemented; iterate
   * `result.data` directly.
   *
   * @example
   * ```typescript
   * const steps = await client.workflows.steps.list("run_abc123");
   * for (const step of steps.data) {
   *     console.log(`${step.block_id}: ${step.lifecycle.status}`);
   * }
   * ```
   */
  async list(runId: string, options?: RequestOptions): Promise<PaginatedList> {
    return this._fetchJson(ZPaginatedList, {
      url: `/workflows/steps?run_id=${runId}`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }
}
