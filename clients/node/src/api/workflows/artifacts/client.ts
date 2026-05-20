import { CompositionClient, RequestOptions } from '../../../client.js';
import {
  PaginatedList,
  StepArtifactRef,
  WorkflowArtifact,
  ZPaginatedList,
  ZWorkflowArtifact,
} from '../../../types.js';

/**
 * Workflow Artifacts API client for dereferencing step artifact refs.
 *
 * Usage: `client.workflows.artifacts.get(step.artifact)`
 */
export default class APIWorkflowArtifacts extends CompositionClient {
  constructor(client: CompositionClient) {
    super(client);
  }

  prepare_get(
    operation: StepArtifactRef,
    artifactId?: string | null
  ): {
    url: string;
    method: string;
  };
  prepare_get(
    operation: string,
    artifactId: string
  ): {
    url: string;
    method: string;
  };
  prepare_get(
    operationOrArtifact: string | StepArtifactRef,
    artifactId?: string | null
  ): {
    url: string;
    method: string;
  } {
    const operation =
      typeof operationOrArtifact === 'string' ? operationOrArtifact : operationOrArtifact.operation;
    const id = typeof operationOrArtifact === 'string' ? artifactId : operationOrArtifact.id;

    if (typeof id !== 'string' || id.length === 0) {
      throw new TypeError('artifact_id is required');
    }
    return {
      url: `/workflows/artifacts/${operation}/${id}`,
      method: 'GET',
    };
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
   * Dereference a workflow step artifact ref into its persisted record.
   *
   * @example
   * ```typescript
   * const step = await client.workflows.runs.steps.get(run.id, "review-1");
   * if (step.artifact) {
   *   const artifact = await client.workflows.artifacts.get(step.artifact);
   *   console.log(artifact.operation, artifact.evaluations);
   * }
   * ```
   */
  async get(artifact: StepArtifactRef, options?: RequestOptions): Promise<WorkflowArtifact>;
  async get(operation: string, id: string, options?: RequestOptions): Promise<WorkflowArtifact>;
  async get(
    operationOrArtifact: string | StepArtifactRef,
    idOrOptions?: string | RequestOptions,
    maybeOptions?: RequestOptions
  ): Promise<WorkflowArtifact> {
    const operation =
      typeof operationOrArtifact === 'string' ? operationOrArtifact : operationOrArtifact.operation;
    const id = typeof operationOrArtifact === 'string' ? idOrOptions : operationOrArtifact.id;
    const options =
      typeof operationOrArtifact === 'string'
        ? maybeOptions
        : (idOrOptions as RequestOptions | undefined);

    if (typeof id !== 'string' || id.length === 0) {
      throw new TypeError('artifact id is required');
    }
    const request = this.prepare_get(operation, id);
    return this._fetchJson(ZWorkflowArtifact, {
      url: request.url,
      method: request.method,
      params: options?.params,
      headers: options?.headers,
    });
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
