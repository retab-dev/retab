import { APIError, CompositionClient, RequestOptions } from '../../../client.js';
import {
  ReviewOverlay,
  ReviewQueueResponse,
  SubmitDecisionResponse,
  ZReviewOverlay,
  ZReviewQueueResponse,
  ZSubmitDecisionResponse,
} from '../../../types.js';

/** Overlay lifecycle filter accepted by {@link APIWorkflowReviews.list}. */
export type ReviewStatus = 'awaiting_review' | 'approved' | 'rejected';

/** Provenance accepted when posting a new version through {@link APIWorkflowReviews.edit}. */
export type EditOrigin = 'human_edit' | 'agent_edit';

type PreparedReviewRequest = {
  url: string;
  method: 'GET' | 'POST';
  params?: Record<string, unknown>;
  body?: Record<string, unknown>;
};

/**
 * Actor-neutral client for the workflow review overlay.
 *
 * Drives the review loop served under `/workflows/reviews`: list the queue,
 * fetch an overlay, post new output versions, and submit verdicts.
 *
 * Actor symmetry is a hard rule here. A proposal authored by a model, an
 * agent, or a human goes through the SAME {@link edit} / {@link approve} pair.
 * `Actor.kind` is data carried on the overlay — no method, parameter, or branch
 * in this client depends on it.
 *
 * Every mutating call carries a `versionStamp` (the overlay's `rev`). If the
 * server's `rev` has advanced since the caller last read it, the request fails
 * with `APIError` whose `.status` is `409` — re-read with {@link get} and retry.
 *
 * Usage: `client.workflows.reviews.list()`
 */
export default class APIWorkflowReviews extends CompositionClient {
  constructor(client: CompositionClient) {
    super(client);
  }

  prepare_list(
    workflowId?: string,
    status: ReviewStatus = 'awaiting_review',
    mine = false,
    limit = 50
  ): PreparedReviewRequest {
    const params: Record<string, unknown> = { status, mine, limit };
    if (workflowId !== undefined) params.workflow_id = workflowId;
    return { url: '/workflows/reviews', method: 'GET', params };
  }

  prepare_get(runId: string, blockId: string): PreparedReviewRequest {
    return { url: `/workflows/reviews/${runId}/${blockId}`, method: 'GET' };
  }

  prepare_edit(
    runId: string,
    blockId: string,
    {
      snapshot,
      reviewableValue,
      versionStamp,
      origin = 'human_edit',
      note,
      commandId,
    }: {
      snapshot?: Record<string, unknown> | null;
      reviewableValue?: Record<string, unknown> | null;
      versionStamp: number;
      origin?: EditOrigin;
      note?: string | null;
      commandId?: string;
    }
  ): PreparedReviewRequest {
    return {
      url: `/workflows/reviews/${runId}/${blockId}/versions`,
      method: 'POST',
      body: {
        snapshot: snapshot ?? null,
        reviewable_value: reviewableValue ?? null,
        version_stamp: versionStamp,
        origin,
        note: note ?? null,
        command_id: commandId ?? null,
      },
    };
  }

  prepare_decision(
    runId: string,
    blockId: string,
    {
      verdict,
      versionStamp,
      editedOutput,
      reviewableValue,
      onSeq,
      effectiveSeq,
      reason,
      commandId,
    }: {
      verdict: 'approved' | 'rejected';
      versionStamp: number;
      editedOutput?: Record<string, unknown> | null;
      reviewableValue?: Record<string, unknown> | null;
      onSeq?: number;
      effectiveSeq?: number;
      reason?: string;
      commandId?: string;
    }
  ): PreparedReviewRequest {
    return {
      url: `/workflows/reviews/${runId}/${blockId}/decision`,
      method: 'POST',
      body: {
        verdict,
        version_stamp: versionStamp,
        edited_output: editedOutput ?? null,
        reviewable_value: reviewableValue ?? null,
        on_seq: onSeq ?? null,
        effective_seq: effectiveSeq ?? null,
        reason: reason ?? null,
        command_id: commandId ?? null,
      },
    };
  }

  prepare_claim(
    runId: string,
    blockId: string,
    {
      versionStamp,
      ttlSeconds = 900,
    }: {
      versionStamp: number;
      ttlSeconds?: number;
    }
  ): PreparedReviewRequest {
    return {
      url: `/workflows/reviews/${runId}/${blockId}/claim`,
      method: 'POST',
      body: {
        version_stamp: versionStamp,
        ttl_seconds: ttlSeconds,
      },
    };
  }

  prepare_release(
    runId: string,
    blockId: string,
    {
      versionStamp,
    }: {
      versionStamp: number;
    }
  ): PreparedReviewRequest {
    return {
      url: `/workflows/reviews/${runId}/${blockId}/release`,
      method: 'POST',
      body: { version_stamp: versionStamp },
    };
  }

  /**
   * List review-queue items.
   *
   * @param workflowId - Restrict the queue to a single workflow.
   * @param status - Lifecycle filter: `awaiting_review` / `approved` / `rejected`. Defaults to `awaiting_review`.
   * @param mine - When true, only items claimed by the calling actor.
   * @param limit - Page size (1-200). Defaults to 50.
   * @returns A page of lightweight `ReviewQueueItem` summaries plus a `has_more` flag.
   *
   * @example
   * ```typescript
   * const queue = await client.workflows.reviews.list({ mine: true });
   * ```
   */
  async list(
    {
      workflowId,
      status = 'awaiting_review',
      mine = false,
      limit = 50,
    }: {
      workflowId?: string;
      status?: ReviewStatus;
      mine?: boolean;
      limit?: number;
    } = {},
    options?: RequestOptions
  ): Promise<ReviewQueueResponse> {
    const request = this.prepare_list(workflowId, status, mine, limit);

    return this._fetchJson(ZReviewQueueResponse, {
      url: request.url,
      method: request.method,
      params: { ...request.params, ...(options?.params || {}) },
      headers: options?.headers,
    });
  }

  /**
   * Fetch the full review overlay for a gated block run.
   *
   * @param runId - The workflow run id.
   * @param blockId - The gated block id.
   * @returns The full overlay, including version history, decisions, and audit log.
   * @throws `APIError` with `.status === 404` when no overlay exists for this block yet.
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
   * @param versionStamp - The overlay `rev` last observed (CAS token).
   * @param editedOutput - Optional replacement output applied with the approval.
   * @param reviewableValue - Optional primitive-specific value compiled by the server.
   * @param onSeq - Version sequence the decision is made against.
   * @param effectiveSeq - Version that becomes effective on apply.
   * @param commandId - Optional idempotency key for deduplicating submissions.
   * @returns Submission status and the updated overlay.
   * @throws `APIError` with `.status === 409` when `versionStamp` is stale — re-read and retry.
   */
  async approve(
    runId: string,
    blockId: string,
    {
      versionStamp,
      editedOutput,
      reviewableValue,
      onSeq,
      effectiveSeq,
      commandId,
    }: {
      versionStamp: number;
      editedOutput?: Record<string, unknown> | null;
      reviewableValue?: Record<string, unknown> | null;
      onSeq?: number;
      effectiveSeq?: number;
      commandId?: string;
    },
    options?: RequestOptions
  ): Promise<SubmitDecisionResponse> {
    return this._decision(
      runId,
      blockId,
      {
        verdict: 'approved',
        versionStamp,
        editedOutput,
        reviewableValue,
        onSeq,
        effectiveSeq,
        commandId,
      },
      options
    );
  }

  /**
   * Reject the gated block output.
   *
   * @param runId - The workflow run id.
   * @param blockId - The gated block id.
   * @param versionStamp - The overlay `rev` last observed (CAS token).
   * @param reason - Why the output was rejected — required so every rejection is auditable.
   * @param commandId - Optional idempotency key for deduplicating submissions.
   * @returns Submission status and the updated overlay.
   * @throws `APIError` with `.status === 409` when `versionStamp` is stale — re-read and retry.
   */
  async reject(
    runId: string,
    blockId: string,
    {
      versionStamp,
      reason,
      commandId,
    }: {
      versionStamp: number;
      reason: string;
      commandId?: string;
    },
    options?: RequestOptions
  ): Promise<SubmitDecisionResponse> {
    return this._decision(
      runId,
      blockId,
      { verdict: 'rejected', versionStamp, reason, commandId },
      options
    );
  }

  /**
   * Append a new output version to the overlay's version history.
   *
   * A proposal authored by a model, an agent, or a human uses this same call —
   * the only difference is `origin`, which is descriptive provenance, not a
   * behavioral switch.
   *
   * @param runId - The workflow run id.
   * @param blockId - The gated block id.
   * @param snapshot - The new output payload to record as a version.
   * @param reviewableValue - Primitive-specific reviewed value compiled by the server.
   * @param versionStamp - The overlay `rev` last observed (CAS token).
   * @param origin - Provenance of the snapshot: `human_edit` or `agent_edit`. Defaults to `human_edit`.
   * @param note - Optional free-text note attached to the version.
   * @param commandId - Optional idempotency key for deduplicating submissions.
   * @returns The overlay with the new version appended.
   * @throws `APIError` with `.status === 409` when `versionStamp` is stale — re-read and retry.
   */
  async edit(
    runId: string,
    blockId: string,
    {
      snapshot,
      reviewableValue,
      versionStamp,
      origin = 'human_edit',
      note,
      commandId,
    }: {
      snapshot?: Record<string, unknown> | null;
      reviewableValue?: Record<string, unknown> | null;
      versionStamp: number;
      origin?: EditOrigin;
      note?: string | null;
      commandId?: string;
    },
    options?: RequestOptions
  ): Promise<ReviewOverlay> {
    const request = this.prepare_edit(runId, blockId, {
      snapshot,
      reviewableValue,
      versionStamp,
      origin,
      note,
      commandId,
    });

    return this._fetchJson(ZReviewOverlay, {
      url: request.url,
      method: request.method,
      body: { ...request.body, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Take the advisory soft lock on the overlay.
   *
   * @param runId - The workflow run id.
   * @param blockId - The gated block id.
   * @param versionStamp - The overlay `rev` last observed (CAS token).
   * @param ttlSeconds - How long the claim is held before it lapses. Defaults to 900.
   * @returns The overlay with the claim recorded.
   * @throws `APIError` with `.status === 409` when `versionStamp` is stale — re-read and retry.
   */
  async claim(
    runId: string,
    blockId: string,
    {
      versionStamp,
      ttlSeconds = 900,
    }: {
      versionStamp: number;
      ttlSeconds?: number;
    },
    options?: RequestOptions
  ): Promise<ReviewOverlay> {
    const request = this.prepare_claim(runId, blockId, { versionStamp, ttlSeconds });

    return this._fetchJson(ZReviewOverlay, {
      url: request.url,
      method: request.method,
      body: { ...request.body, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Release the advisory soft lock on the overlay.
   *
   * @param runId - The workflow run id.
   * @param blockId - The gated block id.
   * @param versionStamp - The overlay `rev` last observed (CAS token).
   * @returns The overlay with the claim cleared.
   * @throws `APIError` with `.status === 409` when `versionStamp` is stale — re-read and retry.
   */
  async release(
    runId: string,
    blockId: string,
    {
      versionStamp,
    }: {
      versionStamp: number;
    },
    options?: RequestOptions
  ): Promise<ReviewOverlay> {
    const request = this.prepare_release(runId, blockId, { versionStamp });

    return this._fetchJson(ZReviewOverlay, {
      url: request.url,
      method: request.method,
      body: { ...request.body, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Poll until the block is gated and awaiting review.
   *
   * Calls {@link get} on a loop until the overlay exists and its `status` is
   * `awaiting_review`. A 404 means the workflow has not reached the gate yet —
   * that is not an error, polling continues.
   *
   * @param runId - The workflow run id.
   * @param blockId - The gated block id.
   * @param timeoutMs - Maximum milliseconds to wait before giving up. Defaults to 120000.
   * @param pollIntervalMs - Milliseconds to sleep between polls. Defaults to 2000.
   * @returns The overlay once it is awaiting review.
   * @throws `Error` when the overlay did not reach `awaiting_review` in time.
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
        if (overlay.status === 'awaiting_review') {
          return overlay;
        }
      } catch (error) {
        if (!(error instanceof APIError && error.status === 404)) {
          throw error;
        }
      }
      if (Date.now() >= deadline) {
        throw new Error(
          `Review overlay for run '${runId}' block '${blockId}' ` +
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
      versionStamp,
      editedOutput,
      reviewableValue,
      onSeq,
      effectiveSeq,
      reason,
      commandId,
    }: {
      verdict: 'approved' | 'rejected';
      versionStamp: number;
      editedOutput?: Record<string, unknown> | null;
      reviewableValue?: Record<string, unknown> | null;
      onSeq?: number;
      effectiveSeq?: number;
      reason?: string;
      commandId?: string;
    },
    options?: RequestOptions
  ): Promise<SubmitDecisionResponse> {
    const request = this.prepare_decision(runId, blockId, {
      verdict,
      versionStamp,
      editedOutput,
      reviewableValue,
      onSeq,
      effectiveSeq,
      reason,
      commandId,
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
