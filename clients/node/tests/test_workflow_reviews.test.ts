import { describe, expect, test } from 'bun:test';

import APIWorkflowReviews from '../src/api/workflows/reviews/client';
import { AbstractClient } from '../src/client';
import {
  ZActor,
  ZDecision,
  ZReview,
  ZWorkflowReviewQueue,
  ZReviewSummary,
  ZReviewVersion,
  ZReviewVersionListResponse,
  ZSubmitDecisionResponse,
} from '../src/types';

class MockClient extends AbstractClient {
  public lastFetchParams: Record<string, unknown> | null = null;
  private responseBody: unknown;
  private status: number;

  constructor(responseBody: unknown, status = 200) {
    super();
    this.responseBody = responseBody;
    this.status = status;
  }

  protected async _fetch(params: {
    url: string;
    method: string;
    params?: Record<string, unknown>;
    headers?: Record<string, unknown>;
    body?: Record<string, unknown>;
  }): Promise<Response> {
    this.lastFetchParams = params;
    return new Response(JSON.stringify(this.responseBody), {
      status: this.status,
      headers: { 'Content-Type': 'application/json' },
    });
  }
}

const REVIEW_ID = 'rev_01HX9A7Y1R6G9J2K8M4P5Q6T7V';
const VERSION_ID = 'rvr_2N9S8Q4F6M1K7C3D5R0T8W6Y2Z';
const CHILD_VERSION_ID = 'rvr_5M8E2K7V9Q1C4T6W3D0R2Y5N8A';

const ACTOR_JSON = {
  kind: 'agent',
  id: 'agent_7',
  display_name: 'Reviewer Agent',
};

const REVIEW_VERSION_JSON = {
  id: VERSION_ID,
  review_id: REVIEW_ID,
  parent_id: null,
  author: { kind: 'model', id: 'model_1', display_name: 'Model' },
  snapshot: { output: { invoice_number: 'INV-001', total: 100 } },
  note: null,
  created_at: '2026-05-18T10:00:00Z',
};

const DECISION_JSON = {
  verdict: 'approved',
  version_id: VERSION_ID,
  author: ACTOR_JSON,
  decided_at: '2026-05-18T10:05:00Z',
  reason: null,
};

const REVIEW_JSON = {
  id: REVIEW_ID,
  workflow_id: 'wf_1',
  workflow_version_id: 'wfv_1',
  workflow_run_id: 'run_1',
  block_id: 'extract-1',
  step_id: 'step_1',
  parent_step_id: null,
  iteration_key: null,
  block_type: 'extract',
  triggered_by: { kind: 'confidence_below', threshold: 0.8 },
  created_at: '2026-05-18T09:58:00Z',
  decision: DECISION_JSON,
};

const REVIEW_SUMMARY_JSON = {
  ...REVIEW_JSON,
  decision: null,
  seed_version_id: VERSION_ID,
  version_count: 1,
};

const CHILD_REVIEW_VERSION_JSON = {
  ...REVIEW_VERSION_JSON,
  id: CHILD_VERSION_ID,
  parent_id: VERSION_ID,
  author: ACTOR_JSON,
  snapshot: { output: { invoice_number: 'INV-001', total: 150 } },
};

describe('review zod schemas', () => {
  test('ZActor parses model / agent / human kinds identically', () => {
    for (const kind of ['model', 'agent', 'human'] as const) {
      const actor = ZActor.parse({ ...ACTOR_JSON, kind });
      expect(actor.kind).toBe(kind);
    }
  });

  test('ZReviewVersion round-trips a full top-level version', () => {
    const version = ZReviewVersion.parse(REVIEW_VERSION_JSON);
    expect(version.id).toBe(VERSION_ID);
    expect(version.review_id).toBe(REVIEW_ID);
    expect(version.parent_id).toBe(null);
    expect(version.author.kind).toBe('model');
    expect(version.snapshot).toEqual({ output: { invoice_number: 'INV-001', total: 100 } });
  });

  test('ZReviewVersion rejects removed origin fields', () => {
    expect(() =>
      ZReviewVersion.parse({ ...REVIEW_VERSION_JSON, origin: 'model_output' })
    ).toThrow();
  });

  test('ZDecision round-trips an approved decision with reason null', () => {
    const decision = ZDecision.parse(DECISION_JSON);
    expect(decision.verdict).toBe('approved');
    expect(decision.version_id).toBe(VERSION_ID);
    expect(decision.reason).toBeNull();
  });

  test('ZDecision round-trips a rejected decision with a non-empty reason', () => {
    const decision = ZDecision.parse({
      ...DECISION_JSON,
      verdict: 'rejected',
      reason: 'wrong total',
    });
    expect(decision.verdict).toBe('rejected');
    expect(decision.reason).toBe('wrong total');
  });

  test('ZDecision rejects an approved decision that carries a reason', () => {
    expect(() =>
      ZDecision.parse({ ...DECISION_JSON, verdict: 'approved', reason: 'should not be set' })
    ).toThrow();
  });

  test('ZDecision rejects a rejected decision missing a reason', () => {
    expect(() =>
      ZDecision.parse({ ...DECISION_JSON, verdict: 'rejected', reason: null })
    ).toThrow();
    expect(() => ZDecision.parse({ ...DECISION_JSON, verdict: 'rejected', reason: '' })).toThrow();
  });

  test('ZReview round-trips review metadata and decision without embedded versions', () => {
    const review = ZReview.parse(REVIEW_JSON);
    expect(review.id).toBe(REVIEW_ID);
    expect('versions' in review).toBe(false);
    expect(review.decision?.version_id).toBe(VERSION_ID);
  });

  test('ZReview rejects embedded versions from the pre-cutover shape', () => {
    expect(() =>
      ZReview.parse({
        ...REVIEW_JSON,
        versions: { [VERSION_ID]: REVIEW_VERSION_JSON },
      })
    ).toThrow();
  });

  test('ZReview accepts for_each reviews addressed by review id', () => {
    const review = ZReview.parse({
      ...REVIEW_JSON,
      block_id: 'blk_for_each',
      step_id: 'step_child',
      parent_step_id: 'step_parent',
      iteration_key: 'line-1',
      block_type: 'for_each',
    });
    expect(review.block_type).toBe('for_each');
    expect(review.parent_step_id).toBe('step_parent');
    expect(review.iteration_key).toBe('line-1');
  });

  test('ZReviewSummary carries seed_version_id and version_count', () => {
    const summary = ZReviewSummary.parse(REVIEW_SUMMARY_JSON);
    expect(summary.id).toBe(REVIEW_ID);
    expect(summary.seed_version_id).toBe(VERSION_ID);
    expect(summary.version_count).toBe(1);
    expect('versions' in (summary as Record<string, unknown>)).toBe(false);
  });

  test('ZReviewSummary rejects rows missing seed_version_id', () => {
    const { seed_version_id: _seed, ...withoutSeed } = REVIEW_SUMMARY_JSON;
    expect(() => ZReviewSummary.parse(withoutSeed)).toThrow();
  });

  test('ZReviewSummary rejects rows missing version_count', () => {
    const { version_count: _count, ...withoutCount } = REVIEW_SUMMARY_JSON;
    expect(() => ZReviewSummary.parse(withoutCount)).toThrow();
  });

  test('ZWorkflowReviewQueue wraps ReviewSummary rows with seed_version_id and version_count', () => {
    const response = ZWorkflowReviewQueue.parse({
      data: [REVIEW_SUMMARY_JSON],
      list_metadata: { before: null, after: REVIEW_ID },
    });
    expect(response.list_metadata.before).toBeNull();
    expect(response.list_metadata.after).toBe(REVIEW_ID);
    expect(response.data[0]?.seed_version_id).toBe(VERSION_ID);
    expect(response.data[0]?.version_count).toBe(1);
  });

  test('ZWorkflowReviewQueue rejects pre-cutover Review rows that omit summary fields', () => {
    expect(() =>
      ZWorkflowReviewQueue.parse({
        data: [{ ...REVIEW_JSON, decision: null }],
        list_metadata: { before: null, after: null },
      })
    ).toThrow();
  });

  test('ZReviewVersionListResponse wraps flat versions for one review', () => {
    const parsed = ZReviewVersionListResponse.parse({
      data: [REVIEW_VERSION_JSON, CHILD_REVIEW_VERSION_JSON],
      list_metadata: { before: null, after: CHILD_VERSION_ID },
    });
    expect(parsed.data.map((version) => version.id)).toEqual([VERSION_ID, CHILD_VERSION_ID]);
    expect(parsed.data[1]?.parent_id).toBe(VERSION_ID);
  });

  test('ZSubmitDecisionResponse round-trips submission status + review + resume_status', () => {
    const parsed = ZSubmitDecisionResponse.parse({
      submission_status: 'accepted',
      review: REVIEW_JSON,
      resume_status: 'resumed',
      resume_error: null,
    });
    expect(parsed.submission_status).toBe('accepted');
    expect(parsed.review.decision?.version_id).toBe(VERSION_ID);
    expect(parsed.resume_status).toBe('resumed');
    expect(parsed.resume_error).toBe(null);
  });

  test('ZSubmitDecisionResponse defaults resume_status to resumed when backend omits it', () => {
    const parsed = ZSubmitDecisionResponse.parse({
      submission_status: 'accepted',
      review: REVIEW_JSON,
    });
    expect(parsed.resume_status).toBe('resumed');
    expect(parsed.resume_error).toBe(null);
  });

  test('ZSubmitDecisionResponse surfaces resume_status="pending" with an error message', () => {
    const parsed = ZSubmitDecisionResponse.parse({
      submission_status: 'accepted',
      review: REVIEW_JSON,
      resume_status: 'pending',
      resume_error: 'Workflow run not found for run_id=run_1',
    });
    expect(parsed.resume_status).toBe('pending');
    expect(parsed.resume_error).toContain('Workflow run not found');
  });

  test('ZSubmitDecisionResponse rejects noncanonical decision/resume statuses', () => {
    expect(() =>
      ZSubmitDecisionResponse.parse({
        submission_status: 'queued',
        review: REVIEW_JSON,
        resume_status: 'pending',
        resume_error: 'Workflow run not found',
      })
    ).toThrow();
    expect(() =>
      ZSubmitDecisionResponse.parse({
        submission_status: 'accepted',
        review: REVIEW_JSON,
        resume_status: 'queued',
        resume_error: 'Workflow run not found',
      })
    ).toThrow();
  });
});

describe('APIWorkflowReviews request shapes', () => {
  test('list() builds the queue GET with snake_case params', async () => {
    const mock = new MockClient({
      data: [REVIEW_SUMMARY_JSON],
      list_metadata: { before: null, after: null },
    });
    const reviews = new APIWorkflowReviews(mock);

    await reviews.list({
      workflowId: 'wf_1',
      runId: 'run_1',
      blockId: 'extract-1',
      limit: 25,
    });

    expect(mock.lastFetchParams).toEqual({
      url: '/workflows/reviews',
      method: 'GET',
      params: {
        decision_status: 'pending',
        limit: 25,
        workflow_id: 'wf_1',
        run_id: 'run_1',
        block_id: 'extract-1',
      },
      headers: undefined,
    });
  });

  test('list() defaults to limit=50 and decision_status="pending"', async () => {
    const mock = new MockClient({
      data: [],
      list_metadata: { before: null, after: null },
    });
    const reviews = new APIWorkflowReviews(mock);

    await reviews.list();

    expect(mock.lastFetchParams?.params).toEqual({ decision_status: 'pending', limit: 50 });
  });

  test('list({ decisionStatus: "all" }) includes every decision state', async () => {
    const mock = new MockClient({
      data: [],
      list_metadata: { before: null, after: null },
    });
    const reviews = new APIWorkflowReviews(mock);

    await reviews.list({ decisionStatus: 'all' });

    expect(mock.lastFetchParams?.params).toEqual({ decision_status: 'all', limit: 50 });
  });

  test('list({ after, stepId, iterationKey }) forwards cursors and execution filters', async () => {
    const mock = new MockClient({
      data: [REVIEW_SUMMARY_JSON],
      list_metadata: { before: null, after: REVIEW_ID },
    });
    const reviews = new APIWorkflowReviews(mock);

    const result = await reviews.list({
      after: 'rev_prev',
      limit: 25,
      stepId: 'step_1',
      iterationKey: 'line-1',
    });

    expect(mock.lastFetchParams?.params).toEqual({
      decision_status: 'pending',
      limit: 25,
      step_id: 'step_1',
      iteration_key: 'line-1',
      after: 'rev_prev',
    });
    expect(result.list_metadata.after).toBe(REVIEW_ID);
    expect(result.list_metadata.before).toBeNull();
    expect(result.data[0]?.seed_version_id).toBe(VERSION_ID);
    expect(result.data[0]?.version_count).toBe(1);
  });

  test('get() builds the review GET route', async () => {
    const mock = new MockClient(REVIEW_JSON);
    const reviews = new APIWorkflowReviews(mock);

    const review = await reviews.get(REVIEW_ID);

    expect(mock.lastFetchParams).toEqual({
      url: `/workflows/reviews/${REVIEW_ID}`,
      method: 'GET',
      params: undefined,
      headers: undefined,
    });
    expect(review.id).toBe(REVIEW_ID);
    expect('versions' in (review as Record<string, unknown>)).toBe(false);
  });

  test('approve() posts version_id to the approve endpoint and surfaces resume_status', async () => {
    const mock = new MockClient({
      submission_status: 'accepted',
      review: REVIEW_JSON,
      resume_status: 'resumed',
      resume_error: null,
    });
    const reviews = new APIWorkflowReviews(mock);

    const result = await reviews.approve(REVIEW_ID, { versionId: VERSION_ID });

    expect(mock.lastFetchParams).toEqual({
      url: `/workflows/reviews/${REVIEW_ID}/approve`,
      method: 'POST',
      body: {
        version_id: VERSION_ID,
      },
      params: undefined,
      headers: undefined,
    });
    expect(result.submission_status).toBe('accepted');
    expect(result.resume_status).toBe('resumed');
    expect(result.review.decision?.version_id).toBe(VERSION_ID);
  });

  test('reject() posts version_id and reason to the reject endpoint', async () => {
    const mock = new MockClient({
      submission_status: 'accepted',
      review: REVIEW_JSON,
      resume_status: 'resumed',
      resume_error: null,
    });
    const reviews = new APIWorkflowReviews(mock);

    await reviews.reject(REVIEW_ID, { versionId: VERSION_ID, reason: 'wrong total' });

    expect(mock.lastFetchParams?.url).toBe(`/workflows/reviews/${REVIEW_ID}/reject`);
    expect(mock.lastFetchParams?.body).toEqual({
      version_id: VERSION_ID,
      reason: 'wrong total',
    });
  });

  test('approve() omits the reason key', async () => {
    const mock = new MockClient({
      submission_status: 'accepted',
      review: REVIEW_JSON,
      resume_status: 'resumed',
      resume_error: null,
    });
    const reviews = new APIWorkflowReviews(mock);

    await reviews.approve(REVIEW_ID, { versionId: VERSION_ID });

    const body = mock.lastFetchParams?.body as Record<string, unknown>;
    expect('reason' in body).toBe(false);
  });

  test('versions.create() omits the note key when none is provided', async () => {
    const mock = new MockClient(CHILD_REVIEW_VERSION_JSON);
    const reviews = new APIWorkflowReviews(mock);

    await reviews.versions.create({
      reviewId: REVIEW_ID,
      snapshot: { category: 'Invoice' },
      parentVersionId: VERSION_ID,
    });

    const body = mock.lastFetchParams?.body as Record<string, unknown>;
    expect('note' in body).toBe(false);
    expect('origin' in body).toBe(false);
    expect(body.parent_id).toBe(VERSION_ID);
  });

  test('versions.create() requires a parentVersionId', () => {
    const mock = new MockClient(CHILD_REVIEW_VERSION_JSON);
    const reviews = new APIWorkflowReviews(mock);

    expect(() =>
      reviews.versions.prepare_create({
        reviewId: REVIEW_ID,
        snapshot: { category: 'Invoice' },
      } as Parameters<typeof reviews.versions.prepare_create>[0])
    ).toThrow('parentVersionId is required');
  });

  test('versions.create() posts a new snapshot version to the flat versions route', async () => {
    const mock = new MockClient(CHILD_REVIEW_VERSION_JSON);
    const reviews = new APIWorkflowReviews(mock);

    const result = await reviews.versions.create({
      reviewId: REVIEW_ID,
      snapshot: { category: 'Invoice' },
      parentVersionId: VERSION_ID,
      note: 'agent proposal',
    });

    expect(mock.lastFetchParams).toEqual({
      url: '/workflows/reviews/versions',
      method: 'POST',
      body: {
        review_id: REVIEW_ID,
        parent_id: VERSION_ID,
        snapshot: { category: 'Invoice' },
        note: 'agent proposal',
      },
      params: undefined,
      headers: undefined,
    });
    expect(result.id).toBe(CHILD_VERSION_ID);
  });

  test('versions.get() fetches one flat review version', async () => {
    const mock = new MockClient(REVIEW_VERSION_JSON);
    const reviews = new APIWorkflowReviews(mock);

    const version = await reviews.versions.get(VERSION_ID);

    expect(mock.lastFetchParams).toEqual({
      url: `/workflows/reviews/versions/${VERSION_ID}`,
      method: 'GET',
      params: undefined,
      headers: undefined,
    });
    expect(version.review_id).toBe(REVIEW_ID);
  });

  test('versions.list() requires review_id and forwards cursors', async () => {
    const mock = new MockClient({
      data: [REVIEW_VERSION_JSON, CHILD_REVIEW_VERSION_JSON],
      list_metadata: { before: null, after: CHILD_VERSION_ID },
    });
    const reviews = new APIWorkflowReviews(mock);

    const result = await reviews.versions.list({
      reviewId: REVIEW_ID,
      after: VERSION_ID,
      limit: 25,
    });

    expect(mock.lastFetchParams).toEqual({
      url: '/workflows/reviews/versions',
      method: 'GET',
      params: {
        review_id: REVIEW_ID,
        after: VERSION_ID,
        limit: 25,
      },
      headers: undefined,
    });
    expect(result.data[1]?.id).toBe(CHILD_VERSION_ID);
  });

  test('wait helpers are not exposed', () => {
    const mock = new MockClient(REVIEW_JSON);
    const reviews = new APIWorkflowReviews(mock);

    expect('waitFor' in reviews).toBe(false);
    expect('wait_for' in reviews).toBe(false);
  });
});
