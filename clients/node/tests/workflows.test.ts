import { describe, expect, test } from "bun:test";

import { AbstractClient } from "../src/client";
import APIWorkflows from "../src/api/workflows/client";
import APIWorkflowRuns from "../src/api/workflows/runs/client";
import APIWorkflowBlocks from "../src/api/workflows/blocks/client";
import APIWorkflowEdges from "../src/api/workflows/edges/client";
import APIWorkflowSpecs from "../src/api/workflows/specs/client";
import { ZExperimentJobStatus } from "../src/api/workflows/experiments/types";
import { ZStepExecutionResponse, ZWorkflowRun, ZWorkflowRunStep } from "../src/types";
import type { WorkflowRunExportResponse } from "../src/types";

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
        workflow = {
            workflow_id: "wf_1",
            version_id: "ver_0123456789abcdef0123456789abcdef",
            name_at_run_time: "Test",
        },
        trigger = { type: "manual" },
        lifecycle = { status: "running" },
        timing = {
            created_at: "2026-01-01T00:00:00Z",
            started_at: "2026-01-01T00:00:00Z",
        },
        ...rest
    } = overrides;
    return {
        id,
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

    test("update() accepts an email trigger policy", async () => {
        const mockClient = new MockClient({
            id: "workflow_123",
            name: "Test Workflow",
            description: "",
            published: null,
            email_trigger: {
                allowed_senders: ["ops@example.com"],
                allowed_domains: ["example.com"],
            },
            created_at: "2026-03-12T10:00:00Z",
            updated_at: "2026-03-12T10:00:00Z",
        });
        const workflowsClient = new APIWorkflows(mockClient);

        const workflow = await workflowsClient.update("workflow_123", {
            emailTrigger: {
                allowedSenders: ["ops@example.com"],
                allowedDomains: ["example.com"],
            },
        });

        expect(mockClient.lastFetchParams).toEqual({
            url: "/workflows/workflow_123",
            method: "PATCH",
            body: {
                email_trigger: {
                    allowed_senders: ["ops@example.com"],
                    allowed_domains: ["example.com"],
                },
            },
            params: undefined,
            headers: undefined,
        });
        expect(workflow.email_trigger.allowed_senders).toEqual(["ops@example.com"]);
    });

    test("exposes specs subresource", () => {
        const workflowsClient = new APIWorkflows(new MockClient({}));

        expect(workflowsClient.specs).toBeInstanceOf(APIWorkflowSpecs);
    });

    test("specs.validate() uses the spec validate route", async () => {
        const mockClient = new MockClient({
            workflow_id: "wf_1",
            block_count: 2,
            edge_count: 1,
            is_valid: true,
            diagnostics: { issues: [] },
        });
        const specsClient = new APIWorkflowSpecs(mockClient);

        const response = await specsClient.validate("spec: {}\n");

        expect(mockClient.lastFetchParams).toEqual({
            url: "/workflows/spec/validate",
            method: "POST",
            body: { yaml_definition: "spec: {}\n" },
            params: undefined,
            headers: undefined,
        });
        expect(response.is_valid).toBe(true);
    });

    test("specs.plan() uses the spec plan route", async () => {
        const mockClient = new MockClient({
            workflow_id: "wf_1",
            action: "noop",
            block_count: 2,
            edge_count: 1,
            diagnostics: { issues: [] },
            format_version: "workflows-plan/v1",
            summary: { add: 0, change: 1, destroy: 0, replace: 0, noop: 0, total: 1, has_changes: true },
            resource_changes: [
                {
                    address: "workflow.wf_1.block.block_extract",
                    target: "block",
                    target_id: "block_extract",
                    name: "Extract",
                    type: "extract",
                    actions: ["update"],
                    summary: "Update block 'Extract'",
                    change: {
                        before: { config: { model: "old" } },
                        after: { config: { model: "new" } },
                        before_sensitive: {},
                        after_sensitive: {},
                        field_changes: [
                            {
                                path: ["config", "model"],
                                path_display: "config.model",
                                action: "update",
                                before: "old",
                                after: "new",
                            },
                        ],
                    },
                },
            ],
            rendered_plan: "Plan: 0 to add, 1 to change, 0 to destroy.",
        });
        const specsClient = new APIWorkflowSpecs(mockClient);

        const response = await specsClient.plan("spec: {}\n");

        expect(mockClient.lastFetchParams).toEqual({
            url: "/workflows/spec/plan",
            method: "POST",
            body: { yaml_definition: "spec: {}\n" },
            params: undefined,
            headers: undefined,
        });
        expect(response.summary.change).toBe(1);
        expect(response.resource_changes[0].change.field_changes[0].path_display).toBe("config.model");
        expect(response.rendered_plan).toContain("1 to change");
    });

    test("specs.apply() uses the spec apply route", async () => {
        const mockClient = new MockClient({
            workflow_id: "wf_1",
            created: false,
            block_count: 2,
            edge_count: 1,
            diagnostics: { issues: [] },
            format_version: "workflows-plan/v1",
            summary: { add: 0, change: 0, destroy: 0, replace: 0, noop: 1, total: 0, has_changes: false },
            resource_changes: [],
            rendered_plan: "No changes. Infrastructure is up-to-date.",
        });
        const specsClient = new APIWorkflowSpecs(mockClient);

        const response = await specsClient.apply("spec: {}\n");

        expect(mockClient.lastFetchParams).toEqual({
            url: "/workflows/spec/apply",
            method: "POST",
            body: { yaml_definition: "spec: {}\n" },
            params: undefined,
            headers: undefined,
        });
        expect(response.workflow_id).toBe("wf_1");
        expect(response.summary.noop).toBe(1);
        expect(response.resource_changes).toEqual([]);
    });

    test("specs.export() uses the spec export route", async () => {
        const mockClient = new MockClient({
            workflow_id: "wf_1",
            yaml_definition: "apiVersion: workflows.retab.com/v1alpha2\n",
        });
        const specsClient = new APIWorkflowSpecs(mockClient);

        const response = await specsClient.export("wf_1");

        expect(mockClient.lastFetchParams).toEqual({
            url: "/workflows/spec/wf_1",
            method: "GET",
            params: undefined,
            headers: undefined,
        });
        expect(response.yaml_definition.startsWith("apiVersion:")).toBe(true);
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
                version_id: "ver_0123456789abcdef0123456789abcdef",
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
        expect(wf.published?.version_id).toBe("ver_0123456789abcdef0123456789abcdef");
    });

    test("getEntities() returns blocks and edges", async () => {
        const mockClient = new MockClient({
            workflow: {
                id: "wf_1", name: "Test",
                created_at: "2026-01-01T00:00:00Z", updated_at: "2026-01-01T00:00:00Z",
            },
            blocks: [
                { id: "start-1", workflow_id: "wf_1", draft_version: "draft_1", type: "start", label: "Doc Input" },
                { id: "extract-1", workflow_id: "wf_1", draft_version: "draft_1", type: "extract", label: "Extract" },
            ],
            edges: [
                {
                    id: "edge-1",
                    workflow_id: "wf_1",
                    draft_version: "draft_1",
                    source_block: "start-1",
                    target_block: "extract-1",
                },
            ],
        });
        const workflowsClient = new APIWorkflows(mockClient);

        const entities = await workflowsClient.getEntities("wf_1");

        expect(mockClient.lastFetchParams?.url).toBe("/workflows/wf_1/entities");
        expect(entities.blocks).toHaveLength(2);
        expect(entities.edges).toHaveLength(1);
        expect(entities.edges[0]?.draft_version).toBe("draft_1");
        expect(entities.blocks.filter(b => b.type === "start")).toHaveLength(1);
    });

    test("prepare_diagnose exposes the Python prepared-request surface", () => {
        const workflowsClient = new APIWorkflows(new MockClient({}));
        const blocks = [{ id: "start-1", type: "start" }];
        const edges = [{ id: "edge-1", source: "start-1", target: "extract-1" }];

        expect(workflowsClient.prepare_diagnose("wf_1", blocks, edges, false)).toEqual({
            url: "/workflows/wf_1/diagnose-graph",
            method: "POST",
            body: {
                blocks,
                edges,
                re_propagate: false,
            },
        });
        expect(workflowsClient.prepare_diagnose("wf_1", blocks, edges).body.re_propagate).toBe(true);
    });

    test("artifacts prepare helpers expose the Python prepared-request surface", () => {
        const workflowsClient = new APIWorkflows(new MockClient({}));

        expect(workflowsClient.artifacts.prepare_get("hil_evaluation", "artifact_1")).toEqual({
            url: "/workflows/artifacts/hil_evaluation/artifact_1",
            method: "GET",
        });
        expect(workflowsClient.artifacts.prepare_get({ operation: "extraction", id: "ext_1" })).toEqual({
            url: "/workflows/artifacts/extraction/ext_1",
            method: "GET",
        });
        expect(workflowsClient.artifacts.prepare_list("run_1", "extraction", "extract-1")).toEqual({
            url: "/workflows/artifacts",
            method: "GET",
            params: {
                run_id: "run_1",
                operation: "extraction",
                block_id: "extract-1",
            },
        });
    });

    test("experiments expose Python snake_case and prepare helpers", () => {
        const workflowsClient = new APIWorkflows(new MockClient({}));

        expect("get_content" in workflowsClient.experiments).toBe(false);
        expect("getContent" in workflowsClient.experiments).toBe(false);
        expect(typeof workflowsClient.experiments.get_metrics).toBe("function");
        expect(typeof workflowsClient.experiments.list_eligible_blocks).toBe("function");
        expect(typeof workflowsClient.experiments.run_batch).toBe("function");
        expect(typeof workflowsClient.experiments.runs.get).toBe("function");
        expect("get_content" in workflowsClient.experiments.runs).toBe(false);
        expect("getContent" in workflowsClient.experiments.runs).toBe(false);
        expect("get_job" in workflowsClient.experiments.runs).toBe(false);
        expect("getJob" in workflowsClient.experiments.runs).toBe(false);
        expect("wait_for_completion" in workflowsClient.experiments.runs).toBe(false);
        expect("waitForCompletion" in workflowsClient.experiments.runs).toBe(false);
        expect(workflowsClient.experiments.runs.prepare_get("wf_1", "exp_1", { runId: "run_1" })).toEqual({
            url: "/workflows/wf_1/experiments/exp_1/content",
            method: "GET",
            params: { run_id: "run_1" },
        });
        expect(workflowsClient.experiments.runs.prepare_create("wf_1", "exp_1")).toEqual({
            url: "/workflows/wf_1/experiments/exp_1/run",
            method: "POST",
            body: {},
        });
        expect("prepare_run_document" in workflowsClient.experiments.runs).toBe(false);
        expect("run_document" in workflowsClient.experiments.runs).toBe(false);
        expect(workflowsClient.experiments.runs.prepare_cancel_document("wf_1", "exp_1", "expdoc_1")).toEqual({
            url: "/workflows/wf_1/experiments/exp_1/documents/expdoc_1/cancel",
            method: "POST",
            body: {},
        });
        expect(workflowsClient.experiments.prepare_get_metrics("wf_1", "exp_1").params).toEqual({
            view: "summary",
            include_prior: true,
        });
        expect(ZExperimentJobStatus.safeParse("completed").success).toBe(true);
        expect(ZExperimentJobStatus.safeParse("cancelled").success).toBe(false);
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
            workflow: {
                workflow_id: "workflow_123",
                version_id: "ver_0123456789abcdef0123456789abcdef",
                name_at_run_time: "Classifier Workflow",
            },
            trigger: { type: "manual" },
            lifecycle: { status: "running" },
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
        expect(run.lifecycle.status).toBe("running");
    });

    test("runs.v2 typed fields parse correctly", async () => {
        const mockClient = new MockClient({
            id: "run_789",
            workflow: {
                workflow_id: "wf_1",
                version_id: "ver_abcdef0123456789abcdef0123456789",
                name_at_run_time: "Test",
                requested_version: "ver_abcdef0123456789abcdef0123456789",
            },
            trigger: { type: "api", api_key_id: "ak_1" },
            lifecycle: { status: "completed" },
            timing: {
                created_at: "2026-01-01T00:00:00Z",
                started_at: "2026-01-01T00:00:00Z",
                completed_at: "2026-01-01T00:00:05Z",
                accumulated_human_waiting_ms: 5000,
            },
            inputs: { documents: {}, json_data: {} },
        });
        const runsClient = new APIWorkflowRuns(mockClient);

        const run = await runsClient.get("run_789");

        expect(run.trigger.type).toBe("api");
        expect(run.workflow.workflow_id).toBe("wf_1");
        expect(run.workflow.version_id).toBe("ver_abcdef0123456789abcdef0123456789");
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
            { id: "start-1", workflow_id: "wf_1", draft_version: "draft_1", type: "start" },
            { id: "extract-1", workflow_id: "wf_1", draft_version: "draft_1", type: "extract" },
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

    test("blocks.prepare_simulate exposes the Python prepared-request surface", () => {
        const blocksClient = new APIWorkflowBlocks(new MockClient({}));

        expect(blocksClient.prepare_simulate("run_1", "extract-1", 5, "step_1", false)).toEqual({
            url: "/workflows/runs/run_1/steps/extract-1/simulate",
            method: "POST",
            params: {
                n_consensus: 5,
                step_id: "step_1",
                check_eligibility: false,
            },
        });
        expect(blocksClient.prepare_simulate("run_1", "extract-1")).toEqual({
            url: "/workflows/runs/run_1/steps/extract-1/simulate",
            method: "POST",
        });
    });

    test("workflow blocks expose live-editing metadata", async () => {
        const mockClient = new MockClient({
            id: "extract-1",
            workflow_id: "wf_1",
            draft_version: "draft_1",
            type: "extract",
            label: "Extract",
        });
        const blocksClient = new APIWorkflowBlocks(mockClient);

        const block = await blocksClient.get("wf_1", "extract-1");

        expect(block.draft_version).toBe("draft_1");
    });

    test("workflow block objects ignore resolved schema sidecars", async () => {
        const mockClient = new MockClient({
            id: "extract-1",
            workflow_id: "wf_1",
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

        expect("resolved_schemas" in block).toBe(false);
    });

    test("workflow resolved schema endpoints parse public field_ref_drift", async () => {
        const workflowSchemasClient = new MockClient({
            workflow_id: "wf_1",
            draft_version: "draft_1",
            schemas: {
                "extract-1": {
                    input_schemas: {},
                    output_schemas: {
                        "output-json-0": {
                            type: "object",
                            properties: { invoice_number: { type: "string" } },
                        },
                    },
                    field_ref_drift: { stale: true },
                },
            },
        });
        const workflowsClient = new APIWorkflows(workflowSchemasClient);
        const workflowSchemas = await workflowsClient.getResolvedSchemas("wf_1");

        expect(workflowSchemasClient.lastFetchParams?.url).toBe("/workflows/wf_1/resolved-schemas");
        expect(workflowSchemas.schemas["extract-1"].field_ref_drift?.stale).toBe(true);
        expect(workflowSchemas.schemas["extract-1"].output_schemas["output-json-0"]?.properties.invoice_number.type).toBe("string");

        const blockSchemasClient = new MockClient({
            workflow_id: "wf_1",
            block_id: "extract-1",
            draft_version: "draft_1",
            schema: {
                input_schemas: {},
                output_schemas: {"output-json-0": {type: "object"}},
                field_ref_drift: null,
            },
        });
        const blocksClient = new APIWorkflowBlocks(blockSchemasClient);
        const blockSchemas = await blocksClient.getResolvedSchemas("wf_1", "extract-1");

        expect(blockSchemasClient.lastFetchParams?.url).toBe("/workflows/wf_1/blocks/extract-1/resolved-schemas");
        expect(blockSchemas.schema.field_ref_drift).toBeNull();
    });

    test("step execution responses ignore removed payload schemas", () => {
        const removedPayloadKey = ["raw", "_", "output"].join("");
        const parsed = ZStepExecutionResponse.parse({
            block_id: "extract-1",
            block_type: "extract",
            block_label: "Extract",
            lifecycle: { status: "completed" },
            output: {
                data: { invoice_number: "INV-001" },
                json_schema: { type: "object" },
            },
            [removedPayloadKey]: {
                json_schema: { type: "object" },
            },
            json_schema: { type: "object" },
            handle_outputs: {
                "output-json-0": {
                    type: "json",
                    data: { invoice_number: "INV-001" },
                },
            },
            handle_inputs: {},
        });

        expect("output" in parsed).toBe(false);
        expect(removedPayloadKey in parsed).toBe(false);
        expect("json_schema" in parsed).toBe(false);
        expect("artifact_view" in parsed).toBe(false);
        expect("status" in parsed).toBe(false);
        expect("terminal" in parsed).toBe(false);
        expect(parsed.handle_outputs?.["output-json-0"]?.data?.invoice_number).toBe("INV-001");
    });

    test("workflow run steps expose top-level model", () => {
        const parsed = ZWorkflowRunStep.parse({
            run_id: "run_123",
            block_id: "extract-1",
            step_id: "extract-1",
            block_type: "extract",
            block_label: "Extract",
            lifecycle: { status: "completed" },
            model: "retab-small",
            handle_outputs: {},
            handle_inputs: {},
        });

        expect(parsed.model).toBe("retab-small");
        expect("status" in parsed).toBe(false);
        expect("terminal" in parsed).toBe(false);
    });

    test("edges.createBatch() accepts camelCase request objects", async () => {
        const mockClient = new MockClient([
            {
                id: "edge-1",
                workflow_id: "wf_1",
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
            run: makeV2Run({ lifecycle: { status: "cancelled" } }),
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
            makeV2Run({ id: "run_2", lifecycle: { status: "running" } })
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

    test("runs.getAgentHilReview() sends GET to /agent-hil-reviews/{blockId}", async () => {
        const mockClient = new MockClient({
            id: "agrev_abc",
            organization_id: "org_1",
            run_id: "run_1",
            block_id: "hil-1",
            workflow_id: "wf_1",
            mode: "pre_review",
            status: "proposed",
            managed_agent_session_id: "sesn_1",
            managed_agent_vault_id: "vlt_1",
            proposed_decision: {
                approved: true,
                modified_data: { total: 325 },
                confidence: 0.92,
                evidence: [
                    {
                        field_path: "total",
                        action: "modified",
                        quote: "Total $325.00",
                        source: { document_index: 0, page_number: 1 },
                        from_value: 505,
                        to_value: 325,
                    },
                ],
                changed_paths: ["total"],
                escalate: false,
            },
            submitted_hil_command_id: null,
            failure_reason: null,
            auto_threshold: 0.85,
            timeout_seconds: 900,
            created_at: "2026-01-01T00:00:00Z",
            updated_at: "2026-01-01T00:00:01Z",
        });
        const runsClient = new APIWorkflowRuns(mockClient);

        const result = await runsClient.getAgentHilReview("run_1", "hil-1");

        expect(mockClient.lastFetchParams).toEqual({
            url: "/workflows/runs/run_1/agent-hil-reviews/hil-1",
            method: "GET",
            params: undefined,
            headers: undefined,
        });
        expect(result.id).toBe("agrev_abc");
        expect(result.mode).toBe("pre_review");
        expect(result.status).toBe("proposed");
        expect(result.proposed_decision?.confidence).toBe(0.92);
        expect(result.proposed_decision?.changed_paths).toEqual(["total"]);
        expect(result.proposed_decision?.evidence[0]?.from_value).toBe(505);
        expect(result.proposed_decision?.evidence[0]?.to_value).toBe(325);
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

describe("workflow runs", () => {
    test("workflow runs do not expose wait helpers", () => {
        const runsClient = new APIWorkflowRuns(new MockClient({}));

        expect("waitForCompletion" in runsClient).toBe(false);
        expect("wait_for_completion" in runsClient).toBe(false);
        expect("createAndWait" in runsClient).toBe(false);
    });
});
