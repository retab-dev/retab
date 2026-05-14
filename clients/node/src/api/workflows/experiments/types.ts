/**
 * Zod schemas + TS types for the workflow experiments API.
 *
 * Mirrors `backend/main_server/main_server/services/v1/workflows/experiments/{models,routes}.py`
 * and the Python SDK's `retab/types/workflows/experiments.py`. Wire-format
 * field names are snake_case.
 *
 * The metrics response is a deeply nested discriminated union (four "views"
 * plus two error envelopes). Modeling the full union here would couple the
 * SDK to internal structural details that are still evolving. We expose:
 *
 *   - typed wrappers for the simple/stable shapes (Experiment, runs,
 *     eligible blocks, run-batch, job)
 *   - `ExperimentMetricsResponse = Record<string, unknown>` — caller
 *     branches on the `view` field for view-specific payloads, or on
 *     `error` for stale/missing envelopes.
 */

import * as z from "zod";

// ---------------------------------------------------------------------------
// Enums
// ---------------------------------------------------------------------------

export const ZNConsensusValue = z.union([z.literal(3), z.literal(5), z.literal(7)]);
export type NConsensusValue = z.infer<typeof ZNConsensusValue>;

export const ZExperimentBlockKind = z.enum(["extract", "classifier", "split", "for_each"]);
export type ExperimentBlockKind = z.infer<typeof ZExperimentBlockKind>;

export const ZExperimentRunStatus = z.enum([
    "pending",
    "running",
    "completed",
    "error",
    "cancelled",
]);
export type ExperimentRunStatus = z.infer<typeof ZExperimentRunStatus>;

export const ZExperimentJobStatus = z.enum([
    "pending",
    "running",
    "completed",
    "error",
]);
export type ExperimentJobStatus = z.infer<typeof ZExperimentJobStatus>;

export const ZExperimentPublicStatus = z.enum([
    "draft",
    "processing",
    "completed",
    "failed",
    "cancelled",
]);
export type ExperimentPublicStatus = z.infer<typeof ZExperimentPublicStatus>;

export const ZSchemaDriftStatus = z.enum(["unknown", "none", "drifted"]);
export type SchemaDriftStatus = z.infer<typeof ZSchemaDriftStatus>;

export const ZExperimentMetricView = z.enum([
    "summary",
    "by_document",
    "by_target",
    "votes",
]);
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

export const ZExperimentResponse = z
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
        status: ZExperimentPublicStatus.default("draft"),
        block_kind: ZExperimentBlockKind,
        score: z.number().nullable().optional(),
        job_id: z.string().nullable().optional(),
        is_stale: z.boolean().default(false),
        schema_drift: ZSchemaDriftStatus.default("unknown"),
        schema_drift_detail: z.string().nullable().optional(),
    })
    .passthrough();
export type ExperimentResponse = z.infer<typeof ZExperimentResponse>;

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
        data: z.array(ZExperimentResponse).default([]),
        list_metadata: ZListMetadata,
    })
    .passthrough();
export type ExperimentList = z.infer<typeof ZExperimentList>;

// ---------------------------------------------------------------------------
// Run (response)
// ---------------------------------------------------------------------------

export const ZPreviousRunSummary = z
    .object({
        run_id: z.string(),
        definition_fingerprint: z.string().nullable().optional(),
        score: z.number().nullable().optional(),
    })
    .passthrough();
export type PreviousRunSummary = z.infer<typeof ZPreviousRunSummary>;

export const ZRunExperimentResponse = z
    .object({
        experiment_id: z.string(),
        run_id: z.string(),
        job_id: z.string().nullable(),
        status: z.string(),
        definition_fingerprint: z.string(),
        total_document_count: z.number().default(0),
        completed_document_count: z.number().default(0),
        document_count: z.number(),
        n_consensus: ZNConsensusValue,
        previous_run: ZPreviousRunSummary.nullable().optional(),
        noop: z.boolean().default(false),
    })
    .passthrough();
export type RunExperimentResponse = z.infer<typeof ZRunExperimentResponse>;

export const ZCancelExperimentResponse = z
    .object({
        status: z.string(),
        run_id: z.string(),
    })
    .passthrough();
export type CancelExperimentResponse = z.infer<typeof ZCancelExperimentResponse>;

export const ZExperimentRunSummary = z
    .object({
        id: z.string(),
        parent_run_id: z.string().nullable().optional(),
        block_config: z.record(z.any()).nullable().optional(),
        definition_fingerprint: z.string(),
        documents_fingerprint: z.string(),
        status: ZExperimentRunStatus,
        block_kind: ZExperimentBlockKind,
        score: z.number().nullable().optional(),
        total_document_count: z.number().default(0),
        completed_document_count: z.number().default(0),
        document_count: z.number().default(0),
        error_count: z.number().default(0),
        n_consensus: ZNConsensusValue,
        created_at: z.string(),
        completed_at: z.string().nullable().optional(),
        duration_ms: z.number().nullable().optional(),
        job_id: z.string().nullable().optional(),
    })
    .passthrough();
export type ExperimentRunSummary = z.infer<typeof ZExperimentRunSummary>;

// Canonical PaginatedList envelope for `GET /experiments/{id}/runs`. The
// route used to return `{"runs": [...]}`; the migration to
// `{data, list_metadata}` matches the rest of the Retab list endpoints.
export const ZExperimentRunListResponse = z
    .object({
        data: z.array(ZExperimentRunSummary).default([]),
        list_metadata: ZListMetadata,
    })
    .passthrough();
export type ExperimentRunListResponse = z.infer<typeof ZExperimentRunListResponse>;

// ---------------------------------------------------------------------------
// Per-document job + content
// ---------------------------------------------------------------------------

export const ZStepArtifactRefMini = z
    .object({
        operation: z.string(),
        id: z.string(),
    })
    .passthrough();
export type StepArtifactRefMini = z.infer<typeof ZStepArtifactRefMini>;

export const ZExperimentJobResponse = z
    .object({
        id: z.string(),
        run_id: z.string(),
        experiment_id: z.string(),
        document_id: z.string(),
        status: ZExperimentJobStatus,
        block_kind: ZExperimentBlockKind,
        handle_inputs: z.record(z.any()).default({}),
        artifact: ZStepArtifactRefMini.nullable().optional(),
        error: z.string().nullable().optional(),
        duration_ms: z.number().nullable().optional(),
        created_at: z.string().nullable().optional(),
        started_at: z.string().nullable().optional(),
        completed_at: z.string().nullable().optional(),
        attempt: z.number().default(0),
        is_placeholder: z.boolean().default(false),
    })
    .passthrough();
export type ExperimentJobResponse = z.infer<typeof ZExperimentJobResponse>;

export const ZExperimentContent = z
    .object({
        jobs: z.array(ZExperimentJobResponse).default([]),
    })
    .passthrough();
export type ExperimentContent = z.infer<typeof ZExperimentContent>;

export const ZExperimentContentResponse = z
    .object({
        experiment_id: z.string(),
        run_id: z.string(),
        content: ZExperimentContent,
    })
    .passthrough();
export type ExperimentContentResponse = z.infer<typeof ZExperimentContentResponse>;

// ---------------------------------------------------------------------------
// Eligible blocks + run-batch
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

export const ZRunBatchResponse = z
    .object({
        block_id: z.string(),
        experiment_count: z.number(),
        runs: z.array(ZRunExperimentResponse).default([]),
    })
    .passthrough();
export type RunBatchResponse = z.infer<typeof ZRunBatchResponse>;

// ---------------------------------------------------------------------------
// Metrics
// ---------------------------------------------------------------------------

// Open shape — caller branches on `view` ("summary" | "by_document" | …)
// or `error` ("stale_metrics" | "no_metrics") to interpret the payload.
export const ZExperimentMetricsResponse = z.record(z.any());
export type ExperimentMetricsResponse = Record<string, unknown>;
