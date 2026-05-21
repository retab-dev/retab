/**
 * Probe: does the TS SDK's Zod schema match the review hard-cutover wire shape?
 */

import { describe, expect, test } from 'bun:test';

import {
  ZReview,
  ZReviewQueueResponse,
  ZReviewSummary,
  ZReviewVersion,
  ZReviewVersionListResponse,
  ZSubmitDecisionResponse,
} from '../src/types';

const REVIEW_ID = 'rev_01HX9A7Y1R6G9J2K8M4P5Q6T7V';
const VERSION_ID = 'rvr_2N9S8Q4F6M1K7C3D5R0T8W6Y2Z';

const REAL_REVIEW_FROM_BACKEND = {
  id: REVIEW_ID,
  workflow_id: 'wf_1',
  workflow_version_id: 'wfv_1',
  workflow_run_id: 'run_1',
  block_id: 'extract-1',
  step_id: 'step_1',
  parent_step_id: null,
  iteration_key: null,
  block_type: 'extract',
  triggered_by: { kind: 'any_required_field_null' },
  created_at: '2026-05-21T09:00:00Z',
  decision: null,
};

const REAL_SUMMARY_FROM_BACKEND = {
  ...REAL_REVIEW_FROM_BACKEND,
  seed_version_id: VERSION_ID,
  version_count: 1,
};

const REAL_VERSION_FROM_BACKEND = {
  id: VERSION_ID,
  review_id: REVIEW_ID,
  parent_id: null,
  author: { kind: 'model', id: 'm', display_name: 'Model' },
  snapshot: { total: 100 },
  note: null,
  created_at: '2026-05-21T09:00:00Z',
};

describe('backend wire-shape vs SDK Zod schemas', () => {
  test('ZReview parses the review-id addressed response without versions', () => {
    const result = ZReview.safeParse(REAL_REVIEW_FROM_BACKEND);
    if (!result.success) {
      console.error('ZReview failed:', JSON.stringify(result.error.issues, null, 2));
    }
    expect(result.success).toBe(true);
  });

  test('ZReviewVersion parses a flat review version', () => {
    const result = ZReviewVersion.safeParse(REAL_VERSION_FROM_BACKEND);
    if (!result.success) {
      console.error('ZReviewVersion failed:', JSON.stringify(result.error.issues, null, 2));
    }
    expect(result.success).toBe(true);
  });

  test('ZReviewSummary parses queue projection fields', () => {
    const result = ZReviewSummary.safeParse(REAL_SUMMARY_FROM_BACKEND);
    if (!result.success) {
      console.error('ZReviewSummary failed:', JSON.stringify(result.error.issues, null, 2));
    }
    expect(result.success).toBe(true);
  });

  test('ZReviewQueueResponse parses {data, list_metadata}', () => {
    const result = ZReviewQueueResponse.safeParse({
      data: [REAL_SUMMARY_FROM_BACKEND],
      list_metadata: { before: null, after: null },
    });
    if (!result.success) {
      console.error('ZReviewQueueResponse failed:', JSON.stringify(result.error.issues, null, 2));
    }
    expect(result.success).toBe(true);
  });

  test('ZReviewVersionListResponse parses {data, list_metadata}', () => {
    const result = ZReviewVersionListResponse.safeParse({
      data: [REAL_VERSION_FROM_BACKEND],
      list_metadata: { before: null, after: null },
    });
    if (!result.success) {
      console.error(
        'ZReviewVersionListResponse failed:',
        JSON.stringify(result.error.issues, null, 2)
      );
    }
    expect(result.success).toBe(true);
  });

  test('ZSubmitDecisionResponse parses {submission_status, review, resume_status, resume_error}', () => {
    const result = ZSubmitDecisionResponse.safeParse({
      submission_status: 'accepted',
      review: REAL_REVIEW_FROM_BACKEND,
      resume_status: 'resumed',
      resume_error: null,
    });
    if (!result.success) {
      console.error(
        'ZSubmitDecisionResponse failed:',
        JSON.stringify(result.error.issues, null, 2)
      );
    }
    expect(result.success).toBe(true);
  });
});
