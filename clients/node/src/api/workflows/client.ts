import { CompositionClient, RequestOptions } from '../../client.js';
import {
  PaginatedList,
  Workflow,
  WorkflowDiagnosisResponse,
  ZPaginatedList,
  ZWorkflow,
  ZWorkflowDiagnosisResponse,
} from '../../types.js';
import APIWorkflowRuns from './runs/client.js';
import APIWorkflowReviews from './reviews/client.js';
import APIWorkflowBlocks from './blocks/client.js';
import APIWorkflowEdges from './edges/client.js';
import APIWorkflowArtifacts from './artifacts/client.js';
import APIWorkflowSpecs from './specs/client.js';
import APIWorkflowTests from './tests/client.js';
import APIWorkflowExperiments from './experiments/client.js';
import APIWorkflowSteps from './steps/client.js';
import APIWorkflowBlockExecutions from './block_executions/client.js';

/**
 * Workflows API client for workflow operations.
 *
 * Sub-clients:
 * - runs: Workflow run operations
 * - reviews: review operations (list, get, versions.create/get/list, approve, reject)
 * - blocks: Workflow block CRUD
 * - edges: Workflow edge CRUD
 * - artifacts: Workflow artifact dereference operations
 * - steps: Workflow run step operations (get, list)
 * - specs: Declarative workflow YAML validation, planning, apply, and export
 * - tests: Workflow tests CRUD + execution + run history
 * - experiments: Consensus-experiments CRUD + per-run + metrics
 * - block executions: Workflow block execution operations
 */
export default class APIWorkflows extends CompositionClient {
  public runs: APIWorkflowRuns;
  public reviews: APIWorkflowReviews;
  public blocks: APIWorkflowBlocks;
  public edges: APIWorkflowEdges;
  public artifacts: APIWorkflowArtifacts;
  public steps: APIWorkflowSteps;
  public specs: APIWorkflowSpecs;
  public tests: APIWorkflowTests;
  public experiments: APIWorkflowExperiments;
  public block_executions: APIWorkflowBlockExecutions;

  constructor(client: CompositionClient) {
    super(client);
    this.runs = new APIWorkflowRuns(this);
    this.reviews = new APIWorkflowReviews(this);
    this.blocks = new APIWorkflowBlocks(this);
    this.edges = new APIWorkflowEdges(this);
    this.artifacts = new APIWorkflowArtifacts(this);
    this.steps = new APIWorkflowSteps(this);
    this.specs = new APIWorkflowSpecs(this);
    this.tests = new APIWorkflowTests(this);
    this.experiments = new APIWorkflowExperiments(this);
    this.block_executions = new APIWorkflowBlockExecutions(this);
  }

  /**
   * Get a workflow by ID.
   */
  async get(workflowId: string, options?: RequestOptions): Promise<Workflow> {
    return this._fetchJson(ZWorkflow, {
      url: `/workflows/${workflowId}`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * List workflows with pagination.
   */
  async list(
    {
      before,
      after,
      limit = 10,
      order = 'desc',
      sortBy,
      fields,
    }: {
      before?: string;
      after?: string;
      limit?: number;
      order?: 'asc' | 'desc';
      sortBy?: string;
      fields?: string;
    } = {},
    options?: RequestOptions
  ): Promise<PaginatedList> {
    const params = Object.fromEntries(
      Object.entries({
        before,
        after,
        limit,
        order,
        sort_by: sortBy,
        fields,
        ...(options?.params || {}),
      }).filter(([_, value]) => value !== undefined)
    );

    return this._fetchJson(ZPaginatedList, {
      url: '/workflows',
      method: 'GET',
      params,
      headers: options?.headers,
    });
  }

  /**
   * Create a new workflow.
   *
   * @param name - Workflow name (default: "Untitled Workflow")
   * @param description - Workflow description (default: "")
   * @returns The created workflow (unpublished, with a default start_document block)
   *
   * @example
   * ```typescript
   * const wf = await client.workflows.create({ name: "Invoice Extractor" });
   * ```
   */
  async create(
    {
      name = 'Untitled Workflow',
      description = '',
    }: {
      name?: string;
      description?: string;
    } = {},
    options?: RequestOptions
  ): Promise<Workflow> {
    return this._fetchJson(ZWorkflow, {
      url: '/workflows',
      method: 'POST',
      body: { name, description, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Update a workflow's metadata.
   *
   * @param workflowId - The workflow ID
   * @param fields - Fields to update (only provided fields are changed)
   *
   * @example
   * ```typescript
   * const wf = await client.workflows.update("wf_abc123", { name: "New Name" });
   * ```
   */
  async update(
    workflowId: string,
    {
      name,
      description,
      emailTrigger,
    }: {
      name?: string;
      description?: string;
      emailTrigger?: {
        allowedSenders?: string[];
        allowedDomains?: string[];
      };
    },
    options?: RequestOptions
  ): Promise<Workflow> {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const body: Record<string, any> = {};
    if (name !== undefined) body.name = name;
    if (description !== undefined) body.description = description;
    if (emailTrigger !== undefined) {
      body.email_trigger = {};
      if (emailTrigger.allowedSenders !== undefined)
        body.email_trigger.allowed_senders = emailTrigger.allowedSenders;
      if (emailTrigger.allowedDomains !== undefined)
        body.email_trigger.allowed_domains = emailTrigger.allowedDomains;
    }

    return this._fetchJson(ZWorkflow, {
      url: `/workflows/${workflowId}`,
      method: 'PATCH',
      body: { ...body, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Delete a workflow and all its associated entities (blocks, edges, snapshots).
   */
  async delete(workflowId: string, options?: RequestOptions): Promise<void> {
    return this._fetchJson({
      url: `/workflows/${workflowId}`,
      method: 'DELETE',
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Publish a workflow's current draft as a new version.
   *
   * @example
   * ```typescript
   * const wf = await client.workflows.publish("wf_abc123", { description: "v1" });
   * console.log(wf.published?.version_id);
   * ```
   */
  async publish(
    workflowId: string,
    { description = '' }: { description?: string } = {},
    options?: RequestOptions
  ): Promise<Workflow> {
    return this._fetchJson(ZWorkflow, {
      url: `/workflows/${workflowId}/publish`,
      method: 'POST',
      body: { description, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  prepare_diagnose(
    workflowId: string,
    rePropagate = true
  ): {
    url: string;
    method: string;
    body: {
      re_propagate: boolean;
    };
  } {
    return {
      url: `/workflows/${workflowId}/diagnose-graph`,
      method: 'POST',
      body: {
        re_propagate: rePropagate,
      },
    };
  }

  /**
   * Diagnose the workflow's draft graph for structural/config issues.
   *
   * POSTs directly to `/v1/workflows/{workflow_id}/diagnose-graph`.
   * Returns a list of `issues` and graph `stats`.
   */
  async diagnose(
    workflowId: string,
    {
      rePropagate = true,
    }: {
      rePropagate?: boolean;
    } = {},
    options?: RequestOptions
  ): Promise<WorkflowDiagnosisResponse> {
    return this._fetchJson(ZWorkflowDiagnosisResponse, {
      url: `/workflows/${workflowId}/diagnose-graph`,
      method: 'POST',
      body: {
        re_propagate: rePropagate,
        ...((options?.body as Record<string, unknown>) || {}),
      },
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Lower-level: POST blocks + edges directly to the diagnose-graph
   * endpoint. Use this when you have an in-memory editor graph that
   * hasn't been saved yet.
   */
  async diagnoseGraph(
    workflowId: string,
    {
      blocks,
      edges,
      rePropagate = true,
    }: {
      blocks: Array<Record<string, unknown>>;
      edges: Array<Record<string, unknown>>;
      rePropagate?: boolean;
    },
    options?: RequestOptions
  ): Promise<WorkflowDiagnosisResponse> {
    return this._fetchJson(ZWorkflowDiagnosisResponse, {
      url: `/workflows/${workflowId}/diagnose-graph`,
      method: 'POST',
      body: {
        blocks,
        edges,
        re_propagate: rePropagate,
        ...((options?.body as Record<string, unknown>) || {}),
      },
      params: options?.params,
      headers: options?.headers,
    });
  }
}
