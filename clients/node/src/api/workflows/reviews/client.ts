import { CompositionClient, RequestOptions } from '../../../client.js';
import {
  Review,
  WorkflowReviewQueue,
  ReviewVersion,
  ReviewVersionListResponse,
  SubmitDecisionResponse,
  ZReview,
  ZWorkflowReviewQueue,
  ZReviewVersion,
  ZReviewVersionListResponse,
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
 * Flat CRUD client for review versions.
 *
 * Versions are immutable resources served under `/workflows/reviews/versions`.
 */
export class APIWorkflowReviewVersions extends CompositionClient {
  constructor(client: CompositionClient) {
    super(client);
  }

  prepare_create({
    reviewId,
    parentVersionId,
    snapshot,
    note,
  }: {
    reviewId: string;
    parentVersionId: string;
    snapshot: Record<string, unknown>;
    note?: string | null;
  }): PreparedReviewRequest {
    if (typeof parentVersionId !== 'string' || parentVersionId.length === 0) {
      throw new Error('parentVersionId is required when creating a review version');
    }
    const body: Record<string, unknown> = {
      review_id: reviewId,
      parent_id: parentVersionId,
      snapshot,
    };
    if (note !== undefined && note !== null) {
      body.note = note;
    }
    return {
      url: '/workflows/reviews/versions',
      method: 'POST',
      body,
    };
  }

  prepare_get(versionId: string): PreparedReviewRequest {
    return { url: `/workflows/reviews/versions/${versionId}`, method: 'GET' };
  }

  prepare_list({
    reviewId,
    limit = 50,
    before,
    after,
  }: {
    reviewId: string;
    limit?: number;
    before?: string;
    after?: string;
  }): PreparedReviewRequest {
    const params: Record<string, unknown> = {
      review_id: reviewId,
      limit,
    };
    if (before !== undefined) params.before = before;
    if (after !== undefined) params.after = after;
    return { url: '/workflows/reviews/versions', method: 'GET', params };
  }

  /** Create an immutable snapshot version for a review. */
  async create(
    request: {
      reviewId: string;
      parentVersionId: string;
      snapshot: Record<string, unknown>;
      note?: string | null;
    },
    options?: RequestOptions
  ): Promise<ReviewVersion> {
    const prepared = this.prepare_create(request);
    return this._fetchJson(ZReviewVersion, {
      url: prepared.url,
      method: prepared.method,
      body: { ...prepared.body, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  /** Fetch one immutable review version. */
  async get(versionId: string, options?: RequestOptions): Promise<ReviewVersion> {
    const request = this.prepare_get(versionId);
    return this._fetchJson(ZReviewVersion, {
      url: request.url,
      method: request.method,
      params: options?.params,
      headers: options?.headers,
    });
  }

  /** List versions for one review. `reviewId` is required. */
  async list(
    request: {
      reviewId: string;
      limit?: number;
      before?: string;
      after?: string;
    },
    options?: RequestOptions
  ): Promise<ReviewVersionListResponse> {
    const prepared = this.prepare_list(request);
    return this._fetchJson(ZReviewVersionListResponse, {
      url: prepared.url,
      method: prepared.method,
      params: { ...prepared.params, ...(options?.params || {}) },
      headers: options?.headers,
    });
  }
}

/**
 * Actor-neutral client for workflow reviews.
 *
 * Drives the review loop served under `/workflows/reviews`: list reviews,
 * fetch a review by id, manage immutable versions, and approve or reject one
 * exact version through action endpoints.
 */
export default class APIWorkflowReviews extends CompositionClient {
  public versions: APIWorkflowReviewVersions;

  constructor(client: CompositionClient) {
    super(client);
    this.versions = new APIWorkflowReviewVersions(this);
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
  ): Promise<WorkflowReviewQueue> {
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

    return this._fetchJson(ZWorkflowReviewQueue, {
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
  async get(reviewId: string, options?: RequestOptions): Promise<Review> {
    const request = this.prepare_get(reviewId);
    return this._fetchJson(ZReview, {
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
}
