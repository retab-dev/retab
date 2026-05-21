/**
 * Probe: does the TS SDK's Zod schema match what the live backend ACTUALLY returns?
 *
 * Synthesized from observed responses (verified end-to-end via the Python MCP
 * handler probe against the dev cluster):
 *
 *   - `ReviewOverlayResponse` strips `organization_id` and `runtime_block_id`.
 *   - `SubmitDecisionResponse` uses the field `overlay`, NOT `review`.
 *   - `SubmitDecisionResponse` now carries `resume_status` and `resume_error`.
 *   - `ReviewQueueItemResponse` strips `organization_id` and `runtime_block_id`.
 *   - List response has `data` + `has_more`.
 *
 * If any of these mismatch, every real SDK call will throw a ZodError.
 */

import { describe, expect, test } from 'bun:test';

import {
  ZReviewOverlay,
  ZReviewQueueItem,
  ZReviewQueueResponse,
  ZSubmitDecisionResponse,
} from '../src/types';

const VERSION_ID = 'a'.repeat(64);

const REAL_OVERLAY_FROM_BACKEND = {
  _id: 'br_1',
  workflow_id: 'wf_1',
  workflow_version_id: 'wfv_1',
  workflow_run_id: 'run_1',
  block_id: 'extract-1',
  block_run_id: 'br_1',
  block_type: 'extract',
  triggered_by: { kind: 'any_required_field_null' },
  awaiting_since: '2026-05-21T09:00:00Z',
  priority: 0,
  versions_by_id: {
    [VERSION_ID]: {
      parent_id: null,
      author: { kind: 'model', id: 'm', display_name: 'Model' },
      origin: 'model_output',
      snapshot: { total: 100 },
      note: null,
      created_at: '2026-05-21T09:00:00Z',
    },
  },
  decision: null,
};

const REAL_QUEUE_ITEM_FROM_BACKEND = {
  _id: 'br_1',
  workflow_id: 'wf_1',
  workflow_version_id: 'wfv_1',
  workflow_run_id: 'run_1',
  block_id: 'extract-1',
  block_run_id: 'br_1',
  block_type: 'extract',
  triggered_by: { kind: 'any_required_field_null' },
  awaiting_since: '2026-05-21T09:00:00Z',
  priority: 0,
};

const REAL_SUBMIT_DECISION_RESPONSE = {
  submission_status: 'accepted',
  review: REAL_OVERLAY_FROM_BACKEND,
  resume_status: 'resumed',
  resume_error: null,
};

describe('backend wire-shape vs SDK Zod schemas', () => {
  test('ZReviewOverlay parses a response WITHOUT organization_id (backend strips it)', () => {
    const result = ZReviewOverlay.safeParse(REAL_OVERLAY_FROM_BACKEND);
    if (!result.success) {
      console.error('ZReviewOverlay failed:', JSON.stringify(result.error.issues, null, 2));
    }
    expect(result.success).toBe(true);
  });

  test('ZReviewQueueItem parses a queue row WITHOUT organization_id', () => {
    const result = ZReviewQueueItem.safeParse(REAL_QUEUE_ITEM_FROM_BACKEND);
    if (!result.success) {
      console.error('ZReviewQueueItem failed:', JSON.stringify(result.error.issues, null, 2));
    }
    expect(result.success).toBe(true);
  });

  test('ZReviewQueueResponse parses {data, has_more} with stripped rows', () => {
    const result = ZReviewQueueResponse.safeParse({
      data: [REAL_QUEUE_ITEM_FROM_BACKEND],
      has_more: false,
    });
    if (!result.success) {
      console.error('ZReviewQueueResponse failed:', JSON.stringify(result.error.issues, null, 2));
    }
    expect(result.success).toBe(true);
  });

  test('ZSubmitDecisionResponse parses {submission_status, review, resume_status, resume_error}', () => {
    const result = ZSubmitDecisionResponse.safeParse(REAL_SUBMIT_DECISION_RESPONSE);
    if (!result.success) {
      console.error(
        'ZSubmitDecisionResponse failed:',
        JSON.stringify(result.error.issues, null, 2)
      );
    }
    expect(result.success).toBe(true);
  });
});
