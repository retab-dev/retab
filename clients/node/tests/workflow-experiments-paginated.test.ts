/**
 * Pin the canonical PaginatedList envelope (`{data, list_metadata}`) for the
 * two list endpoints under `/v1/workflows/{wf}/experiments`:
 *
 *   GET /v1/workflows/{wf}/experiments
 *   GET /v1/workflows/{wf}/experiments/{id}/runs
 *
 * Both routes used to return non-canonical shapes (a bare list and the
 * `{"runs": [...]}` named-key envelope respectively). The migration to the
 * shared envelope is a deliberate breaking change.
 */
import { describe, expect, test } from "bun:test";

import { AbstractClient } from "../src/client";
import APIWorkflowExperiments from "../src/api/workflows/experiments/client";
import APIWorkflowExperimentRuns from "../src/api/workflows/experiments/runs/client";

class MockClient extends AbstractClient {
    public lastFetchParams: Record<string, unknown> | null = null;
    private readonly responseBody: unknown;

    constructor(responseBody: unknown) {
        super();
        this.responseBody = responseBody;
    }

    protected async _fetch(params: {
        url: string;
        method: string;
        params?: Record<string, unknown>;
        headers?: Record<string, unknown>;
    }): Promise<Response> {
        this.lastFetchParams = params;
        return new Response(JSON.stringify(this.responseBody), {
            status: 200,
            headers: { "Content-Type": "application/json" },
        });
    }
}

const NOW = "2026-05-01T14:30:00Z";

const EXPERIMENT = {
    id: "exp_1",
    workflow_id: "wf_1",
    block_id: "block_1",
    n_consensus: 5,
    document_count: 0,
    name: "Q1",
    last_run_id: null,
    created_at: NOW,
    updated_at: NOW,
    status: "draft",
    block_kind: "extract",
    score: null,
    job_id: null,
    is_stale: false,
    schema_drift: "unknown",
    schema_drift_detail: null,
};

const RUN = {
    id: "exprun_1",
    parent_run_id: null,
    block_config: null,
    definition_fingerprint: "fp",
    documents_fingerprint: "fp_doc",
    status: "completed",
    block_kind: "extract",
    score: 0.83,
    total_document_count: 1,
    completed_document_count: 1,
    document_count: 1,
    error_count: 0,
    n_consensus: 5,
    created_at: NOW,
    completed_at: NOW,
    duration_ms: 100,
    job_id: "job_1",
};

describe("workflows.experiments.list paginated envelope", () => {
    test("returns {data, list_metadata}", async () => {
        const mockClient = new MockClient({
            data: [EXPERIMENT],
            list_metadata: { before: null, after: null },
        });
        const experiments = new APIWorkflowExperiments(mockClient);

        const page = await experiments.list("wf_1");

        expect(mockClient.lastFetchParams).toMatchObject({
            url: "/workflows/wf_1/experiments",
            method: "GET",
        });
        expect(page.data).toHaveLength(1);
        expect(page.data[0]?.id).toBe("exp_1");
        expect(page.list_metadata.before).toBeNull();
        expect(page.list_metadata.after).toBeNull();
    });
});

describe("workflows.experiments.runs.list paginated envelope", () => {
    test("returns {data, list_metadata} (was {runs: [...]})", async () => {
        const mockClient = new MockClient({
            data: [RUN],
            list_metadata: { before: null, after: null },
        });
        const runs = new APIWorkflowExperimentRuns(mockClient);

        const page = await runs.list({
            workflowId: "wf_1",
            experimentId: "exp_1",
        });

        expect(mockClient.lastFetchParams).toMatchObject({
            url: "/workflows/wf_1/experiments/exp_1/runs",
            method: "GET",
        });
        expect(page.data).toHaveLength(1);
        expect(page.data[0]?.id).toBe("exprun_1");
        expect(page.list_metadata.before).toBeNull();
        expect(page.list_metadata.after).toBeNull();
    });
});
