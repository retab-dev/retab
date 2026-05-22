import { CompositionClient, RequestOptions } from '../../../client.js';
import APIWorkflowTestRuns, { APIWorkflowTestRunResults } from './runs/client.js';
import {
  AssertionSpec,
  WorkflowTestListResponse,
  WorkflowTest,
  WorkflowTestBlockTarget,
  WorkflowTestSource,
  ZWorkflowTestListResponse,
  ZWorkflowTest,
} from './types.js';

/**
 * Workflow tests API client. Mirrors the eight backend endpoints
 * under `/v1/workflows/tests` plus the nested run
 * endpoints.
 *
 * Sub-clients:
 * - runs: read-only access to per-test execution history
 *
 * @example
 * ```typescript
 * const test = await client.workflows.tests.create({
 *     workflowId: "wf_abc123",
 *     target: { type: "block", block_id: "block_extract" },
 *     source: { type: "manual", handle_inputs: {} },
 *     assertion: {
 *         target: { output_handle_id: "output-json-0", path: "total" },
 *         condition: { kind: "equals", expected: 1234.56 },
 *     },
 *     name: "Q1 invoice total",
 * });
 *
 * // Run all tests for the workflow asynchronously.
 * const run = await client.workflows.tests.runs.create({
 *     workflowId: "wf_abc123",
 * });
 * console.log(run.id, run.lifecycle.status);
 * ```
 */
export default class APIWorkflowTests extends CompositionClient {
  public runs: APIWorkflowTestRuns;
  public results: APIWorkflowTestRunResults;

  constructor(client: CompositionClient) {
    super(client);
    this.runs = new APIWorkflowTestRuns(this);
    this.results = new APIWorkflowTestRunResults(this);
  }

  /**
   * Create a new workflow test against a single block in the workflow.
   *
   * `assertion` is required. Pass `target` and `source` as discriminated
   * union literals — see types.ts for the variants.
   */
  async create(
    {
      workflowId,
      target,
      source,
      assertion,
      name,
    }: {
      workflowId: string;
      target: WorkflowTestBlockTarget;
      // `WorkflowTestSource` IS `Manual | RunStep` already — no need
      // to repeat the union members. Callers benefit from the named
      // alias showing up in IDE hover.
      source: WorkflowTestSource;
      assertion: AssertionSpec;
      // Narrowed to `string | undefined` so the typed contract matches
      // behavior — the SDK omits `name` from the body whether the
      // caller passes `undefined` or `null`, but the TS signature
      // shouldn't lie about supporting `null`.
      name?: string;
    },
    options?: RequestOptions
  ): Promise<WorkflowTest> {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const body: Record<string, any> = {
      workflow_id: workflowId,
      target,
      source,
      assertion,
    };
    if (name !== undefined) {
      body.name = name;
    }
    return this._fetchJson(ZWorkflowTest, {
      url: '/workflows/tests',
      method: 'POST',
      body: { ...body, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Fetch a single workflow test by id (refreshes drift state).
   */
  async get(
    {
      testId,
    }: {
      testId: string;
    },
    options?: RequestOptions
  ): Promise<WorkflowTest> {
    return this._fetchJson(ZWorkflowTest, {
      url: `/workflows/tests/${testId}`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * List all workflow tests for a workflow.
   */
  async list(
    {
      workflowId,
      targetBlockId,
      limit = 50,
    }: {
      workflowId: string;
      targetBlockId?: string;
      limit?: number;
    },
    options?: RequestOptions
  ): Promise<WorkflowTestListResponse> {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const baseParams: Record<string, any> = { limit };
    if (targetBlockId !== undefined) {
      baseParams.target_block_id = targetBlockId;
    }
    // `options.params` wins so callers can override the typed args via
    // the escape hatch — matches sibling `APIWorkflowRuns.list`
    // precedence (runs/client.ts).
    return this._fetchJson(ZWorkflowTestListResponse, {
      url: `/workflows/tests?workflow_id=${workflowId}`,
      method: 'GET',
      params: { ...baseParams, ...(options?.params || {}) },
      headers: options?.headers,
    });
  }

  /**
   * Patch the name, assertion, and/or source of a workflow test. Each
   * field is optional — fields you omit are left untouched. Setting
   * `assertion: null` is rejected by the backend; use `delete()` to
   * remove the test entirely.
   */
  async update(
    {
      testId,
      name,
      assertion,
      source,
    }: {
      testId: string;
      name?: string;
      assertion?: AssertionSpec;
      source?: WorkflowTestSource;
    },
    options?: RequestOptions
  ): Promise<WorkflowTest> {
    // Only include fields the caller actually passed — the backend
    // treats missing fields as `leave untouched` and explicitly rejects
    // an `assertion: null` payload.
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const body: Record<string, any> = {};
    if (name !== undefined) body.name = name;
    if (assertion !== undefined) body.assertion = assertion;
    if (source !== undefined) body.source = source;

    return this._fetchJson(ZWorkflowTest, {
      url: `/workflows/tests/${testId}`,
      method: 'PATCH',
      body: { ...body, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Delete a workflow test. Returns void on success.
   */
  async delete(
    {
      testId,
    }: {
      testId: string;
    },
    options?: RequestOptions
  ): Promise<void> {
    return this._fetchJson({
      url: `/workflows/tests/${testId}`,
      method: 'DELETE',
      params: options?.params,
      headers: options?.headers,
    });
  }
}
