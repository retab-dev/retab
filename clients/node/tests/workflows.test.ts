import { describe, expect, test } from "bun:test";

import { AbstractClient } from "../src/client";
import APIWorkflows from "../src/api/workflows/client";
import APIWorkflowRuns from "../src/api/workflows/runs/client";
import { getWorkflowRunOutput, raiseForStatus, WorkflowRunError } from "../src/types";
import type { WorkflowRun } from "../src/types";

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

describe("workflows client", () => {
    test("get() uses the workflow detail route", async () => {
        const mockClient = new MockClient({
            id: "workflow_123",
            name: "Test Workflow",
            description: "",
            is_published: false,
            email_senders_whitelist: [],
            email_domains_whitelist: [],
            created_at: "2026-03-12T10:00:00Z",
            updated_at: "2026-03-12T10:00:00Z",
        });
        const workflowsClient = new APIWorkflows(mockClient);

        const workflow = await workflowsClient.get("workflow_123");

        expect(mockClient.lastFetchParams).toEqual({
            url: "/workflows/workflow_123",
            method: "GET",
            params: undefined,
            headers: undefined,
        });
        expect(workflow.id).toBe("workflow_123");
    });

    test("list() uses the workflows route with pagination params", async () => {
        const mockClient = new MockClient({
            data: [],
            list_metadata: { before: null, after: "cursor_1" },
        });
        const workflowsClient = new APIWorkflows(mockClient);

        const result = await workflowsClient.list({
            limit: 5,
            order: "asc",
            sortBy: "updated_at",
            fields: "id,name",
            after: "cursor_0",
        });

        expect(mockClient.lastFetchParams).toEqual({
            url: "/workflows",
            method: "GET",
            params: {
                limit: 5,
                order: "asc",
                sort_by: "updated_at",
                fields: "id,name",
                after: "cursor_0",
            },
            headers: undefined,
        });
        expect(result.list_metadata.after).toBe("cursor_1");
    });

    test("create() sends POST to /workflows", async () => {
        const mockClient = new MockClient({
            id: "wf_new",
            name: "My Workflow",
            description: "A test",
            is_published: false,
            email_senders_whitelist: [],
            email_domains_whitelist: [],
            created_at: "2026-03-12T10:00:00Z",
            updated_at: "2026-03-12T10:00:00Z",
        });
        const workflowsClient = new APIWorkflows(mockClient);

        const wf = await workflowsClient.create({ name: "My Workflow", description: "A test" });

        expect(mockClient.lastFetchParams?.url).toBe("/workflows");
        expect(mockClient.lastFetchParams?.method).toBe("POST");
        expect(wf.id).toBe("wf_new");
        expect(wf.name).toBe("My Workflow");
    });

    test("publish() sends POST to /workflows/{id}/publish", async () => {
        const mockClient = new MockClient({
            id: "wf_1",
            name: "Test",
            is_published: true,
            published_snapshot_id: "snap_1",
            email_senders_whitelist: [],
            email_domains_whitelist: [],
            created_at: "2026-03-12T10:00:00Z",
            updated_at: "2026-03-12T10:00:00Z",
        });
        const workflowsClient = new APIWorkflows(mockClient);

        const wf = await workflowsClient.publish("wf_1", { description: "v1" });

        expect(mockClient.lastFetchParams?.url).toBe("/workflows/wf_1/publish");
        expect(mockClient.lastFetchParams?.method).toBe("POST");
        expect(wf.is_published).toBe(true);
    });

    test("getEntities() returns blocks, edges, subflows", async () => {
        const mockClient = new MockClient({
            workflow: {
                id: "wf_1", name: "Test",
                created_at: "2026-01-01T00:00:00Z", updated_at: "2026-01-01T00:00:00Z",
            },
            blocks: [
                { id: "start-1", workflow_id: "wf_1", type: "start", label: "Doc Input" },
                { id: "extract-1", workflow_id: "wf_1", type: "extract", label: "Extract" },
            ],
            edges: [
                { id: "edge-1", workflow_id: "wf_1", source_block: "start-1", target_block: "extract-1" },
            ],
            subflows: [],
        });
        const workflowsClient = new APIWorkflows(mockClient);

        const entities = await workflowsClient.getEntities("wf_1");

        expect(mockClient.lastFetchParams?.url).toBe("/workflows/wf_1/entities");
        expect(entities.blocks).toHaveLength(2);
        expect(entities.edges).toHaveLength(1);
        expect(entities.blocks.filter(b => b.type === "start")).toHaveLength(1);
    });

    test("delete() sends DELETE to /workflows/{id}", async () => {
        const mockClient = new MockClient(null);
        const workflowsClient = new APIWorkflows(mockClient);

        await workflowsClient.delete("wf_1");

        expect(mockClient.lastFetchParams?.url).toBe("/workflows/wf_1");
        expect(mockClient.lastFetchParams?.method).toBe("DELETE");
    });

    test("duplicate() sends POST to /workflows/{id}/duplicate", async () => {
        const mockClient = new MockClient({
            id: "wf_copy",
            name: "Test (Copy)",
            is_published: false,
            email_senders_whitelist: [],
            email_domains_whitelist: [],
            created_at: "2026-03-12T10:00:00Z",
            updated_at: "2026-03-12T10:00:00Z",
        });
        const workflowsClient = new APIWorkflows(mockClient);

        const wf = await workflowsClient.duplicate("wf_1");

        expect(mockClient.lastFetchParams?.url).toBe("/workflows/wf_1/duplicate");
        expect(wf.id).toBe("wf_copy");
    });

    test("runs.get() accepts newer step node types and skipped statuses", async () => {
        const mockClient = new MockClient({
            id: "run_123",
            workflow_id: "workflow_123",
            workflow_name: "Classifier Workflow",
            organization_id: "org_123",
            status: "running",
            started_at: "2026-03-13T10:00:00Z",
            steps: [
                {
                    node_id: "classifier-1",
                    node_type: "classifier",
                    node_label: "Classifier",
                    status: "completed",
                },
                {
                    node_id: "extract-2",
                    node_type: "extract",
                    node_label: "Skipped branch",
                    status: "skipped",
                },
            ],
            created_at: "2026-03-13T10:00:00Z",
            updated_at: "2026-03-13T10:00:00Z",
            waiting_for_node_ids: [],
        });
        const runsClient = new APIWorkflowRuns(mockClient);

        const run = await runsClient.get("run_123");

        expect(mockClient.lastFetchParams).toEqual({
            url: "/workflows/runs/run_123",
            method: "GET",
            params: undefined,
            headers: undefined,
        });
        expect(run.steps[0]?.node_type).toBe("classifier");
        expect(run.steps[1]?.status).toBe("skipped");
    });

    test("runs.counts() sends GET to /workflows/runs/counts", async () => {
        const mockClient = new MockClient({ total: 42, completed: 30, error: 5 });
        const runsClient = new APIWorkflowRuns(mockClient);

        const counts = await runsClient.counts({ workflowId: "wf_1" });

        expect(mockClient.lastFetchParams?.url).toBe("/workflows/runs/counts");
        expect(counts.total).toBe(42);
    });

    test("runs.enriched fields parse correctly", async () => {
        const mockClient = new MockClient({
            id: "run_789",
            workflow_id: "wf_1",
            workflow_name: "Test",
            organization_id: "org_1",
            status: "completed",
            started_at: "2026-01-01T00:00:00Z",
            created_at: "2026-01-01T00:00:00Z",
            updated_at: "2026-01-01T00:00:00Z",
            steps: [],
            waiting_for_node_ids: [],
            trigger_type: "api",
            config_snapshot_id: "snap_abc",
            cost_summary: { total: 0.05 },
            human_waiting_duration_ms: 5000,
        });
        const runsClient = new APIWorkflowRuns(mockClient);

        const run = await runsClient.get("run_789");

        expect(run.trigger_type).toBe("api");
        expect(run.config_snapshot_id).toBe("snap_abc");
        expect(run.cost_summary).toEqual({ total: 0.05 });
        expect(run.human_waiting_duration_ms).toBe(5000);
    });
});

describe("workflow run utilities", () => {
    const makeRun = (overrides: Partial<WorkflowRun>): WorkflowRun => ({
        id: "run_1",
        workflow_id: "wf_1",
        workflow_name: "Test",
        organization_id: "org_1",
        status: "completed",
        started_at: "2026-01-01T00:00:00Z",
        created_at: "2026-01-01T00:00:00Z",
        updated_at: "2026-01-01T00:00:00Z",
        steps: [],
        waiting_for_node_ids: [],
        human_waiting_duration_ms: 0,
        ...overrides,
    });

    test("getWorkflowRunOutput() single end node returns data directly", () => {
        const run = makeRun({
            final_outputs: {
                "end-1": { document: null, data: { invoice: "INV-001" } },
            },
        });
        expect(getWorkflowRunOutput(run)).toEqual({ invoice: "INV-001" });
    });

    test("getWorkflowRunOutput() multiple end nodes returns {endNodeId: data}", () => {
        const run = makeRun({
            final_outputs: {
                "end-1": { document: null, data: { count: 3 } },
                "end-2": { document: null, data: { count: 5 } },
            },
        });
        expect(getWorkflowRunOutput(run)).toEqual({
            "end-1": { count: 3 },
            "end-2": { count: 5 },
        });
    });

    test("getWorkflowRunOutput() returns null when no final_outputs", () => {
        const run = makeRun({ final_outputs: null });
        expect(getWorkflowRunOutput(run)).toBeNull();
    });

    test("raiseForStatus() throws WorkflowRunError on error status", () => {
        const run = makeRun({ status: "error", error: "Node failed" });
        expect(() => raiseForStatus(run)).toThrow(WorkflowRunError);
        try {
            raiseForStatus(run);
        } catch (e) {
            expect(e).toBeInstanceOf(WorkflowRunError);
            expect((e as WorkflowRunError).run).toBe(run);
            expect((e as WorkflowRunError).message).toContain("Node failed");
        }
    });

    test("raiseForStatus() does not throw on completed status", () => {
        const run = makeRun({ status: "completed" });
        expect(() => raiseForStatus(run)).not.toThrow();
    });
});
