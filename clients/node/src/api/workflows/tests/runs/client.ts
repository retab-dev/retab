import { CompositionClient, RequestOptions } from '../../../../client.js';
import {
  WorkflowTestBlockTarget,
  WorkflowTestResult,
  WorkflowTestResultListResponse,
  WorkflowTestRunListResponse,
  WorkflowTestRun,
  ZWorkflowTestResult,
  ZWorkflowTestResultListResponse,
  ZWorkflowTestRunListResponse,
  ZWorkflowTestRun,
} from '../types.js';

/**
 * Access to workflow-test run lifecycle and child results.
 */
export default class APIWorkflowTestRuns extends CompositionClient {
  public results: APIWorkflowTestRunResults;

  constructor(client: CompositionClient) {
    super(client);
    this.results = new APIWorkflowTestRunResults(this);
  }

  async create(
    {
      workflowId,
      testId,
      target,
      nConsensus,
    }: {
      workflowId: string;
      testId?: string;
      target?: WorkflowTestBlockTarget;
      nConsensus?: number;
    },
    options?: RequestOptions
  ): Promise<WorkflowTestRun> {
    const body: Record<string, unknown> = {};
    if (testId !== undefined) body.test_id = testId;
    if (target !== undefined) body.target = target;
    if (nConsensus !== undefined) body.n_consensus = nConsensus;

    return this._fetchJson(ZWorkflowTestRun, {
      url: `/workflows/${workflowId}/tests/runs`,
      method: 'POST',
      body: { ...body, ...((options?.body as Record<string, unknown>) || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * List workflow-test runs, newest first.
   *
   * @example
   * ```typescript
   * const runs = await client.workflows.tests.runs.list({
   *     workflowId: "wf_abc123",
   *     testId: "wfnodetest_abc",
   *     limit: 10,
   * });
   * ```
   */
  async list(
    {
      workflowId,
      testId,
      targetBlockId,
      status,
      statuses,
      excludeStatus,
      triggerType,
      triggerTypes,
      fromDate,
      toDate,
      sortBy,
      fields,
      before,
      after,
      limit = 20,
      order,
    }: {
      workflowId?: string;
      testId?: string;
      targetBlockId?: string;
      status?: string;
      statuses?: string[] | string;
      excludeStatus?: string;
      triggerType?: string;
      triggerTypes?: string[] | string;
      fromDate?: string;
      toDate?: string;
      sortBy?: string;
      fields?: string[] | string;
      before?: string;
      after?: string;
      limit?: number;
      order?: string;
    },
    options?: RequestOptions
  ): Promise<WorkflowTestRunListResponse> {
    const params = Object.fromEntries(
      Object.entries({
        workflow_id: workflowId,
        test_id: testId,
        target_block_id: targetBlockId,
        status,
        statuses: Array.isArray(statuses) ? statuses.join(',') : statuses,
        exclude_status: excludeStatus,
        trigger_type: triggerType,
        trigger_types: Array.isArray(triggerTypes) ? triggerTypes.join(',') : triggerTypes,
        from_date: fromDate,
        to_date: toDate,
        sort_by: sortBy,
        fields: Array.isArray(fields) ? fields.join(',') : fields,
        before,
        after,
        limit,
        order,
      }).filter(([, value]) => value !== undefined)
    );
    return this._fetchJson(ZWorkflowTestRunListResponse, {
      url: '/workflows/tests/runs',
      method: 'GET',
      params: { ...params, ...(options?.params || {}) },
      headers: options?.headers,
    });
  }

  /**
   * Fetch a parent workflow-test run by run id.
   */
  async get(
    {
      runId,
    }: {
      runId: string;
    },
    options?: RequestOptions
  ): Promise<WorkflowTestRun> {
    return this._fetchJson(ZWorkflowTestRun, {
      url: `/workflows/tests/runs/${runId}`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }

  async cancel({ runId }: { runId: string }, options?: RequestOptions): Promise<WorkflowTestRun> {
    return this._fetchJson(ZWorkflowTestRun, {
      url: `/workflows/tests/runs/${runId}/cancel`,
      method: 'POST',
      body: { ...((options?.body as Record<string, unknown>) || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }
}

export class APIWorkflowTestRunResults extends CompositionClient {
  constructor(client: CompositionClient) {
    super(client);
  }

  async list(
    { runId, limit = 20 }: { runId: string; limit?: number },
    options?: RequestOptions
  ): Promise<WorkflowTestResultListResponse> {
    return this._fetchJson(ZWorkflowTestResultListResponse, {
      url: `/workflows/tests/runs/${runId}/results`,
      method: 'GET',
      params: { limit, ...(options?.params || {}) },
      headers: options?.headers,
    });
  }

  async get(
    { runId, testId }: { runId: string; testId: string },
    options?: RequestOptions
  ): Promise<WorkflowTestResult> {
    return this._fetchJson(ZWorkflowTestResult, {
      url: `/workflows/tests/runs/${runId}/results/${testId}`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }
}
