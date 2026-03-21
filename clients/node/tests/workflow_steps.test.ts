import { describe, expect, test } from "bun:test";

import { AbstractClient } from "../src/client";
import APIWorkflowRunSteps from "../src/api/workflows/runs/steps/client";

class MockClient extends AbstractClient {
    public lastFetchParams: Record<string, unknown> | null = null;

    protected async _fetch(params: {
        url: string;
        method: string;
        params?: Record<string, unknown>;
        headers?: Record<string, unknown>;
        body?: Record<string, unknown>;
    }): Promise<Response> {
        this.lastFetchParams = params;
        return new Response(JSON.stringify([
            {
                run_id: "run_123",
                organization_id: "org_123",
                node_id: "extract-1",
                step_id: "extract-1",
                node_type: "extract",
                node_label: "Extract",
                status: "completed",
            },
        ]), {
            status: 200,
            headers: { "Content-Type": "application/json" },
        });
    }
}

describe("workflow run steps client", () => {
    test("get() uses the public step route without raw output", async () => {
        class GetMockClient extends AbstractClient {
            public lastFetchParams: Record<string, unknown> | null = null;

            protected async _fetch(params: {
                url: string;
                method: string;
                params?: Record<string, unknown>;
                headers?: Record<string, unknown>;
                body?: Record<string, unknown>;
            }): Promise<Response> {
                this.lastFetchParams = params;
                return new Response(JSON.stringify({
                    node_id: "extract-1",
                    node_type: "extract",
                    node_label: "Extract",
                    status: "completed",
                    handle_outputs: {
                        "output-json-0": {
                            type: "json",
                            data: { invoice_number: "INV-001" },
                        },
                    },
                    handle_inputs: null,
                }), {
                    status: 200,
                    headers: { "Content-Type": "application/json" },
                });
            }
        }

        const mockClient = new GetMockClient();
        const stepsClient = new APIWorkflowRunSteps(mockClient);

        const step = await stepsClient.get("run_123", "extract-1");

        expect(mockClient.lastFetchParams).toEqual({
            url: "/workflows/runs/run_123/steps/extract-1",
            method: "GET",
            params: undefined,
            headers: undefined,
        });
        expect(step.node_id).toBe("extract-1");
        expect("output" in step).toBe(false);
        expect(step.handle_outputs?.["output-json-0"]).toBeDefined();
    });

    test("list() uses the full steps route", async () => {
        const mockClient = new MockClient();
        const stepsClient = new APIWorkflowRunSteps(mockClient);

        const steps = await stepsClient.list("run_123");

        expect(mockClient.lastFetchParams).toEqual({
            url: "/workflows/runs/run_123/steps",
            method: "GET",
            params: undefined,
            headers: undefined,
        });
        expect(steps).toHaveLength(1);
        expect(steps[0]?.node_id).toBe("extract-1");
        expect(steps[0]?.status).toBe("completed");
        expect(steps[0] && "output" in steps[0]).toBe(false);
    });

    test("getMany() sends POST to /steps/batch", async () => {
        class BatchMockClient extends AbstractClient {
            public lastFetchParams: Record<string, unknown> | null = null;

            protected async _fetch(params: {
                url: string;
                method: string;
                params?: Record<string, unknown>;
                headers?: Record<string, unknown>;
                body?: unknown;
            }): Promise<Response> {
                this.lastFetchParams = params;
                return new Response(JSON.stringify({
                    outputs: {
                        "extract-1": {
                            node_id: "extract-1",
                            node_type: "extract",
                            node_label: "Extract",
                            status: "completed",
                            handle_outputs: {
                                "output-json-0": { type: "json", data: { field: "value" } },
                            },
                            handle_inputs: null,
                        },
                    },
                }), {
                    status: 200,
                    headers: { "Content-Type": "application/json" },
                });
            }
        }

        const mockClient = new BatchMockClient();
        const stepsClient = new APIWorkflowRunSteps(mockClient);

        const batch = await stepsClient.getMany("run_123", ["extract-1"]);

        expect(mockClient.lastFetchParams?.url).toBe("/workflows/runs/run_123/steps/batch");
        expect(mockClient.lastFetchParams?.method).toBe("POST");
        expect(batch.outputs["extract-1"]?.node_id).toBe("extract-1");
    });
});
