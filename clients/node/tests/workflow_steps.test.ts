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
    test("list() uses the full steps route", async () => {
        const mockClient = new MockClient();
        const stepsClient = new APIWorkflowRunSteps(mockClient);

        const steps = await stepsClient.list("run_123");

        expect(mockClient.lastFetchParams).toEqual({
            url: "/v1/workflows/runs/run_123/steps",
            method: "GET",
            params: undefined,
            headers: undefined,
        });
        expect(steps).toHaveLength(1);
        expect(steps[0]?.node_id).toBe("extract-1");
        expect(steps[0]?.status).toBe("completed");
    });
});
