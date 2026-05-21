import { describe, expect, test } from 'bun:test';

import APIWorkflowReviews from '../src/api/workflows/reviews/client';
import { AbstractClient } from '../src/client';
import {
  ZActor,
  ZAppendVersionResponse,
  ZOutputVersion,
  ZReviewDecision,
  ZReviewOverlay,
  ZReviewQueueItem,
  ZReviewQueueResponse,
  ZReviewSummary,
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
const VERSION_ID = 'ver_2N9S8Q4F6M1K7C3D5R0T8W6Y2Z';
const CHILD_VERSION_ID = 'ver_5M8E2K7V9Q1C4T6W3D0R2Y5N8A';

const ACTOR_JSON = {
  kind: 'agent',
  id: 'agent_7',
  display_name: 'Reviewer Agent',
};

const OUTPUT_VERSION_JSON = {
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

const OVERLAY_JSON = {
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
  awaiting_since: '2026-05-18T09:58:00Z',
  priority: 5,
  versions: {
    [VERSION_ID]: OUTPUT_VERSION_JSON,
    [CHILD_VERSION_ID]: {
      ...OUTPUT_VERSION_JSON,
      parent_id: VERSION_ID,
      author: ACTOR_JSON,
      snapshot: { output: { invoice_number: 'INV-001', total: 150 } },
    },
  },
  decision: DECISION_JSON,
};

const QUEUE_ITEM_JSON = {
  id: REVIEW_ID,
  workflow_id: 'wf_1',
  workflow_run_id: 'run_1',
  block_id: 'extract-1',
  step_id: 'step_1',
  parent_step_id: null,
  iteration_key: null,
  block_type: 'extract',
  triggered_by: { kind: 'always' },
  awaiting_since: '2026-05-18T09:58:00Z',
  priority: 5,
  seed_version_id: VERSION_ID,
  version_count: 1,
  decision: null,
};

describe('review zod schemas', () => {
  test('ZActor parses model / agent / human kinds identically', () => {
    for (const kind of ['model', 'agent', 'human'] as const) {
      const actor = ZActor.parse({ ...ACTOR_JSON, kind });
      expect(actor.kind).toBe(kind);
    }
  });

  test('ZOutputVersion round-trips a full version', () => {
    const version = ZOutputVersion.parse(OUTPUT_VERSION_JSON);
    expect(version.parent_id).toBe(null);
    expect(version.author.kind).toBe('model');
    expect(version.snapshot).toEqual({ output: { invoice_number: 'INV-001', total: 100 } });
  });

  test('ZOutputVersion rejects removed origin fields', () => {
    expect(() =>
      ZOutputVersion.parse({ ...OUTPUT_VERSION_JSON, origin: 'model_output' })
    ).toThrow();
  });

  test('ZReviewDecision round-trips and carries version_id', () => {
    for (const verdict of ['approved', 'rejected'] as const) {
      const decision = ZReviewDecision.parse({ ...DECISION_JSON, verdict });
      expect(decision.verdict).toBe(verdict);
      expect(decision.version_id).toBe(VERSION_ID);
    }
  });

  test('ZReviewOverlay round-trips versions and decision', () => {
    const overlay = ZReviewOverlay.parse(OVERLAY_JSON);
    expect(overlay.id).toBe(REVIEW_ID);
    expect(Object.keys(overlay.versions)).toEqual([VERSION_ID, CHILD_VERSION_ID]);
    expect(overlay.versions[CHILD_VERSION_ID]?.parent_id).toBe(VERSION_ID);
    expect(overlay.decision?.version_id).toBe(VERSION_ID);
  });

  test('ZReviewOverlay accepts for_each reviews addressed by review id', () => {
    const overlay = ZReviewOverlay.parse({
      ...OVERLAY_JSON,
      block_id: 'blk_for_each',
      step_id: 'step_child',
      parent_step_id: 'step_parent',
      iteration_key: 'line-1',
      block_type: 'for_each',
    });
    expect(overlay.block_type).toBe('for_each');
    expect(overlay.parent_step_id).toBe('step_parent');
    expect(overlay.iteration_key).toBe('line-1');
  });

  test('ZReviewSummary round-trips seed_version_id and ZReviewQueueResponse wraps it', () => {
    const response = ZReviewQueueResponse.parse({
      data: [QUEUE_ITEM_JSON],
      list_metadata: { before: null, after: REVIEW_ID },
    });
    expect(response.list_metadata.before).toBeNull();
    expect(response.list_metadata.after).toBe(REVIEW_ID);
    expect(response.data[0]?.seed_version_id).toBe(VERSION_ID);
    expect(ZReviewSummary.parse(QUEUE_ITEM_JSON).version_count).toBe(1);
    expect(ZReviewQueueItem.parse(QUEUE_ITEM_JSON).step_id).toBe('step_1');
  });

  test('ZAppendVersionResponse round-trips append status + version id + review', () => {
    for (const status of ['accepted', 'already_exists'] as const) {
      const parsed = ZAppendVersionResponse.parse({
        append_status: status,
        version_id: CHILD_VERSION_ID,
        review: OVERLAY_JSON,
      });
      expect(parsed.append_status).toBe(status);
      expect(parsed.version_id).toBe(CHILD_VERSION_ID);
      expect(parsed.review.versions[CHILD_VERSION_ID]?.parent_id).toBe(VERSION_ID);
    }
  });

  test('ZSubmitDecisionResponse round-trips submission status + review + resume_status', () => {
    for (const status of ['accepted', 'already_applied'] as const) {
      const parsed = ZSubmitDecisionResponse.parse({
        submission_status: status,
        review: OVERLAY_JSON,
        resume_status: 'resumed',
        resume_error: null,
      });
      expect(parsed.submission_status).toBe(status);
      expect(parsed.review.decision?.version_id).toBe(VERSION_ID);
      expect(parsed.resume_status).toBe('resumed');
      expect(parsed.resume_error).toBe(null);
    }
  });

  test('ZSubmitDecisionResponse defaults resume_status to resumed when backend omits it', () => {
    const parsed = ZSubmitDecisionResponse.parse({
      submission_status: 'accepted',
      review: OVERLAY_JSON,
    });
    expect(parsed.resume_status).toBe('resumed');
    expect(parsed.resume_error).toBe(null);
  });

  test('ZSubmitDecisionResponse surfaces resume_status="failed" with an error message', () => {
    const parsed = ZSubmitDecisionResponse.parse({
      submission_status: 'accepted',
      review: OVERLAY_JSON,
      resume_status: 'failed',
      resume_error: 'Workflow run not found for run_id=run_1',
    });
    expect(parsed.resume_status).toBe('failed');
    expect(parsed.resume_error).toContain('Workflow run not found');
  });

  test('ZSubmitDecisionResponse rejects removed accepted_pending_resume status', () => {
    expect(() =>
      ZSubmitDecisionResponse.parse({
        submission_status: 'accepted_pending_resume',
        review: OVERLAY_JSON,
        resume_status: 'failed',
        resume_error: 'Workflow run not found',
      })
    ).toThrow();
  });
});

describe('APIWorkflowReviews request shapes', () => {
  test('list() builds the queue GET with snake_case params', async () => {
    const mock = new MockClient({
      data: [QUEUE_ITEM_JSON],
      list_metadata: { before: null, after: null },
    });
    const reviews = new APIWorkflowReviews(mock);

    await reviews.list({ workflowId: 'wf_1', runId: 'run_1', blockId: 'extract-1', limit: 25 });

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
      data: [QUEUE_ITEM_JSON],
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
  });

  test('get() builds the review GET route', async () => {
    const mock = new MockClient(OVERLAY_JSON);
    const reviews = new APIWorkflowReviews(mock);

    const overlay = await reviews.get(REVIEW_ID);

    expect(mock.lastFetchParams).toEqual({
      url: `/workflows/reviews/${REVIEW_ID}`,
      method: 'GET',
      params: undefined,
      headers: undefined,
    });
    expect(overlay.id).toBe(REVIEW_ID);
  });

  test('approve() posts version_id to the approve endpoint and surfaces resume_status', async () => {
    const mock = new MockClient({
      submission_status: 'accepted',
      review: OVERLAY_JSON,
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
      review: OVERLAY_JSON,
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
      review: OVERLAY_JSON,
      resume_status: 'resumed',
      resume_error: null,
    });
    const reviews = new APIWorkflowReviews(mock);

    await reviews.approve(REVIEW_ID, { versionId: VERSION_ID });

    const body = mock.lastFetchParams?.body as Record<string, unknown>;
    expect('reason' in body).toBe(false);
  });

  test('appendVersion() omits the note key when none is provided', async () => {
    const mock = new MockClient({
      append_status: 'accepted',
      version_id: CHILD_VERSION_ID,
      review: OVERLAY_JSON,
    });
    const reviews = new APIWorkflowReviews(mock);

    await reviews.appendVersion(REVIEW_ID, {
      snapshot: { category: 'Invoice' },
      parentVersionId: VERSION_ID,
    });

    const body = mock.lastFetchParams?.body as Record<string, unknown>;
    expect('note' in body).toBe(false);
    expect('origin' in body).toBe(false);
  });

  test('append_version() strips origin from request option body overrides', async () => {
    const mock = new MockClient({
      append_status: 'accepted',
      version_id: CHILD_VERSION_ID,
      review: OVERLAY_JSON,
    });
    const reviews = new APIWorkflowReviews(mock);

    await reviews.append_version(
      REVIEW_ID,
      {
        snapshot: { category: 'Invoice' },
        parentVersionId: VERSION_ID,
      },
      { body: { origin: 'human_created' } }
    );

    const body = mock.lastFetchParams?.body as Record<string, unknown>;
    expect('origin' in body).toBe(false);
  });

  test('appendVersion() posts a new snapshot version to /versions', async () => {
    const mock = new MockClient({
      append_status: 'accepted',
      version_id: CHILD_VERSION_ID,
      review: OVERLAY_JSON,
    });
    const reviews = new APIWorkflowReviews(mock);

    const result = await reviews.appendVersion(REVIEW_ID, {
      snapshot: { category: 'Invoice' },
      parentVersionId: VERSION_ID,
      note: 'agent proposal',
    });

    expect(mock.lastFetchParams).toEqual({
      url: `/workflows/reviews/${REVIEW_ID}/versions`,
      method: 'POST',
      body: {
        snapshot: { category: 'Invoice' },
        parent_version_id: VERSION_ID,
        note: 'agent proposal',
      },
      params: undefined,
      headers: undefined,
    });
    expect(result.append_status).toBe('accepted');
    expect(result.version_id).toBe(CHILD_VERSION_ID);
  });

  test('waitFor() returns the review once it has no decision', async () => {
    const mock = new MockClient({ ...OVERLAY_JSON, decision: null });
    const reviews = new APIWorkflowReviews(mock);

    const overlay = await reviews.waitFor(REVIEW_ID, { pollIntervalMs: 1 });

    expect(overlay.decision).toBe(null);
    expect(mock.lastFetchParams?.url).toBe(`/workflows/reviews/${REVIEW_ID}`);
  });

  test('wait_for() aliases waitFor() for Python parity', async () => {
    const mock = new MockClient({ ...OVERLAY_JSON, decision: null });
    const reviews = new APIWorkflowReviews(mock);

    const overlay = await reviews.wait_for(REVIEW_ID, { pollIntervalMs: 1 });

    expect(overlay.decision).toBe(null);
  });

  test('waitFor() throws a plain Error on timeout when the review is already decided', async () => {
    const mock = new MockClient(OVERLAY_JSON);
    const reviews = new APIWorkflowReviews(mock);

    await expect(reviews.waitFor(REVIEW_ID, { timeoutMs: 5, pollIntervalMs: 1 })).rejects.toThrow(
      'was not pending'
    );
  });

  test('waitFor() swallows 404 and keeps polling until the review appears', async () => {
    let calls = 0;
    class FlakyClient extends AbstractClient {
      protected async _fetch(): Promise<Response> {
        calls += 1;
        if (calls < 2) {
          return new Response(JSON.stringify({ detail: 'no review' }), {
            status: 404,
            headers: { 'Content-Type': 'application/json' },
          });
        }
        return new Response(JSON.stringify({ ...OVERLAY_JSON, decision: null }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
    }

    const reviews = new APIWorkflowReviews(new FlakyClient());
    const overlay = await reviews.waitFor(REVIEW_ID, { pollIntervalMs: 1 });

    expect(calls).toBe(2);
    expect(overlay.decision).toBe(null);
  });
});
