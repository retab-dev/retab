import { APIError, CompositionClient, RequestOptions } from '../../../client.js';
import {
  ReviewOverlay,
  ReviewQueueResponse,
  SubmitDecisionResponse,
  ZReviewOverlay,
  ZReviewQueueResponse,
  ZSubmitDecisionResponse,
} from '../../../types.js';

/**
 * Decision-state filter for {@link APIWorkflowReviews.list}.
 *
 * - `'none'` (default): only open reviews (the work queue).
 * - `'any'`: include reviews that already have a terminal decision.
 */
export type ReviewDecisionFilter = 'none' | 'any';

type PreparedReviewRequest = {
  url: string;
  method: 'GET' | 'POST';
  params?: Record<string, unknown>;
  body?: Record<string, unknown>;
};

function withoutOrigin(body: Record<string, unknown> | undefined): Record<string, unknown> {
  const { origin: _origin, ...cleanBody } = body || {};
  return cleanBody;
}

/**
 * Actor-neutral client for workflow reviews.
 *
 * Drives the review loop served under `/workflows/reviews`: list the queue,
 * fetch a review, post new output versions, and submit verdicts.
 *
 * Actor symmetry is a hard rule here. A proposal authored by a model, an
 * agent, or a human goes through the same create-version / approve pair.
 * `Actor.kind` is data carried on the review — no method, parameter, or branch
 * in this client depends on it.
 *
 * Usage: `client.workflows.reviews.list()`
 */
export default class APIWorkflowReviews extends CompositionClient {
  constructor(client: CompositionClient) {
    super(client);
  }

  prepare_list(
    workflowId?: string,
    limit = 50,
    decision: ReviewDecisionFilter = 'none',
    before?: string,
    after?: string
  ): PreparedReviewRequest {
    const params: Record<string, unknown> = { limit, decision };
    if (workflowId !== undefined) params.workflow_id = workflowId;
    if (before !== undefined) params.before = before;
    if (after !== undefined) params.after = after;
    return { url: '/workflows/reviews', method: 'GET', params };
  }

  prepare_get(runId: string, blockId: string): PreparedReviewRequest {
    return { url: `/workflows/reviews/${runId}/${blockId}`, method: 'GET' };
  }

  prepare_create_version(
    runId: string,
    blockId: string,
    {
      snapshot,
      parentId,
      note,
    }: {
      snapshot: Record<string, unknown>;
      parentId: string;
      note?: string | null;
    }
  ): PreparedReviewRequest {
    const body: Record<string, unknown> = {
      snapshot,
      parent_id: parentId,
    };
    if (note !== undefined && note !== null) {
      body.note = note;
    }
    return {
      url: `/workflows/reviews/${runId}/${blockId}/versions`,
      method: 'POST',
      body,
    };
  }

  prepare_decision(
    runId: string,
    blockId: string,
    {
      verdict,
      versionId,
      reason,
    }: {
      verdict: 'approved' | 'rejected';
      versionId: string;
      reason?: string;
    }
  ): PreparedReviewRequest {
    const body: Record<string, unknown> = {
      verdict,
      version_id: versionId,
    };
    if (reason !== undefined) {
      body.reason = reason;
    }
    return {
      url: `/workflows/reviews/${runId}/${blockId}/decision`,
      method: 'POST',
      body,
    };
  }

  /**
   * List review-queue items.
   *
   * @param workflowId - Restrict the queue to a single workflow.
   * @param limit - Page size (1-200). Defaults to 50.
   * @param decision - `'none'` (default) returns only open reviews; `'any'` includes decided ones.
   * @param before - Cursor — only return reviews that appear before this review id
   *   in the result order. Use `list_metadata.before` from the previous page.
   *   Mutually exclusive with `after`.
   * @param after - Cursor — only return reviews that appear after this review id
   *   in the result order. Use `list_metadata.after` from the previous page.
   *   Mutually exclusive with `before`.
   * @returns A page of lightweight `ReviewQueueItem` summaries plus a `list_metadata`
   *   cursor (`{ before, after }`). `list_metadata.after !== null` means another page exists.
   *
   * @example
   * ```typescript
   * const queue = await client.workflows.reviews.list({ workflowId: 'wf_1' });
   * const audit = await client.workflows.reviews.list({ decision: 'any' });
   * const next = await client.workflows.reviews.list({ after: queue.list_metadata.after! });
   * ```
   */
  async list(
    {
      workflowId,
      limit = 50,
      decision = 'none',
      before,
      after,
    }: {
      workflowId?: string;
      limit?: number;
      decision?: ReviewDecisionFilter;
      before?: string;
      after?: string;
    } = {},
    options?: RequestOptions
  ): Promise<ReviewQueueResponse> {
    const request = this.prepare_list(workflowId, limit, decision, before, after);

    return this._fetchJson(ZReviewQueueResponse, {
      url: request.url,
      method: request.method,
      params: { ...request.params, ...(options?.params || {}) },
      headers: options?.headers,
    });
  }

  /**
   * Fetch the full review for a gated block run.
   *
   * @param runId - The workflow run id.
   * @param blockId - The gated block id.
   * @returns The full review, including version history, the terminal decision.
   * @throws `APIError` with `.status === 404` when no review exists for this block yet.
   */
  async get(runId: string, blockId: string, options?: RequestOptions): Promise<ReviewOverlay> {
    const request = this.prepare_get(runId, blockId);
    return this._fetchJson(ZReviewOverlay, {
      url: request.url,
      method: request.method,
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Approve the gated block output.
   *
   * @param runId - The workflow run id.
   * @param blockId - The gated block id.
   * @param versionId - Content-hash id of the exact version being approved.
   * @returns Submission status and the updated review.
   * @throws `APIError` with `.status === 409` when the review already has a terminal decision.
   */
  async approve(
    runId: string,
    blockId: string,
    {
      versionId,
    }: {
      versionId: string;
    },
    options?: RequestOptions
  ): Promise<SubmitDecisionResponse> {
    return this._decision(
      runId,
      blockId,
      {
        verdict: 'approved',
        versionId,
      },
      options
    );
  }

  /**
   * Reject the gated block output.
   *
   * @param runId - The workflow run id.
   * @param blockId - The gated block id.
   * @param versionId - Content-hash id of the exact version being rejected.
   * @param reason - Why the output was rejected — required so every rejection is auditable.
   * @returns Submission status and the updated review.
   * @throws `APIError` with `.status === 409` when the review already has a terminal decision.
   */
  async reject(
    runId: string,
    blockId: string,
    {
      versionId,
      reason,
    }: {
      versionId: string;
      reason: string;
    },
    options?: RequestOptions
  ): Promise<SubmitDecisionResponse> {
    return this._decision(runId, blockId, { verdict: 'rejected', versionId, reason }, options);
  }

  /**
   * Append a new output version to the review's version history.
   *
   * @param runId - The workflow run id.
   * @param blockId - The gated block id.
   * @param snapshot - The new output payload to record as a version.
   * @param parentId - Content-hash id of the parent version.
   * @param note - Optional free-text note attached to the version.
   * @returns The review with the new version appended.
   * @throws `APIError` with `.status === 409` when the review already has a terminal decision.
   */
  async createVersion(
    runId: string,
    blockId: string,
    {
      snapshot,
      parentId,
      note,
    }: {
      snapshot: Record<string, unknown>;
      parentId: string;
      note?: string | null;
    },
    options?: RequestOptions
  ): Promise<ReviewOverlay> {
    const request = this.prepare_create_version(runId, blockId, {
      snapshot,
      parentId,
      note,
    });

    return this._fetchJson(ZReviewOverlay, {
      url: request.url,
      method: request.method,
      body: withoutOrigin({ ...request.body, ...(options?.body || {}) }),
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Poll until the block is gated and awaiting review.
   *
   * Calls {@link get} on a loop until the review exists and has no terminal
   * decision. A 404 means the workflow has not reached the gate yet —
   * that is not an error, polling continues.
   *
   * @param runId - The workflow run id.
   * @param blockId - The gated block id.
   * @param timeoutMs - Maximum milliseconds to wait before giving up. Defaults to 120000.
   * @param pollIntervalMs - Milliseconds to sleep between polls. Defaults to 2000.
   * @returns The review once it is awaiting review.
   * @throws `Error` when the review did not reach `awaiting_review` in time.
   */
  async waitFor(
    runId: string,
    blockId: string,
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
        const overlay = await this.get(runId, blockId, options);
        if (overlay.decision === null) {
          return overlay;
        }
      } catch (error) {
        if (!(error instanceof APIError && error.status === 404)) {
          throw error;
        }
      }
      if (Date.now() >= deadline) {
        throw new Error(
          `Review for run '${runId}' block '${blockId}' ` +
            `was not awaiting_review within ${timeoutMs}ms`
        );
      }
      await new Promise((resolve) => setTimeout(resolve, pollIntervalMs));
    }
  }

  async wait_for(
    runId: string,
    blockId: string,
    options?: {
      timeoutMs?: number;
      pollIntervalMs?: number;
    },
    requestOptions?: RequestOptions
  ): Promise<ReviewOverlay> {
    return this.waitFor(runId, blockId, options, requestOptions);
  }

  /** Shared verdict submission against `POST /workflows/reviews/{runId}/{blockId}/decision`. */
  private async _decision(
    runId: string,
    blockId: string,
    {
      verdict,
      versionId,
      reason,
    }: {
      verdict: 'approved' | 'rejected';
      versionId: string;
      reason?: string;
    },
    options?: RequestOptions
  ): Promise<SubmitDecisionResponse> {
    const request = this.prepare_decision(runId, blockId, {
      verdict,
      versionId,
      reason,
    });

    return this._fetchJson(ZSubmitDecisionResponse, {
      url: request.url,
      method: request.method,
      body: { ...request.body, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }
}
