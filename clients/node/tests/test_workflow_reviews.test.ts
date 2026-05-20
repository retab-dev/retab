import { describe, expect, test } from 'bun:test';

import { AbstractClient } from '../src/client';
import APIWorkflowReviews from '../src/api/workflows/reviews/client';
import {
  ZActor,
  ZAuditEntry,
  ZOutputVersion,
  ZReviewClaim,
  ZReviewDecision,
  ZReviewOverlay,
  ZReviewQueueItem,
  ZReviewQueueResponse,
  ZSubmitDecisionResponse,
} from '../src/types';

// ---------------------------------------------------------------------------
// Mock client helpers
// ---------------------------------------------------------------------------

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

const ACTOR_JSON = {
  kind: 'agent',
  id: 'agent_7',
  display_name: 'Reviewer Agent',
};

const OUTPUT_VERSION_JSON = {
  seq: 1,
  parent_seq: 0,
  author: ACTOR_JSON,
  origin: 'agent_edit',
  snapshot: { invoice_number: 'INV-001' },
  content_sha256: 'abc123',
  note: 'tweaked the total',
  created_at: '2026-05-18T10:00:00Z',
};

const DECISION_JSON = {
  decision_id: 'dec_1',
  verdict: 'approved',
  decided_by: ACTOR_JSON,
  decided_at: '2026-05-18T10:05:00Z',
  on_seq: 1,
  effective_seq: 1,
  reason: null,
  supersedes_decision_id: null,
};

const AUDIT_JSON = {
  entry_id: 'aud_1',
  action: 'version.created',
  actor: ACTOR_JSON,
  at: '2026-05-18T10:00:00Z',
  rev_observed: 2,
  rev_result: 3,
  detail: { seq: 1 },
};

const CLAIM_JSON = {
  holder: ACTOR_JSON,
  claimed_at: '2026-05-18T09:59:00Z',
  expires_at: '2026-05-18T10:14:00Z',
};

const OVERLAY_JSON = {
  _id: 'ov_1',
  organization_id: 'org_1',
  workflow_id: 'wf_1',
  workflow_version_id: 'wfv_1',
  workflow_run_id: 'run_1',
  block_id: 'extract-1',
  block_run_id: 'br_1',
  block_type: 'extract',
  triggered_by: { kind: 'confidence_below', threshold: 0.8 },
  status: 'awaiting_review',
  awaiting_since: '2026-05-18T09:58:00Z',
  decided_at: null,
  priority: 5,
  rev: 3,
  claim: CLAIM_JSON,
  versions: [OUTPUT_VERSION_JSON],
  decisions: [DECISION_JSON],
  audit: [AUDIT_JSON],
  head_seq: 1,
  effective_seq: null,
};

const QUEUE_ITEM_JSON = {
  _id: 'ov_1',
  organization_id: 'org_1',
  workflow_id: 'wf_1',
  workflow_version_id: 'wfv_1',
  workflow_run_id: 'run_1',
  block_id: 'extract-1',
  block_run_id: 'br_1',
  block_type: 'extract',
  triggered_by: { kind: 'always' },
  status: 'awaiting_review',
  awaiting_since: '2026-05-18T09:58:00Z',
  decided_at: null,
  priority: 5,
  rev: 3,
  claim: null,
  head_seq: 1,
  effective_seq: null,
};

// ---------------------------------------------------------------------------
// Zod schema round-trips
// ---------------------------------------------------------------------------

describe('review overlay zod schemas', () => {
  test('ZActor parses model / agent / human kinds identically', () => {
    for (const kind of ['model', 'agent', 'human'] as const) {
      const actor = ZActor.parse({ ...ACTOR_JSON, kind });
      expect(actor.kind).toBe(kind);
    }
  });

  test('ZOutputVersion round-trips a full version', () => {
    const version = ZOutputVersion.parse(OUTPUT_VERSION_JSON);
    expect(version.seq).toBe(1);
    expect(version.parent_seq).toBe(0);
    expect(version.origin).toBe('agent_edit');
    expect(version.snapshot).toEqual({ invoice_number: 'INV-001' });
    expect(version.content_sha256).toBe('abc123');
  });

  test('ZReviewDecision round-trips and accepts every verdict', () => {
    for (const verdict of ['approved', 'rejected'] as const) {
      const decision = ZReviewDecision.parse({ ...DECISION_JSON, verdict });
      expect(decision.verdict).toBe(verdict);
    }
  });

  test('ZAuditEntry round-trips an audit row', () => {
    const entry = ZAuditEntry.parse(AUDIT_JSON);
    expect(entry.rev_observed).toBe(2);
    expect(entry.rev_result).toBe(3);
    expect(entry.detail).toEqual({ seq: 1 });
  });

  test('ZReviewClaim round-trips a soft lock', () => {
    const claim = ZReviewClaim.parse(CLAIM_JSON);
    expect(claim.holder.id).toBe('agent_7');
    expect(claim.expires_at).toBe('2026-05-18T10:14:00Z');
  });

  test('ZReviewOverlay round-trips the full overlay with nested arrays', () => {
    const overlay = ZReviewOverlay.parse(OVERLAY_JSON);
    expect(overlay._id).toBe('ov_1');
    expect(overlay.rev).toBe(3);
    expect(overlay.versions).toHaveLength(1);
    expect(overlay.decisions).toHaveLength(1);
    expect(overlay.audit).toHaveLength(1);
    expect(overlay.claim?.holder.kind).toBe('agent');
  });

  test('ZReviewOverlay accepts for_each overlays without runtime projection identity', () => {
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
      has_more: true,
    });
    expect(response.has_more).toBe(true);
    expect(response.data).toHaveLength(1);
    expect(response.data[0]?.block_id).toBe('extract-1');
    expect(ZReviewQueueItem.parse(QUEUE_ITEM_JSON).rev).toBe(3);
  });

  test('ZReviewQueueItem accepts for_each items without runtime projection identity', () => {
    const item = ZReviewQueueItem.parse({
      ...QUEUE_ITEM_JSON,
      block_id: 'blk_for_each',
      runtime_block_id: 'blk_for_each__map_partition',
      block_type: 'for_each',
    });
    expect(item.block_type).toBe('for_each');
    expect(item.block_id).toBe('blk_for_each');
    expect('runtime_block_id' in item).toBe(false);
  });

  test('ZSubmitDecisionResponse round-trips submission status + overlay', () => {
    for (const status of ['accepted', 'already_received', 'already_applied'] as const) {
      const parsed = ZSubmitDecisionResponse.parse({
        submission_status: status,
        overlay: OVERLAY_JSON,
      });
      expect(parsed.submission_status).toBe(status);
      expect(parsed.overlay._id).toBe('ov_1');
    }
  });
});

// ---------------------------------------------------------------------------
// Request shapes the methods build
// ---------------------------------------------------------------------------

describe('APIWorkflowReviews request shapes', () => {
  test('list() builds the queue GET with snake_case params', async () => {
    const mock = new MockClient({ data: [QUEUE_ITEM_JSON], has_more: false });
    const reviews = new APIWorkflowReviews(mock);

    const result = await reviews.list({ workflowId: 'wf_1', status: 'approved', mine: true, limit: 25 });

    expect(mock.lastFetchParams).toEqual({
      url: '/workflows/reviews',
      method: 'GET',
      params: { status: 'approved', mine: true, limit: 25, workflow_id: 'wf_1' },
      headers: undefined,
    });
    expect(result.data[0]?.block_id).toBe('extract-1');
  });

  test('list() defaults to awaiting_review / mine=false / limit=50', async () => {
    const mock = new MockClient({ data: [], has_more: false });
    const reviews = new APIWorkflowReviews(mock);

    await reviews.list();

    expect(mock.lastFetchParams?.params).toEqual({
      status: 'awaiting_review',
      mine: false,
      limit: 50,
    });
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
    expect(overlay._id).toBe('ov_1');
  });

  test('approve() posts verdict=approved to /decision with full body', async () => {
    const mock = new MockClient({ submission_status: 'accepted', overlay: OVERLAY_JSON });
    const reviews = new APIWorkflowReviews(mock);

    const result = await reviews.approve('run_1', 'extract-1', {
      versionStamp: 3,
      editedOutput: { invoice_number: 'INV-002' },
      onSeq: 1,
      effectiveSeq: 2,
      commandId: 'cmd_1',
    });

    expect(mock.lastFetchParams).toEqual({
      url: '/workflows/reviews/run_1/extract-1/decision',
      method: 'POST',
      body: {
        verdict: 'approved',
        version_stamp: 3,
        edited_output: { invoice_number: 'INV-002' },
        reviewable_value: null,
        on_seq: 1,
        effective_seq: 2,
        reason: null,
        command_id: 'cmd_1',
      },
      params: undefined,
      headers: undefined,
    });
    expect(result.submission_status).toBe('accepted');
  });

  test('approve() can post reviewable_value instead of edited_output', async () => {
    const mock = new MockClient({ submission_status: 'accepted', overlay: OVERLAY_JSON });
    const reviews = new APIWorkflowReviews(mock);

    await reviews.approve('run_1', 'extract-1', {
      versionStamp: 3,
      reviewableValue: { chunks: [{ key: 'legal-mentions', pages: [1] }] },
    });

    expect(mock.lastFetchParams?.body).toEqual({
      verdict: 'approved',
      version_stamp: 3,
      edited_output: null,
      reviewable_value: { chunks: [{ key: 'legal-mentions', pages: [1] }] },
      on_seq: null,
      effective_seq: null,
      reason: null,
      command_id: null,
    });
  });

  test('reject() posts verdict=rejected with the required reason', async () => {
    const mock = new MockClient({ submission_status: 'accepted', overlay: OVERLAY_JSON });
    const reviews = new APIWorkflowReviews(mock);

    await reviews.reject('run_1', 'extract-1', { versionStamp: 3, reason: 'wrong total' });

    expect(mock.lastFetchParams?.body).toEqual({
      verdict: 'rejected',
      version_stamp: 3,
      edited_output: null,
      reviewable_value: null,
      on_seq: null,
      effective_seq: null,
      reason: 'wrong total',
      command_id: null,
    });
  });

  test('edit() posts a new version to /versions, origin defaults to human_edit', async () => {
    const mock = new MockClient(OVERLAY_JSON);
    const reviews = new APIWorkflowReviews(mock);

    await reviews.edit('run_1', 'extract-1', {
      snapshot: { invoice_number: 'INV-003' },
      versionStamp: 3,
    });

    expect(mock.lastFetchParams).toEqual({
      url: '/workflows/reviews/run_1/extract-1/versions',
      method: 'POST',
      body: {
        snapshot: { invoice_number: 'INV-003' },
        reviewable_value: null,
        version_stamp: 3,
        origin: 'human_edit',
        note: null,
        command_id: null,
      },
      params: undefined,
      headers: undefined,
    });
  });

  test('edit() carries an agent_edit origin through the same call', async () => {
    const mock = new MockClient(OVERLAY_JSON);
    const reviews = new APIWorkflowReviews(mock);

    await reviews.edit('run_1', 'extract-1', {
      snapshot: { invoice_number: 'INV-004' },
      versionStamp: 3,
      origin: 'agent_edit',
      note: 'agent proposal',
      commandId: 'cmd_9',
    });

    expect(mock.lastFetchParams?.body).toEqual({
      snapshot: { invoice_number: 'INV-004' },
      reviewable_value: null,
      version_stamp: 3,
      origin: 'agent_edit',
      note: 'agent proposal',
      command_id: 'cmd_9',
    });
  });

  test('claim() posts version_stamp + ttl_seconds, default ttl 900', async () => {
    const mock = new MockClient(OVERLAY_JSON);
    const reviews = new APIWorkflowReviews(mock);

    await reviews.claim('run_1', 'extract-1', { versionStamp: 3 });

    expect(mock.lastFetchParams).toEqual({
      url: '/workflows/reviews/run_1/extract-1/claim',
      method: 'POST',
      body: { version_stamp: 3, ttl_seconds: 900 },
      params: undefined,
      headers: undefined,
    });
  });

  test('prepare_claim() exposes the Python-style request builder', () => {
    const reviews = new APIWorkflowReviews(new MockClient(OVERLAY_JSON));

    expect(reviews.prepare_claim('run_1', 'extract-1', { versionStamp: 3, ttlSeconds: 600 })).toEqual({
      url: '/workflows/reviews/run_1/extract-1/claim',
      method: 'POST',
      body: { version_stamp: 3, ttl_seconds: 600 },
    });
  });

  test('release() posts version_stamp to /release', async () => {
    const mock = new MockClient(OVERLAY_JSON);
    const reviews = new APIWorkflowReviews(mock);

    await reviews.release('run_1', 'extract-1', { versionStamp: 3 });

    expect(mock.lastFetchParams).toEqual({
      url: '/workflows/reviews/run_1/extract-1/release',
      method: 'POST',
      body: { version_stamp: 3 },
      params: undefined,
      headers: undefined,
    });
  });

  test('waitFor() returns the overlay once it is awaiting_review', async () => {
    const mock = new MockClient(OVERLAY_JSON);
    const reviews = new APIWorkflowReviews(mock);

    const overlay = await reviews.waitFor('run_1', 'extract-1', { pollIntervalMs: 1 });

    expect(overlay.status).toBe('awaiting_review');
    expect(mock.lastFetchParams?.url).toBe('/workflows/reviews/run_1/extract-1');
  });

  test('wait_for() aliases waitFor() for Python parity', async () => {
    const mock = new MockClient(OVERLAY_JSON);
    const reviews = new APIWorkflowReviews(mock);

    const overlay = await reviews.wait_for('run_1', 'extract-1', { pollIntervalMs: 1 });

    expect(overlay.status).toBe('awaiting_review');
    expect(mock.lastFetchParams?.url).toBe('/workflows/reviews/run_1/extract-1');
  });

  test('waitFor() throws a plain Error on timeout when the overlay never gates', async () => {
    const mock = new MockClient({ ...OVERLAY_JSON, status: 'approved' });
    const reviews = new APIWorkflowReviews(mock);

    await expect(
      reviews.waitFor('run_1', 'extract-1', { timeoutMs: 5, pollIntervalMs: 1 })
    ).rejects.toThrow('was not awaiting_review');
  });

  test('waitFor() swallows 404 and keeps polling until the overlay appears', async () => {
    let calls = 0;
    class FlakyClient extends AbstractClient {
      protected async _fetch(): Promise<Response> {
        calls += 1;
        if (calls < 2) {
          return new Response(JSON.stringify({ detail: 'no overlay' }), {
            status: 404,
            headers: { 'Content-Type': 'application/json' },
          });
        }
        return new Response(JSON.stringify(OVERLAY_JSON), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }
    }

    const reviews = new APIWorkflowReviews(new FlakyClient());
    const overlay = await reviews.waitFor('run_1', 'extract-1', { pollIntervalMs: 1 });

    expect(calls).toBe(2);
    expect(overlay.status).toBe('awaiting_review');
  });
});
