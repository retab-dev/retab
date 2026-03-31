import { describe, expect, test } from "bun:test";

import { AbstractClient } from "../src/client";
import APIWorkflows from "../src/api/workflows/client";
import APIWorkflowRuns from "../src/api/workflows/runs/client";
import APIWorkflowBlocks from "../src/api/workflows/blocks/client";
import APIWorkflowEdges from "../src/api/workflows/edges/client";
import { raiseForStatus, WorkflowRunError } from "../src/types";
import type { WorkflowRun, WorkflowRunExportResponse } from "../src/types";
import { ZFunctionStepOutput } from "../src/generated_types";

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

    test("function step output schema matches python sdk fields", () => {
        const parsed = ZFunctionStepOutput.parse({
            message: "Function executed successfully",
            execution_time_ms: 12.5,
            stdout: "hello",
            stderr: null,
            error: null,
            traceback_str: null,
            json_schema: {
                type: "object",
                properties: {
                    result: { type: "string" },
                },
            },
        });

        expect(parsed.message).toBe("Function executed successfully");
        expect(parsed.execution_time_ms).toBe(12.5);
        expect(parsed.stdout).toBe("hello");
        expect(parsed.json_schema?.properties).toEqual({
            result: { type: "string" },
        });
    });

    test("runs.list() serializes array and date filters", async () => {
        const mockClient = new MockClient({
            data: [],
            list_metadata: { before: null, after: null },
        });
        const runsClient = new APIWorkflowRuns(mockClient);

        await runsClient.list({
            workflowId: "wf_1",
            statuses: ["completed", "error"],
            triggerTypes: ["api", "email"],
            fromDate: new Date("2026-01-01T00:00:00.000Z"),
            toDate: new Date("2026-01-31T00:00:00.000Z"),
            fields: ["id", "status"],
            after: "cursor_1",
        });

        expect(mockClient.lastFetchParams).toEqual({
            url: "/workflows/runs",
            method: "GET",
            params: {
                workflow_id: "wf_1",
                statuses: "completed,error",
                trigger_types: "api,email",
                from_date: "2026-01-01",
                to_date: "2026-01-31",
                fields: "id,status",
                after: "cursor_1",
                limit: 20,
                order: "desc",
            },
            headers: undefined,
        });
    });

    test("blocks.createBatch() accepts camelCase request objects", async () => {
        const mockClient = new MockClient([
            { id: "start-1", workflow_id: "wf_1", type: "start" },
            { id: "extract-1", workflow_id: "wf_1", type: "extract" },
        ]);
        const blocksClient = new APIWorkflowBlocks(mockClient);

        const blocks = await blocksClient.createBatch("wf_1", [
            { id: "start-1", type: "start" },
            { id: "extract-1", type: "extract", positionX: 120, positionY: 80 },
        ]);

        expect(mockClient.lastFetchParams).toEqual({
            url: "/workflows/wf_1/blocks/batch",
            method: "POST",
            body: [
                { id: "start-1", type: "start", label: "", position_x: 0, position_y: 0, width: undefined, height: undefined, config: undefined, subflow_id: undefined, parent_id: undefined },
                { id: "extract-1", type: "extract", label: "", position_x: 120, position_y: 80, width: undefined, height: undefined, config: undefined, subflow_id: undefined, parent_id: undefined },
            ],
            params: undefined,
            headers: undefined,
        });
        expect(blocks.map((block) => block.id)).toEqual(["start-1", "extract-1"]);
    });

    test("edges.createBatch() accepts camelCase request objects", async () => {
        const mockClient = new MockClient([
            { id: "edge-1", workflow_id: "wf_1", source_block: "start-1", target_block: "extract-1" },
        ]);
        const edgesClient = new APIWorkflowEdges(mockClient);

        const edges = await edgesClient.createBatch("wf_1", [
            { id: "edge-1", sourceBlock: "start-1", targetBlock: "extract-1" },
        ]);

        expect(mockClient.lastFetchParams).toEqual({
            url: "/workflows/wf_1/edges/batch",
            method: "POST",
            body: [
                { id: "edge-1", source_block: "start-1", target_block: "extract-1", source_handle: undefined, target_handle: undefined },
            ],
            params: undefined,
            headers: undefined,
        });
        expect(edges.map((edge) => edge.id)).toEqual(["edge-1"]);
    });

    test("runs.cancel() sends POST to /cancel", async () => {
        const mockClient = new MockClient({
            run: {
                id: "run_1",
                workflow_id: "wf_1",
                workflow_name: "Test",
                organization_id: "org_1",
                status: "cancelled",
                started_at: "2026-01-01T00:00:00Z",
                created_at: "2026-01-01T00:00:00Z",
                updated_at: "2026-01-01T00:00:00Z",
                steps: [],
                waiting_for_node_ids: [],
            },
            cancellation_status: "cancelled",
        });
        const runsClient = new APIWorkflowRuns(mockClient);

        const result = await runsClient.cancel("run_1", { commandId: "cmd_1" });

        expect(mockClient.lastFetchParams).toEqual({
            url: "/workflows/runs/run_1/cancel",
            method: "POST",
            body: { command_id: "cmd_1" },
            params: undefined,
            headers: undefined,
        });
        expect(result.cancellation_status).toBe("cancelled");
    });

    test("runs.restart() sends POST to /restart", async () => {
        const mockClient = new MockClient({
            id: "run_2",
            workflow_id: "wf_1",
            workflow_name: "Test",
            organization_id: "org_1",
            status: "running",
            started_at: "2026-01-01T00:00:00Z",
            created_at: "2026-01-01T00:00:00Z",
            updated_at: "2026-01-01T00:00:00Z",
            steps: [],
            waiting_for_node_ids: [],
        });
        const runsClient = new APIWorkflowRuns(mockClient);

        const run = await runsClient.restart("run_1", { commandId: "cmd_2" });

        expect(mockClient.lastFetchParams).toEqual({
            url: "/workflows/runs/run_1/restart",
            method: "POST",
            body: { command_id: "cmd_2" },
            params: undefined,
            headers: undefined,
        });
        expect(run.id).toBe("run_2");
    });

    test("runs.submitHilDecision() sends POST to /hil-decisions", async () => {
        const mockClient = new MockClient({
            submission_status: "accepted",
            decision: {
                run_id: "run_1",
                node_id: "hil-1",
                node_status: "waiting_for_hil",
                decision_received: true,
                decision_applied: false,
                approved: true,
                modified_data: { field: "value" },
                payload_hash: "hash_1",
                received_at: "2026-01-01T00:00:01Z",
                applied_at: null,
            },
        });
        const runsClient = new APIWorkflowRuns(mockClient);

        const result = await runsClient.submitHilDecision("run_1", {
            nodeId: "hil-1",
            approved: true,
            modifiedData: { field: "value" },
            commandId: "cmd_3",
        });

        expect(mockClient.lastFetchParams).toEqual({
            url: "/workflows/runs/run_1/hil-decisions",
            method: "POST",
            body: {
                node_id: "hil-1",
                approved: true,
                modified_data: { field: "value" },
                command_id: "cmd_3",
            },
            params: undefined,
            headers: undefined,
        });
        expect(result.submission_status).toBe("accepted");
    });

    test("runs.submitHilDecision() accepts already_applied responses", async () => {
        const mockClient = new MockClient({
            submission_status: "already_applied",
            decision: {
                run_id: "run_1",
                node_id: "hil-1",
                node_status: "completed",
                decision_received: true,
                decision_applied: true,
                approved: true,
                modified_data: { field: "value" },
                payload_hash: "hash_1",
                received_at: "2026-01-01T00:00:01Z",
                applied_at: "2026-01-01T00:00:05Z",
            },
        });
        const runsClient = new APIWorkflowRuns(mockClient);

        const result = await runsClient.submitHilDecision("run_1", {
            nodeId: "hil-1",
            approved: true,
        });

        expect(result.submission_status).toBe("already_applied");
        expect(result.decision.decision_applied).toBe(true);
    });

    test("runs.getHilDecision() sends GET to /hil-decisions/{nodeId}", async () => {
        const mockClient = new MockClient({
            run_id: "run_1",
            node_id: "hil-1",
            node_status: "completed",
            decision_received: true,
            decision_applied: true,
            approved: true,
            modified_data: { field: "value" },
            payload_hash: "hash_1",
            received_at: "2026-01-01T00:00:01Z",
            applied_at: "2026-01-01T00:00:05Z",
        });
        const runsClient = new APIWorkflowRuns(mockClient);

        const result = await runsClient.getHilDecision("run_1", "hil-1");

        expect(mockClient.lastFetchParams).toEqual({
            url: "/workflows/runs/run_1/hil-decisions/hil-1",
            method: "GET",
            params: undefined,
            headers: undefined,
        });
        expect(result.decision_applied).toBe(true);
    });

    test("runs.export() sends POST to /export_payload", async () => {
        const mockClient = new MockClient({
            csv_data: "a,b\n1,2\n",
            rows: 1,
            columns: 2,
        } satisfies WorkflowRunExportResponse);
        const runsClient = new APIWorkflowRuns(mockClient);

        const result = await runsClient.export({
            workflowId: "wf_1",
            nodeId: "extract-1",
            exportSource: "outputs",
            selectedRunIds: ["run_1", "run_2"],
            fromDate: new Date("2026-01-01T00:00:00.000Z"),
            toDate: new Date("2026-01-31T00:00:00.000Z"),
            triggerTypes: ["api"],
            preferredColumns: ["invoice_number", "total_amount"],
        });

        expect(mockClient.lastFetchParams).toEqual({
            url: "/workflows/runs/export_payload",
            method: "POST",
            body: {
                workflow_id: "wf_1",
                node_id: "extract-1",
                export_source: "outputs",
                preferred_columns: ["invoice_number", "total_amount"],
                selected_run_ids: ["run_1", "run_2"],
                from_date: "2026-01-01",
                to_date: "2026-01-31",
                trigger_types: ["api"],
            },
            params: undefined,
            headers: undefined,
        });
        expect(result.rows).toBe(1);
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

    test("final_outputs are preserved as raw backend payload", () => {
        const run = makeRun({
            final_outputs: {
                "end-1": { document: null, data: { invoice: "INV-001" } },
                "end-2": { document: { id: "file_1", filename: "out.pdf", mime_type: "application/pdf" }, data: { count: 5 } },
            },
        });
        expect(run.final_outputs).toEqual({
            "end-1": { document: null, data: { invoice: "INV-001" } },
            "end-2": { document: { id: "file_1", filename: "out.pdf", mime_type: "application/pdf" }, data: { count: 5 } },
        });
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

    test("raiseForStatus() throws WorkflowRunError on cancelled status", () => {
        const run = makeRun({ status: "cancelled" });
        expect(() => raiseForStatus(run)).toThrow(WorkflowRunError);
        try {
            raiseForStatus(run);
        } catch (e) {
            expect(e).toBeInstanceOf(WorkflowRunError);
            expect((e as WorkflowRunError).message).toContain("cancelled");
        }
    });

    test("waitForCompletion() returns waiting_for_human runs", async () => {
        class WaitMockClient extends AbstractClient {
            public lastFetchParams: Record<string, unknown> | null = null;
            private responses = [
                {
                    id: "run_1",
                    workflow_id: "wf_1",
                    workflow_name: "Test",
                    organization_id: "org_1",
                    status: "running",
                    started_at: "2026-01-01T00:00:00Z",
                    created_at: "2026-01-01T00:00:00Z",
                    updated_at: "2026-01-01T00:00:00Z",
                    steps: [],
                    waiting_for_node_ids: [],
                },
                {
                    id: "run_1",
                    workflow_id: "wf_1",
                    workflow_name: "Test",
                    organization_id: "org_1",
                    status: "waiting_for_human",
                    started_at: "2026-01-01T00:00:00Z",
                    created_at: "2026-01-01T00:00:00Z",
                    updated_at: "2026-01-01T00:00:01Z",
                    steps: [],
                    waiting_for_node_ids: ["hil-1"],
                },
            ];

            protected async _fetch(params: {
                url: string;
                method: string;
                params?: Record<string, unknown>;
                headers?: Record<string, unknown>;
                body?: unknown;
            }): Promise<Response> {
                this.lastFetchParams = params;
                return new Response(JSON.stringify(this.responses.shift()), {
                    status: 200,
                    headers: { "Content-Type": "application/json" },
                });
            }
        }

        const runsClient = new APIWorkflowRuns(new WaitMockClient());
        const run = await runsClient.waitForCompletion("run_1", { pollIntervalMs: 1 });

        expect(run.status).toBe("waiting_for_human");
        expect(run.waiting_for_node_ids).toEqual(["hil-1"]);
    });

    test("waitForCompletion() awaits async status callbacks", async () => {
        const mockClient = new MockClient({
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
        });
        const runsClient = new APIWorkflowRuns(mockClient);
        const seenRunIds: string[] = [];

        const run = await runsClient.waitForCompletion("run_1", {
            onStatus: async (workflowRun) => {
                seenRunIds.push(workflowRun.id);
            },
        });

        expect(run.status).toBe("completed");
        expect(seenRunIds).toEqual(["run_1"]);
    });
});
