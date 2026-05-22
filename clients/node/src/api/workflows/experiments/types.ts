/**
 * Zod schemas + TS types for the workflow experiments API.
 *
 * Mirrors `backend/main_server/main_server/services/v1/workflows/experiments/{models,routes}.py`
 * and the Python SDK's `retab/types/workflows/experiments.py`. Wire-format
 * field names are snake_case.
 *
 * The metrics response is a deeply nested discriminated union (four views
 * plus two error envelopes). Modeling the full union here would couple the
 * SDK to internal structural details that are still evolving. We expose:
 *
 *   - typed wrappers for the simple/stable shapes (Experiment, runs,
 *     eligible blocks, experiment run results)
 *   - `ExperimentMetricsResponse = Record<string, unknown>` — caller
 *     branches on the shared `kind` field.
 */

import * as z from 'zod';
import { ZErrorDetails, ZWorkflowSnapshotRef } from '../../../generated_types.js';

// ---------------------------------------------------------------------------
// Enums
// ---------------------------------------------------------------------------

export const ZNConsensusValue = z.union([z.literal(3), z.literal(5), z.literal(7)]);
export type NConsensusValue = z.infer<typeof ZNConsensusValue>;

export const ZExperimentBlockType = z.enum(['extract', 'classifier', 'split', 'for_each']);
export type ExperimentBlockType = z.infer<typeof ZExperimentBlockType>;

export const ZExperimentRunStatus = z.enum([
  'pending',
  'running',
  'completed',
  'error',
  'cancelled',
]);
export type ExperimentRunStatus = z.infer<typeof ZExperimentRunStatus>;

export const ZExperimentPublicStatus = z.enum([
  'draft',
  'processing',
  'completed',
  'failed',
  'cancelled',
]);
export type ExperimentPublicStatus = z.infer<typeof ZExperimentPublicStatus>;

export const ZSchemaDriftStatus = z.enum(['unknown', 'none', 'partial', 'drifted']);
export type SchemaDriftStatus = z.infer<typeof ZSchemaDriftStatus>;

export const ZExperimentMetricView = z.enum(['summary', 'by_document', 'by_target', 'votes']);
export type ExperimentMetricView = z.infer<typeof ZExperimentMetricView>;

// ---------------------------------------------------------------------------
// Request shapes (SDK callers pass these as object literals)
// ---------------------------------------------------------------------------

export type ExperimentDocumentProvenance = {
  workflow_run_id?: string | null;
  step_id?: string | null;
};

export type ExperimentDocumentCaptureRequest = {
  workflow_run_id: string;
  step_id?: string | null;
};

export type ExplicitExperimentDocumentRequest = {
  handle_inputs: Record<string, unknown>;
  provenance?: ExperimentDocumentProvenance | null;
};

// ---------------------------------------------------------------------------
// Experiment (response)
// ---------------------------------------------------------------------------

export const ZWorkflowExperiment = z
  .object({
    id: z.string(),
    workflow_id: z.string(),
    block_id: z.string(),
    n_consensus: ZNConsensusValue,
    document_count: z.number().default(0),
    name: z.string(),
    last_run_id: z.string().nullable().optional(),
    created_at: z.string(),
    updated_at: z.string(),
    status: ZExperimentPublicStatus.default('draft'),
    block_type: ZExperimentBlockType,
    score: z.number().nullable().optional(),
    is_stale: z.boolean().default(false),
    schema_drift: ZSchemaDriftStatus.default('unknown'),
    schema_drift_detail: z.string().nullable().optional(),
  })
  .passthrough();
export type WorkflowExperiment = z.infer<typeof ZWorkflowExperiment>;

// Canonical PaginatedList envelope. The route used to return a bare array;
// the migration to `{data, list_metadata}` brings it in line with workflows
// list, files list, extractions list, etc.
export const ZListMetadata = z
  .object({
    before: z.string().nullable(),
    after: z.string().nullable(),
  })
  .passthrough();
export type ListMetadata = z.infer<typeof ZListMetadata>;

export const ZExperimentList = z
  .object({
    data: z.array(ZWorkflowExperiment).default([]),
    list_metadata: ZListMetadata,
  })
  .passthrough();
export type ExperimentList = z.infer<typeof ZExperimentList>;

// ---------------------------------------------------------------------------
// Run (response)
// ---------------------------------------------------------------------------

export const ZExperimentRunTrigger = z
  .object({
    type: z.string().nullable().optional(),
  })
  .passthrough();
export type ExperimentRunTrigger = z.infer<typeof ZExperimentRunTrigger>;

// --- ExperimentRun lifecycle discriminated union ---------------------------
//
// Mirror of the backend wire shape. Per the meta-pattern blueprint
// principle #3 (invariants in shape, not prose), ``lifecycle`` is a
// discriminated union so per-state fields can carry data
// (``error.message`` / ``cancelled.reason``).

export const ZPendingWorkflowExperimentRun = z
  .object({ status: z.literal('pending').default('pending') })
  .strip();
export type PendingWorkflowExperimentRun = z.infer<typeof ZPendingWorkflowExperimentRun>;

export const ZQueuedWorkflowExperimentRun = z
  .object({ status: z.literal('queued').default('queued') })
  .strip();
export type QueuedWorkflowExperimentRun = z.infer<typeof ZQueuedWorkflowExperimentRun>;

export const ZRunningWorkflowExperimentRun = z
  .object({ status: z.literal('running').default('running') })
  .strip();
export type RunningWorkflowExperimentRun = z.infer<typeof ZRunningWorkflowExperimentRun>;

export const ZCompletedWorkflowExperimentRun = z
  .object({ status: z.literal('completed').default('completed') })
  .strip();
export type CompletedWorkflowExperimentRun = z.infer<typeof ZCompletedWorkflowExperimentRun>;

export const ZErrorWorkflowExperimentRun = z
  .object({
    status: z.literal('error').default('error'),
    message: z.string().default('(no message)'),
    details: ZErrorDetails.nullable().optional(),
  })
  .strip();
export type ErrorWorkflowExperimentRun = z.infer<typeof ZErrorWorkflowExperimentRun>;

export const ZCancelledWorkflowExperimentRun = z
  .object({
    status: z.literal('cancelled').default('cancelled'),
    reason: z.string().nullable().optional(),
  })
  .strip();
export type CancelledWorkflowExperimentRun = z.infer<typeof ZCancelledWorkflowExperimentRun>;

export const ZExperimentRunLifecycle = z.discriminatedUnion('status', [
  ZPendingWorkflowExperimentRun,
  ZQueuedWorkflowExperimentRun,
  ZRunningWorkflowExperimentRun,
  ZCompletedWorkflowExperimentRun,
  ZErrorWorkflowExperimentRun,
  ZCancelledWorkflowExperimentRun,
]);
export type ExperimentRunLifecycle = z.infer<typeof ZExperimentRunLifecycle>;

export const ZExperimentRunTiming = z
  .object({
    created_at: z.string(),
    started_at: z.string().nullable().optional(),
    completed_at: z.string().nullable().optional(),
    duration_ms: z.number().nullable().optional(),
  })
  .passthrough();
export type ExperimentRunTiming = z.infer<typeof ZExperimentRunTiming>;

export const ZExperimentRun = z
  .object({
    id: z.string(),
    workflow: ZWorkflowSnapshotRef,
    trigger: ZExperimentRunTrigger,
    lifecycle: ZExperimentRunLifecycle,
    timing: ZExperimentRunTiming,
    experiment_id: z.string(),
    block_id: z.string(),
    block_type: ZExperimentBlockType,
    n_consensus: ZNConsensusValue,
    definition_fingerprint: z.string(),
    documents_fingerprint: z.string(),
    score: z.number().nullable().optional(),
    total_document_count: z.number().default(0),
    completed_document_count: z.number().default(0),
    document_count: z.number().default(0),
    error_count: z.number().default(0),
  })
  .strip();
export type ExperimentRun = z.infer<typeof ZExperimentRun>;

export const ZCancelWorkflowExperimentRunResponse = z
  .object({
    id: z.string(),
    lifecycle: ZExperimentRunLifecycle,
  })
  .strip();
export type CancelWorkflowExperimentRunResponse = z.infer<
  typeof ZCancelWorkflowExperimentRunResponse
>;

// Canonical PaginatedList envelope for `GET /workflows/experiments/runs`.
// The route used to return `{"runs": [...]}`; the migration to
// `{data, list_metadata}` matches the rest of the Retab list endpoints.
export const ZExperimentRunListResponse = z
  .object({
    data: z.array(ZExperimentRun).default([]),
    list_metadata: ZListMetadata,
  })
  .passthrough();
export type ExperimentRunListResponse = z.infer<typeof ZExperimentRunListResponse>;

// ---------------------------------------------------------------------------
// Per-document results
// ---------------------------------------------------------------------------

export const ZExperimentResultStatus = z.enum(['pending', 'running', 'completed', 'error']);
export type ExperimentResultStatus = z.infer<typeof ZExperimentResultStatus>;

// --- ExperimentResult lifecycle discriminated union ------------------------

export const ZPendingWorkflowExperimentResult = z
  .object({ status: z.literal('pending').default('pending') })
  .strip();
export type PendingWorkflowExperimentResult = z.infer<typeof ZPendingWorkflowExperimentResult>;

export const ZQueuedWorkflowExperimentResult = z
  .object({ status: z.literal('queued').default('queued') })
  .strip();
export type QueuedWorkflowExperimentResult = z.infer<typeof ZQueuedWorkflowExperimentResult>;

export const ZRunningWorkflowExperimentResult = z
  .object({ status: z.literal('running').default('running') })
  .strip();
export type RunningWorkflowExperimentResult = z.infer<typeof ZRunningWorkflowExperimentResult>;

export const ZCompletedWorkflowExperimentResult = z
  .object({ status: z.literal('completed').default('completed') })
  .strip();
export type CompletedWorkflowExperimentResult = z.infer<typeof ZCompletedWorkflowExperimentResult>;

export const ZErrorWorkflowExperimentResult = z
  .object({
    status: z.literal('error').default('error'),
    message: z.string().default('(no message)'),
    details: ZErrorDetails.nullable().optional(),
  })
  .strip();
export type ErrorWorkflowExperimentResult = z.infer<typeof ZErrorWorkflowExperimentResult>;

export const ZCancelledWorkflowExperimentResult = z
  .object({
    status: z.literal('cancelled').default('cancelled'),
    reason: z.string().nullable().optional(),
  })
  .strip();
export type CancelledWorkflowExperimentResult = z.infer<typeof ZCancelledWorkflowExperimentResult>;

export const ZExperimentResultLifecycle = z.discriminatedUnion('status', [
  ZPendingWorkflowExperimentResult,
  ZQueuedWorkflowExperimentResult,
  ZRunningWorkflowExperimentResult,
  ZCompletedWorkflowExperimentResult,
  ZErrorWorkflowExperimentResult,
  ZCancelledWorkflowExperimentResult,
]);
export type ExperimentResultLifecycle = z.infer<typeof ZExperimentResultLifecycle>;

export const ZExperimentResultTiming = z
  .object({
    created_at: z.string().nullable().optional(),
    started_at: z.string().nullable().optional(),
    completed_at: z.string().nullable().optional(),
    duration_ms: z.number().nullable().optional(),
  })
  .passthrough();
export type ExperimentResultTiming = z.infer<typeof ZExperimentResultTiming>;

export const ZStepArtifactRefMini = z
  .object({
    operation: z.string(),
    id: z.string(),
  })
  .passthrough();
export type StepArtifactRefMini = z.infer<typeof ZStepArtifactRefMini>;

export const ZExperimentResult = z
  .object({
    id: z.string(),
    run_id: z.string(),
    experiment_id: z.string(),
    document_id: z.string(),
    lifecycle: ZExperimentResultLifecycle,
    timing: ZExperimentResultTiming,
    block_type: ZExperimentBlockType,
    handle_inputs: z.record(z.any()).default({}),
    artifact: ZStepArtifactRefMini.nullable().optional(),
    duration_ms: z.number().nullable().optional(),
    created_at: z.string().nullable().optional(),
    started_at: z.string().nullable().optional(),
    completed_at: z.string().nullable().optional(),
    attempt: z.number().default(0),
    is_placeholder: z.boolean().default(false),
  })
  .passthrough();
export type ExperimentResult = z.infer<typeof ZExperimentResult>;

export const ZExperimentResultListResponse = z
  .object({
    data: z.array(ZExperimentResult).default([]),
    list_metadata: ZListMetadata,
  })
  .passthrough();
export type ExperimentResultListResponse = z.infer<typeof ZExperimentResultListResponse>;

// ---------------------------------------------------------------------------
// Eligible blocks
// ---------------------------------------------------------------------------

export const ZEligibleBlockSummary = z
  .object({
    block_id: z.string(),
    block_label: z.string(),
    block_type: z.string(),
    experiment_count: z.number(),
    drifted_experiment_count: z.number(),
    stale_experiment_count: z.number(),
    latest_run_at: z.string().nullable().optional(),
    mean_score: z.number().nullable().optional(),
  })
  .passthrough();
export type EligibleBlockSummary = z.infer<typeof ZEligibleBlockSummary>;

export const ZEligibleBlockListResponse = z
  .object({
    blocks: z.array(ZEligibleBlockSummary).default([]),
  })
  .passthrough();
export type EligibleBlockListResponse = z.infer<typeof ZEligibleBlockListResponse>;

// ---------------------------------------------------------------------------
// Metrics
// ---------------------------------------------------------------------------

// Open shape — caller branches on `kind`
// ("summary" | "by_document" | "by_target" | "votes" | "stale_metrics" | "no_metrics").
export const ZExperimentMetricsResponse = z.record(z.any());
export type ExperimentMetricsResponse = Record<string, unknown>;
