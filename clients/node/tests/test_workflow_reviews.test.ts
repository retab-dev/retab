import { describe, expect, test } from 'bun:test';

import APIWorkflowReviews from '../src/api/workflows/reviews/client';
import { AbstractClient } from '../src/client';
import {
  ZActor,
  ZOutputVersion,
  ZReviewDecision,
  ZReviewOverlay,
  ZReviewQueueItem,
  ZReviewQueueResponse,
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

const VERSION_ID = 'a'.repeat(64);
const CHILD_VERSION_ID = 'b'.repeat(64);

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
  _id: 'br_1',
  workflow_id: 'wf_1',
  workflow_version_id: 'wfv_1',
  workflow_run_id: 'run_1',
  block_id: 'extract-1',
  block_run_id: 'br_1',
  block_type: 'extract',
  triggered_by: { kind: 'confidence_below', threshold: 0.8 },
  awaiting_since: '2026-05-18T09:58:00Z',
  priority: 5,
  versions_by_id: {
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
  _id: 'br_1',
  workflow_id: 'wf_1',
  workflow_version_id: 'wfv_1',
  workflow_run_id: 'run_1',
  block_id: 'extract-1',
  block_run_id: 'br_1',
  block_type: 'extract',
  triggered_by: { kind: 'always' },
  awaiting_since: '2026-05-18T09:58:00Z',
  priority: 5,
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

  test('ZReviewOverlay round-trips versions_by_id and decision', () => {
    const overlay = ZReviewOverlay.parse(OVERLAY_JSON);
    expect(overlay._id).toBe('br_1');
    expect(Object.keys(overlay.versions_by_id)).toEqual([VERSION_ID, CHILD_VERSION_ID]);
    expect(overlay.versions_by_id[CHILD_VERSION_ID]?.parent_id).toBe(VERSION_ID);
    expect(overlay.decision?.version_id).toBe(VERSION_ID);
  });

  test('ZReviewOverlay accepts for_each overlays addressed by the parent block_id', () => {
    // The backend strips runtime_block_id from public responses — reviews are
    // always addressed by the parent block_id, even for_each iterations.
    // Schema MUST tolerate (and silently drop) runtime_block_id if a caller
    // somehow forwards a raw record.
    const overlay = ZReviewOverlay.parse({
      ...OVERLAY_JSON,
      block_id: 'blk_for_each',
      runtime_block_id: 'blk_for_each__map_partition',
      block_type: 'for_each',
    });
    expect(overlay.block_type).toBe('for_each');
    expect(overlay.block_id).toBe('blk_for_each');
    expect('runtime_block_id' in overlay).toBe(false);
  });

  test('ZReviewQueueItem round-trips and ZReviewQueueResponse wraps it', () => {
    const response = ZReviewQueueResponse.parse({
      data: [QUEUE_ITEM_JSON],
      list_metadata: { before: null, after: 'brun_xxx' },
    });
    expect(response.list_metadata.before).toBeNull();
    expect(response.list_metadata.after).toBe('brun_xxx');
    expect(response.data[0]?.block_id).toBe('extract-1');
    expect(ZReviewQueueItem.parse(QUEUE_ITEM_JSON).priority).toBe(5);
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

  test('ZSubmitDecisionResponse accepts submission_status="accepted_pending_resume"', () => {
    const parsed = ZSubmitDecisionResponse.parse({
      submission_status: 'accepted_pending_resume',
      review: OVERLAY_JSON,
      resume_status: 'failed',
      resume_error: 'Workflow run not found',
    });
    expect(parsed.submission_status).toBe('accepted_pending_resume');
    expect(parsed.resume_status).toBe('failed');
  });
});

describe('APIWorkflowReviews request shapes', () => {
  test('list() builds the queue GET with snake_case params', async () => {
    const mock = new MockClient({
      data: [QUEUE_ITEM_JSON],
      list_metadata: { before: null, after: null },
    });
    const reviews = new APIWorkflowReviews(mock);

    await reviews.list({ workflowId: 'wf_1', limit: 25 });

    expect(mock.lastFetchParams).toEqual({
      url: '/workflows/reviews',
      method: 'GET',
      params: { decision: 'none', limit: 25, workflow_id: 'wf_1' },
      headers: undefined,
    });
  });

  test('list() defaults to limit=50 and decision="none"', async () => {
    const mock = new MockClient({
      data: [],
      list_metadata: { before: null, after: null },
    });
    const reviews = new APIWorkflowReviews(mock);

    await reviews.list();

    expect(mock.lastFetchParams?.params).toEqual({ decision: 'none', limit: 50 });
  });

  test('list({ decision: "any" }) includes decided reviews', async () => {
    const mock = new MockClient({
      data: [],
      list_metadata: { before: null, after: null },
    });
    const reviews = new APIWorkflowReviews(mock);

    await reviews.list({ decision: 'any' });

    expect(mock.lastFetchParams?.params).toEqual({ decision: 'any', limit: 50 });
  });

  test('list({ after }) forwards cursor and surfaces list_metadata', async () => {
    const mock = new MockClient({
      data: [QUEUE_ITEM_JSON],
      list_metadata: { before: null, after: 'brun_xxx' },
    });
    const reviews = new APIWorkflowReviews(mock);

    const result = await reviews.list({ after: 'brun_prev', limit: 25 });

    expect(mock.lastFetchParams?.params).toEqual({
      decision: 'none',
      limit: 25,
      after: 'brun_prev',
    });
    expect(result.list_metadata.after).toBe('brun_xxx');
    expect(result.list_metadata.before).toBeNull();
  });

  test('get() builds the overlay GET route', async () => {
    const mock = new MockClient(OVERLAY_JSON);
    const reviews = new APIWorkflowReviews(mock);

    const overlay = await reviews.get('run_1', 'extract-1');

    expect(mock.lastFetchParams).toEqual({
      url: '/workflows/reviews/run_1/extract-1',
      method: 'GET',
      params: undefined,
      headers: undefined,
    });
    expect(overlay._id).toBe('br_1');
  });

  test('approve() posts verdict=approved with version_id and surfaces resume_status', async () => {
    const mock = new MockClient({
      submission_status: 'accepted',
      review: OVERLAY_JSON,
      resume_status: 'resumed',
      resume_error: null,
    });
    const reviews = new APIWorkflowReviews(mock);

    const result = await reviews.approve('run_1', 'extract-1', { versionId: VERSION_ID });

    expect(mock.lastFetchParams).toEqual({
      url: '/workflows/reviews/run_1/extract-1/decision',
      method: 'POST',
      body: {
        verdict: 'approved',
        version_id: VERSION_ID,
      },
      params: undefined,
      headers: undefined,
    });
    expect(result.submission_status).toBe('accepted');
    expect(result.resume_status).toBe('resumed');
    expect(result.review.decision?.version_id).toBe(VERSION_ID);
  });

  test('reject() posts verdict=rejected with version_id and reason', async () => {
    const mock = new MockClient({
      submission_status: 'accepted',
      review: OVERLAY_JSON,
      resume_status: 'resumed',
      resume_error: null,
    });
    const reviews = new APIWorkflowReviews(mock);

    await reviews.reject('run_1', 'extract-1', { versionId: VERSION_ID, reason: 'wrong total' });

    expect(mock.lastFetchParams?.body).toEqual({
      verdict: 'rejected',
      version_id: VERSION_ID,
      reason: 'wrong total',
    });
  });

  test('approve() omits the reason key when none is provided', async () => {
    const mock = new MockClient({
      submission_status: 'accepted',
      review: OVERLAY_JSON,
      resume_status: 'resumed',
      resume_error: null,
    });
    const reviews = new APIWorkflowReviews(mock);

    await reviews.approve('run_1', 'extract-1', { versionId: VERSION_ID });

    const body = mock.lastFetchParams?.body as Record<string, unknown>;
    expect('reason' in body).toBe(false);
  });

  test('createVersion() omits the note key when none is provided', async () => {
    const mock = new MockClient(OVERLAY_JSON);
    const reviews = new APIWorkflowReviews(mock);

    await reviews.createVersion('run_1', 'extract-1', {
      snapshot: { category: 'Invoice' },
      parentId: VERSION_ID,
    });

    const body = mock.lastFetchParams?.body as Record<string, unknown>;
    expect('note' in body).toBe(false);
    expect('origin' in body).toBe(false);
  });

  test('createVersion() strips origin from request option body overrides', async () => {
    const mock = new MockClient(OVERLAY_JSON);
    const reviews = new APIWorkflowReviews(mock);

    await reviews.createVersion(
      'run_1',
      'extract-1',
      {
        snapshot: { category: 'Invoice' },
        parentId: VERSION_ID,
      },
      { body: { origin: 'human_created' } }
    );

    const body = mock.lastFetchParams?.body as Record<string, unknown>;
    expect('origin' in body).toBe(false);
  });

  test('createVersion() posts a new snapshot version to /versions', async () => {
    const mock = new MockClient(OVERLAY_JSON);
    const reviews = new APIWorkflowReviews(mock);

    await reviews.createVersion('run_1', 'extract-1', {
      snapshot: { category: 'Invoice' },
      parentId: VERSION_ID,
      note: 'agent proposal',
    });

    expect(mock.lastFetchParams).toEqual({
      url: '/workflows/reviews/run_1/extract-1/versions',
      method: 'POST',
      body: {
        snapshot: { category: 'Invoice' },
        parent_id: VERSION_ID,
        note: 'agent proposal',
      },
      params: undefined,
      headers: undefined,
    });
  });

  test('waitFor() returns the review once it has no decision', async () => {
    const mock = new MockClient({ ...OVERLAY_JSON, decision: null });
    const reviews = new APIWorkflowReviews(mock);

    const overlay = await reviews.waitFor('run_1', 'extract-1', { pollIntervalMs: 1 });

    expect(overlay.decision).toBe(null);
    expect(mock.lastFetchParams?.url).toBe('/workflows/reviews/run_1/extract-1');
  });

  test('wait_for() aliases waitFor() for Python parity', async () => {
    const mock = new MockClient({ ...OVERLAY_JSON, decision: null });
    const reviews = new APIWorkflowReviews(mock);

    const overlay = await reviews.wait_for('run_1', 'extract-1', { pollIntervalMs: 1 });

    expect(overlay.decision).toBe(null);
  });

  test('waitFor() throws a plain Error on timeout when the review is already decided', async () => {
    const mock = new MockClient(OVERLAY_JSON);
    const reviews = new APIWorkflowReviews(mock);

    await expect(
      reviews.waitFor('run_1', 'extract-1', { timeoutMs: 5, pollIntervalMs: 1 })
    ).rejects.toThrow('was not awaiting_review');
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
    const overlay = await reviews.waitFor('run_1', 'extract-1', { pollIntervalMs: 1 });

    expect(calls).toBe(2);
    expect(overlay.decision).toBe(null);
  });
});
