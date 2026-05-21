import { APIError, CompositionClient, RequestOptions } from '../../../client.js';
import {
  AppendVersionResponse,
  ReviewOverlay,
  ReviewQueueResponse,
  SubmitDecisionResponse,
  ZAppendVersionResponse,
  ZReviewOverlay,
  ZReviewQueueResponse,
  ZSubmitDecisionResponse,
} from '../../../types.js';

/**
 * Decision-state filter for {@link APIWorkflowReviews.list}.
 *
 * - `'pending'` (default): only open reviews.
 * - `'approved'`: approved reviews only.
 * - `'rejected'`: rejected reviews only.
 * - `'decided'`: approved or rejected reviews.
 * - `'all'`: no decision-state filter.
 */
export type ReviewDecisionStatus = 'pending' | 'approved' | 'rejected' | 'decided' | 'all';

type PreparedReviewRequest = {
  url: string;
  method: 'GET' | 'POST';
  params?: Record<string, unknown>;
  body?: Record<string, unknown>;
};

/**
 * Actor-neutral client for workflow reviews.
 *
 * Drives the review loop served under `/workflows/reviews`: list summaries,
 * fetch a review by id, append output versions, and approve or reject one
 * exact version.
 */
export default class APIWorkflowReviews extends CompositionClient {
  constructor(client: CompositionClient) {
    super(client);
  }

  prepare_list({
    workflowId,
    runId,
    blockId,
    stepId,
    iterationKey,
    limit = 50,
    decisionStatus = 'pending',
    before,
    after,
  }: {
    workflowId?: string;
    runId?: string;
    blockId?: string;
    stepId?: string;
    iterationKey?: string;
    limit?: number;
    decisionStatus?: ReviewDecisionStatus;
    before?: string;
    after?: string;
  } = {}): PreparedReviewRequest {
    const params: Record<string, unknown> = { limit, decision_status: decisionStatus };
    if (workflowId !== undefined) params.workflow_id = workflowId;
    if (runId !== undefined) params.run_id = runId;
    if (blockId !== undefined) params.block_id = blockId;
    if (stepId !== undefined) params.step_id = stepId;
    if (iterationKey !== undefined) params.iteration_key = iterationKey;
    if (before !== undefined) params.before = before;
    if (after !== undefined) params.after = after;
    return { url: '/workflows/reviews', method: 'GET', params };
  }

  prepare_get(reviewId: string): PreparedReviewRequest {
    return { url: `/workflows/reviews/${reviewId}`, method: 'GET' };
  }

  prepare_append_version(
    reviewId: string,
    {
      snapshot,
      parentVersionId,
      note,
    }: {
      snapshot: Record<string, unknown>;
      parentVersionId: string;
      note?: string | null;
    }
  ): PreparedReviewRequest {
    const body: Record<string, unknown> = {
      snapshot,
      parent_version_id: parentVersionId,
    };
    if (note !== undefined && note !== null) {
      body.note = note;
    }
    return {
      url: `/workflows/reviews/${reviewId}/versions`,
      method: 'POST',
      body,
    };
  }

  prepare_approve(reviewId: string, versionId: string): PreparedReviewRequest {
    return {
      url: `/workflows/reviews/${reviewId}/approve`,
      method: 'POST',
      body: { version_id: versionId },
    };
  }

  prepare_reject(reviewId: string, versionId: string, reason: string): PreparedReviewRequest {
    return {
      url: `/workflows/reviews/${reviewId}/reject`,
      method: 'POST',
      body: { version_id: versionId, reason },
    };
  }

  /**
   * List review summaries.
   *
   * @param workflowId - Restrict the queue to a single workflow.
   * @param runId - Restrict the queue to a single workflow run.
   * @param blockId - Restrict the queue to a workflow block.
   * @param stepId - Restrict the queue to one execution step.
   * @param iterationKey - Restrict the queue to one for_each iteration key.
   * @param limit - Page size (1-200). Defaults to 50.
   * @param decisionStatus - Defaults to `'pending'`.
   * @param before - Cursor from `list_metadata.before`.
   * @param after - Cursor from `list_metadata.after`.
   */
  async list(
    {
      workflowId,
      runId,
      blockId,
      stepId,
      iterationKey,
      limit = 50,
      decisionStatus = 'pending',
      before,
      after,
    }: {
      workflowId?: string;
      runId?: string;
      blockId?: string;
      stepId?: string;
      iterationKey?: string;
      limit?: number;
      decisionStatus?: ReviewDecisionStatus;
      before?: string;
      after?: string;
    } = {},
    options?: RequestOptions
  ): Promise<ReviewQueueResponse> {
    const request = this.prepare_list({
      workflowId,
      runId,
      blockId,
      stepId,
      iterationKey,
      limit,
      decisionStatus,
      before,
      after,
    });

    return this._fetchJson(ZReviewQueueResponse, {
      url: request.url,
      method: request.method,
      params: { ...request.params, ...(options?.params || {}) },
      headers: options?.headers,
    });
  }

  /**
   * Fetch the full review by id.
   *
   * @throws `APIError` with `.status === 404` when no review exists.
   */
  async get(reviewId: string, options?: RequestOptions): Promise<ReviewOverlay> {
    const request = this.prepare_get(reviewId);
    return this._fetchJson(ZReviewOverlay, {
      url: request.url,
      method: request.method,
      params: options?.params,
      headers: options?.headers,
    });
  }

  /** Approve one exact output version. */
  async approve(
    reviewId: string,
    {
      versionId,
    }: {
      versionId: string;
    },
    options?: RequestOptions
  ): Promise<SubmitDecisionResponse> {
    const request = this.prepare_approve(reviewId, versionId);
    return this._fetchJson(ZSubmitDecisionResponse, {
      url: request.url,
      method: request.method,
      body: { ...request.body, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  /** Reject one exact output version. */
  async reject(
    reviewId: string,
    {
      versionId,
      reason,
    }: {
      versionId: string;
      reason: string;
    },
    options?: RequestOptions
  ): Promise<SubmitDecisionResponse> {
    const request = this.prepare_reject(reviewId, versionId, reason);
    return this._fetchJson(ZSubmitDecisionResponse, {
      url: request.url,
      method: request.method,
      body: { ...request.body, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  /** Append a new output version to the review's version graph. */
  async appendVersion(
    reviewId: string,
    {
      snapshot,
      parentVersionId,
      note,
    }: {
      snapshot: Record<string, unknown>;
      parentVersionId: string;
      note?: string | null;
    },
    options?: RequestOptions
  ): Promise<AppendVersionResponse> {
    const request = this.prepare_append_version(reviewId, {
      snapshot,
      parentVersionId,
      note,
    });

    return this._fetchJson(ZAppendVersionResponse, {
      url: request.url,
      method: request.method,
      body: { ...request.body, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  async append_version(
    reviewId: string,
    request: {
      snapshot: Record<string, unknown>;
      parentVersionId: string;
      note?: string | null;
    },
    options?: RequestOptions
  ): Promise<AppendVersionResponse> {
    return this.appendVersion(reviewId, request, options);
  }

  /**
   * Poll until a review exists and has no terminal decision.
   */
  async waitFor(
    reviewId: string,
    {
      timeoutMs = 120000,
      pollIntervalMs = 2000,
    }: {
      timeoutMs?: number;
      pollIntervalMs?: number;
    } = {},
    options?: RequestOptions
  ): Promise<ReviewOverlay> {
    const deadline = Date.now() + timeoutMs;
    for (;;) {
      try {
        const review = await this.get(reviewId, options);
        if (review.decision === null) {
          return review;
        }
      } catch (error) {
        if (!(error instanceof APIError && error.status === 404)) {
          throw error;
        }
      }
      if (Date.now() >= deadline) {
        throw new Error(`Review '${reviewId}' was not pending within ${timeoutMs}ms`);
      }
      await new Promise((resolve) => setTimeout(resolve, pollIntervalMs));
    }
  }

  async wait_for(
    reviewId: string,
    options?: {
      timeoutMs?: number;
      pollIntervalMs?: number;
    },
    requestOptions?: RequestOptions
  ): Promise<ReviewOverlay> {
    return this.waitFor(reviewId, options, requestOptions);
  }
}
