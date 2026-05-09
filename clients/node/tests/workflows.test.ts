import { describe, expect, test } from "bun:test";

import { AbstractClient } from "../src/client";
import APIWorkflows from "../src/api/workflows/client";
import APIWorkflowRuns from "../src/api/workflows/runs/client";
import APIWorkflowBlocks from "../src/api/workflows/blocks/client";
import APIWorkflowEdges from "../src/api/workflows/edges/client";
import { raiseForStatus, WorkflowRunError, ZStepExecutionResponse, ZWorkflowRun } from "../src/types";
import type { WorkflowRun, WorkflowRunExportResponse } from "../src/types";

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

// Build a minimal v2 WorkflowRun JSON payload for fixtures.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function makeV2Run(overrides: Record<string, any> = {}): Record<string, any> {
    const {
        id = "run_1",
        organization_id = "org_1",
        workflow = {
            workflow_id: "wf_1",
            snapshot_id: "snap_1",
            name_at_run_time: "Test",
        },
        trigger = { type: "manual" },
        lifecycle = { kind: "running" },
        timing = {
            created_at: "2026-01-01T00:00:00Z",
            started_at: "2026-01-01T00:00:00Z",
        },
        ...rest
    } = overrides;
    return {
        id,
        organization_id,
        workflow,
        trigger,
        lifecycle,
        timing,
        ...rest,
    };
}

describe("workflows client", () => {
    test("get() uses the workflow detail route", async () => {
        const mockClient = new MockClient({
            id: "workflow_123",
            name: "Test Workflow",
            description: "",
            published: null,
            email_trigger: { allowed_senders: [], allowed_domains: [] },
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
            published: null,
            email_trigger: { allowed_senders: [], allowed_domains: [] },
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
            published: {
                snapshot_id: "snap_1",
                published_at: "2026-03-12T10:00:00Z",
            },
            email_trigger: { allowed_senders: [], allowed_domains: [] },
            created_at: "2026-03-12T10:00:00Z",
            updated_at: "2026-03-12T10:00:00Z",
        });
        const workflowsClient = new APIWorkflows(mockClient);

        const wf = await workflowsClient.publish("wf_1", { description: "v1" });

        expect(mockClient.lastFetchParams?.url).toBe("/workflows/wf_1/publish");
        expect(mockClient.lastFetchParams?.method).toBe("POST");
        expect(wf.published?.snapshot_id).toBe("snap_1");
    });

    test("getEntities() returns blocks, edges, subflows", async () => {
        const mockClient = new MockClient({
            workflow: {
                id: "wf_1", name: "Test",
                created_at: "2026-01-01T00:00:00Z", updated_at: "2026-01-01T00:00:00Z",
            },
            blocks: [
                { id: "start-1", workflow_id: "wf_1", organization_id: "org_1", draft_version: "draft_1", type: "start", label: "Doc Input" },
                { id: "extract-1", workflow_id: "wf_1", organization_id: "org_1", draft_version: "draft_1", type: "extract", label: "Extract" },
            ],
            edges: [
                {
                    id: "edge-1",
                    workflow_id: "wf_1",
                    organization_id: "org_1",
                    draft_version: "draft_1",
                    source_block: "start-1",
                    target_block: "extract-1",
                },
            ],
            subflows: [],
        });
        const workflowsClient = new APIWorkflows(mockClient);

        const entities = await workflowsClient.getEntities("wf_1");

        expect(mockClient.lastFetchParams?.url).toBe("/workflows/wf_1/entities");
        expect(entities.blocks).toHaveLength(2);
        expect(entities.edges).toHaveLength(1);
        expect(entities.edges[0]?.organization_id).toBe("org_1");
        expect(entities.edges[0]?.draft_version).toBe("draft_1");
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
            published: null,
            email_trigger: { allowed_senders: [], allowed_domains: [] },
            created_at: "2026-03-12T10:00:00Z",
            updated_at: "2026-03-12T10:00:00Z",
        });
        const workflowsClient = new APIWorkflows(mockClient);

        const wf = await workflowsClient.duplicate("wf_1");

        expect(mockClient.lastFetchParams?.url).toBe("/workflows/wf_1/duplicate");
        expect(wf.id).toBe("wf_copy");
    });

    test("runs.get() ignores legacy steps field (now fetched via steps.list())", async () => {
        // Older servers may still include `steps` in the run payload; the SDK
        // no longer surfaces it on the WorkflowRun type. Per-step records are
        // fetched via `client.workflows.runs.steps.list(run_id)`.
        const mockClient = new MockClient({
            id: "run_123",
            organization_id: "org_123",
            workflow: {
                workflow_id: "workflow_123",
                snapshot_id: "snap_1",
                name_at_run_time: "Classifier Workflow",
            },
            trigger: { type: "manual" },
            lifecycle: { kind: "running" },
            timing: {
                created_at: "2026-03-13T10:00:00Z",
                started_at: "2026-03-13T10:00:00Z",
            },
            steps: [
                {
                    block_id: "classifier-1",
                    step_id: "classifier-1",
                    block_type: "classifier",
                    block_label: "Classifier",
                    status: "completed",
                },
            ],
        });
        const runsClient = new APIWorkflowRuns(mockClient);

        const run = await runsClient.get("run_123");

        expect(mockClient.lastFetchParams).toEqual({
            url: "/workflows/runs/run_123",
            method: "GET",
            params: undefined,
            headers: undefined,
        });
        // The legacy `steps` payload is silently dropped by the Zod schema —
        // it is no longer a property on the parsed WorkflowRun.
        expect((run as unknown as Record<string, unknown>).steps).toBeUndefined();
        expect(run.id).toBe("run_123");
        expect(run.lifecycle.kind).toBe("running");
    });

    test("runs.v2 typed fields parse correctly", async () => {
        const mockClient = new MockClient({
            id: "run_789",
            organization_id: "org_1",
            workflow: {
                workflow_id: "wf_1",
                snapshot_id: "snap_abc",
                name_at_run_time: "Test",
            },
            trigger: { type: "api", api_key_id: "ak_1" },
            lifecycle: { kind: "completed" },
            timing: {
                created_at: "2026-01-01T00:00:00Z",
                started_at: "2026-01-01T00:00:00Z",
                completed_at: "2026-01-01T00:00:05Z",
                accumulated_human_waiting_ms: 5000,
            },
            inputs: { documents: {}, json_data: {} },
            observability: {
                total_cost: { total: 0.05 },
                total_tokens: {},
                cost_by_model: {},
                tokens_by_model: {},
            },
        });
        const runsClient = new APIWorkflowRuns(mockClient);

        const run = await runsClient.get("run_789");

        expect(run.trigger.type).toBe("api");
        expect(run.workflow.workflow_id).toBe("wf_1");
        expect(run.workflow.snapshot_id).toBe("snap_abc");
        expect(run.observability?.total_cost).toEqual({ total: 0.05 });
        expect(run.timing.accumulated_human_waiting_ms).toBe(5000);
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
            { id: "start-1", workflow_id: "wf_1", organization_id: "org_1", draft_version: "draft_1", type: "start" },
            { id: "extract-1", workflow_id: "wf_1", organization_id: "org_1", draft_version: "draft_1", type: "extract" },
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
                { id: "start-1", type: "start", label: "", position_x: 0, position_y: 0, width: undefined, height: undefined, config: undefined, parent_id: undefined },
                { id: "extract-1", type: "extract", label: "", position_x: 120, position_y: 80, width: undefined, height: undefined, config: undefined, parent_id: undefined },
            ],
            params: undefined,
            headers: undefined,
        });
        expect(blocks.map((block) => block.id)).toEqual(["start-1", "extract-1"]);
    });

    test("workflow blocks expose live-editing metadata", async () => {
        const mockClient = new MockClient({
            id: "extract-1",
            workflow_id: "wf_1",
            organization_id: "org_1",
            draft_version: "draft_1",
            type: "extract",
            label: "Extract",
        });
        const blocksClient = new APIWorkflowBlocks(mockClient);

        const block = await blocksClient.get("wf_1", "extract-1");

        expect(block.organization_id).toBe("org_1");
        expect(block.draft_version).toBe("draft_1");
    });

    test("workflow blocks expose resolved schema sidecars", async () => {
        const mockClient = new MockClient({
            id: "extract-1",
            workflow_id: "wf_1",
            organization_id: "org_1",
            draft_version: "draft_1",
            type: "extract",
            label: "Extract",
            resolved_schemas: {
                input_schemas: {},
                output_schemas: {
                    "output-json-0": {
                        type: "object",
                        properties: { invoice_number: { type: "string" } },
                    },
                },
            },
        });
        const blocksClient = new APIWorkflowBlocks(mockClient);

        const block = await blocksClient.get("wf_1", "extract-1");

        expect(block.resolved_schemas?.output_schemas["output-json-0"]?.properties.invoice_number.type).toBe("string");
    });

    test("step execution responses ignore removed payload schemas", () => {
        const removedPayloadKey = ["raw", "_", "output"].join("");
        const parsed = ZStepExecutionResponse.parse({
            block_id: "extract-1",
            block_type: "extract",
            block_label: "Extract",
            status: "completed",
            output: {
                data: { invoice_number: "INV-001" },
                json_schema: { type: "object" },
            },
            [removedPayloadKey]: {
                json_schema: { type: "object" },
            },
            json_schema: { type: "object" },
            artifact_view: {
                block_type: "extract",
                artifact: { operation: "extraction", id: "ext_123" },
                data: {
                    output: { invoice_number: "INV-001" },
                    extraction_id: "ext_123",
                },
            },
            handle_outputs: {
                "output-json-0": {
                    type: "json",
                    data: { invoice_number: "INV-001" },
                },
            },
        });

        expect("output" in parsed).toBe(false);
        expect(removedPayloadKey in parsed).toBe(false);
        expect("json_schema" in parsed).toBe(false);
        // artifact_view is part of StepExecutionResponse (block-specific render data).
        expect(parsed.artifact_view?.block_type).toBe("extract");
        expect(parsed.artifact_view?.artifact?.id).toBe("ext_123");
        expect(parsed.handle_outputs?.["output-json-0"]?.data?.invoice_number).toBe("INV-001");
    });

    test("edges.createBatch() accepts camelCase request objects", async () => {
        const mockClient = new MockClient([
            {
                id: "edge-1",
                workflow_id: "wf_1",
                organization_id: "org_1",
                draft_version: "draft_1",
                source_block: "start-1",
                target_block: "extract-1",
            },
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
            run: makeV2Run({ lifecycle: { kind: "cancelled" } }),
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
        const mockClient = new MockClient(
            makeV2Run({ id: "run_2", lifecycle: { kind: "running" } })
        );
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
                block_id: "hil-1",
                block_status: "waiting_for_hil",
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
            blockId: "hil-1",
            approved: true,
            modifiedData: { field: "value" },
            commandId: "cmd_3",
        });

        expect(mockClient.lastFetchParams).toEqual({
            url: "/workflows/runs/run_1/hil-decisions",
            method: "POST",
            body: {
                block_id: "hil-1",
                approved: true,
                modified_data: { field: "value" },
                command_id: "cmd_3",
            },
            params: undefined,
            headers: undefined,
        });
        expect(result.submission_status).toBe("accepted");
        expect(result.decision.block_status).toBe("waiting_for_hil");
    });

    test("runs.submitHilDecision() accepts already_applied responses", async () => {
        const mockClient = new MockClient({
            submission_status: "already_applied",
            decision: {
                run_id: "run_1",
                block_id: "hil-1",
                block_status: "completed",
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
            blockId: "hil-1",
            approved: true,
        });

        expect(result.submission_status).toBe("already_applied");
        expect(result.decision.decision_applied).toBe(true);
        expect(result.decision.block_status).toBe("completed");
    });

    test("runs.getHilDecision() sends GET to /hil-decisions/{blockId}", async () => {
        const mockClient = new MockClient({
            run_id: "run_1",
            block_id: "hil-1",
            block_status: "completed",
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
        expect(result.block_status).toBe("completed");
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
            blockId: "extract-1",
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
                block_id: "extract-1",
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
    // Parses a v2 JSON payload into a fully-typed WorkflowRun (so we can
    // hand `raiseForStatus`/`finalOutputs` real `WorkflowRun` objects).
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const parseRun = (overrides: Record<string, any> = {}): WorkflowRun =>
        ZWorkflowRun.parse(makeV2Run(overrides));

    test("raiseForStatus() throws WorkflowRunError on error lifecycle", () => {
        const run = parseRun({
            lifecycle: { kind: "error", message: "Node failed" },
            timing: {
                created_at: "2026-01-01T00:00:00Z",
                started_at: "2026-01-01T00:00:00Z",
                completed_at: "2026-01-01T00:00:01Z",
            },
        });
        expect(() => raiseForStatus(run)).toThrow(WorkflowRunError);
        try {
            raiseForStatus(run);
        } catch (e) {
            expect(e).toBeInstanceOf(WorkflowRunError);
            expect((e as WorkflowRunError).run).toBe(run);
            expect((e as WorkflowRunError).message).toContain("Node failed");
        }
    });

    test("raiseForStatus() does not throw on completed lifecycle", () => {
        const run = parseRun({
            lifecycle: { kind: "completed" },
            timing: {
                created_at: "2026-01-01T00:00:00Z",
                started_at: "2026-01-01T00:00:00Z",
                completed_at: "2026-01-01T00:00:01Z",
            },
        });
        expect(() => raiseForStatus(run)).not.toThrow();
    });

    test("raiseForStatus() throws WorkflowRunError on cancelled lifecycle", () => {
        const run = parseRun({
            lifecycle: { kind: "cancelled", reason: "user cancelled" },
            timing: {
                created_at: "2026-01-01T00:00:00Z",
                started_at: "2026-01-01T00:00:00Z",
                completed_at: "2026-01-01T00:00:01Z",
            },
        });
        expect(() => raiseForStatus(run)).toThrow(WorkflowRunError);
        try {
            raiseForStatus(run);
        } catch (e) {
            expect(e).toBeInstanceOf(WorkflowRunError);
            expect((e as WorkflowRunError).message).toContain("cancelled");
        }
    });

    test("finalOutputs() returns end-block handle_outputs for completed runs", async () => {
        class StepsMockClient extends AbstractClient {
            public lastFetchParams: Record<string, unknown> | null = null;

            protected async _fetch(params: {
                url: string;
                method: string;
                params?: Record<string, unknown>;
                headers?: Record<string, unknown>;
                body?: unknown;
            }): Promise<Response> {
                this.lastFetchParams = params;
                return new Response(
                    JSON.stringify([
                        {
                            run_id: "run_done",
                            organization_id: "org_1",
                            block_id: "end-1",
                            step_id: "end-1",
                            block_type: "end",
                            block_label: "End",
                            status: "completed",
                            handle_outputs: {
                                "output-json-0": {
                                    type: "json",
                                    data: { invoice_number: "INV-001" },
                                },
                            },
                            handle_inputs: {},
                        },
                        {
                            run_id: "run_done",
                            organization_id: "org_1",
                            block_id: "extract-1",
                            step_id: "extract-1",
                            block_type: "extract",
                            block_label: "Extract",
                            status: "completed",
                            handle_outputs: {},
                            handle_inputs: {},
                        },
                    ]),
                    { status: 200, headers: { "Content-Type": "application/json" } }
                );
            }
        }

        const run = parseRun({
            id: "run_done",
            lifecycle: { kind: "completed" },
            timing: {
                created_at: "2026-01-01T00:00:00Z",
                started_at: "2026-01-01T00:00:00Z",
                completed_at: "2026-01-01T00:00:01Z",
            },
        });
        const runsClient = new APIWorkflowRuns(new StepsMockClient());
        const outputs = await runsClient.finalOutputs(run);

        expect(Object.keys(outputs)).toEqual(["end-1"]);
        expect(outputs["end-1"]?.["output-json-0"]?.data).toEqual({
            invoice_number: "INV-001",
        });
    });

    test("finalOutputs() short-circuits to {} for non-completed runs (no fetch)", async () => {
        class NoOpClient extends AbstractClient {
            public fetched = false;
            protected async _fetch(): Promise<Response> {
                this.fetched = true;
                return new Response("[]", {
                    status: 200,
                    headers: { "Content-Type": "application/json" },
                });
            }
        }

        const run = parseRun({ lifecycle: { kind: "running" } });
        const noOp = new NoOpClient();
        const runsClient = new APIWorkflowRuns(noOp);

        expect(await runsClient.finalOutputs(run)).toEqual({});
        expect(noOp.fetched).toBe(false);
    });

    test("waitForCompletion() returns waiting_for_human runs", async () => {
        class WaitMockClient extends AbstractClient {
            public lastFetchParams: Record<string, unknown> | null = null;
            private responses = [
                makeV2Run({ lifecycle: { kind: "running" } }),
                makeV2Run({
                    lifecycle: {
                        kind: "waiting_for_human",
                        waiting_for_block_ids: ["hil-1"],
                    },
                }),
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

        expect(run.lifecycle.kind).toBe("waiting_for_human");
        if (run.lifecycle.kind === "waiting_for_human") {
            expect(run.lifecycle.waiting_for_block_ids).toEqual(["hil-1"]);
        }
    });

    test("waitForCompletion() awaits async status callbacks", async () => {
        const mockClient = new MockClient(
            makeV2Run({ lifecycle: { kind: "completed" } })
        );
        const runsClient = new APIWorkflowRuns(mockClient);
        const seenRunIds: string[] = [];

        const run = await runsClient.waitForCompletion("run_1", {
            onStatus: async (workflowRun) => {
                seenRunIds.push(workflowRun.id);
            },
        });

        expect(run.lifecycle.kind).toBe("completed");
        expect(seenRunIds).toEqual(["run_1"]);
    });
});
