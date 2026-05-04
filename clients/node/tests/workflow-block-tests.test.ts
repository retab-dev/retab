import { describe, expect, test } from "bun:test";

import { AbstractClient } from "../src/client";
import APIWorkflows from "../src/api/workflows/client";
import APIWorkflowTests from "../src/api/workflows/tests/client";
import APIWorkflowTestRuns from "../src/api/workflows/tests/runs/client";

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

const TEST_RESPONSE = {
    id: "wfnodetest_abc",
    workflow_id: "wf_abc123",
    organization_id: "org_x",
    target: { type: "block", block_id: "block_extract" },
    source: { type: "manual", handle_inputs: {} },
    name: "Q1 invoice total",
    assertion: {
        id: "assert_xyz",
        target: { output_handle_id: "output-json-0", path: "total" },
        condition: { kind: "equals", expected: 1234.56 },
        label: null,
    },
    schema_drift: "fresh",
    validation_status: "valid",
    validation_issues: [],
    latest_run_summary: null,
    latest_passing_run_summary: null,
    latest_failing_run_summary: null,
    created_at: NOW,
    updated_at: NOW,
};

const RUN_RESPONSE = {
    id: "wfnodetestrun_abc",
    test_id: "wfnodetest_abc",
    workflow_id: "wf_abc123",
    organization_id: "org_x",
    target: { type: "block", block_id: "block_extract" },
    source: { type: "manual", handle_inputs: {} },
    status: "passed",
    execution_fingerprint: "f1",
    started_at: NOW,
    completed_at: NOW,
    duration_ms: 1234,
    outputs: { "output-json-0": { total: 1234.56 } },
    warnings: [],
    skipped: false,
};

const EXECUTE_RESPONSE = {
    batch_id: "btbatch_q1z2",
    job_id: "job_abc",
    status: "queued",
    workflow_id: "wf_abc123",
    target: { type: "block", block_id: "block_extract" },
    test_id: null,
    total_tests: 4,
};

describe("workflows.tests wiring", () => {
    test("APIWorkflows exposes a `tests` sub-resource", () => {
        // Without this wire, the docs example
        // `client.workflows.tests.create(...)` raises a TypeError at
        // call time. Pin the wire so a future refactor can't drop it
        // silently.
        const workflows = new APIWorkflows(new MockClient({}));
        expect(workflows.tests).toBeInstanceOf(APIWorkflowTests);
    });

    test("APIWorkflowTests exposes a `runs` sub-resource", () => {
        const tests = new APIWorkflowTests(new MockClient({}));
        expect(tests.runs).toBeInstanceOf(APIWorkflowTestRuns);
    });
});

describe("workflows.tests.create()", () => {
    test("POSTs to the block-tests route with target/source/assertion/name in the body", async () => {
        const mockClient = new MockClient(TEST_RESPONSE);
        const tests = new APIWorkflowTests(mockClient);

        const test = await tests.create({
            workflowId: "wf_abc123",
            target: { type: "block", block_id: "block_extract" },
            source: { type: "manual", handle_inputs: {} },
            assertion: {
                target: { output_handle_id: "output-json-0", path: "total" },
                condition: { kind: "equals", expected: 1234.56 },
            },
            name: "Q1 invoice total",
        });

        expect(mockClient.lastFetchParams).toMatchObject({
            url: "/workflows/wf_abc123/block-tests",
            method: "POST",
            body: {
                target: { type: "block", block_id: "block_extract" },
                source: { type: "manual", handle_inputs: {} },
                assertion: {
                    target: { output_handle_id: "output-json-0", path: "total" },
                    condition: { kind: "equals", expected: 1234.56 },
                },
                name: "Q1 invoice total",
            },
        });
        expect(test.id).toBe("wfnodetest_abc");
    });

    test("omits `name` from the body when the caller doesn't supply it", async () => {
        const mockClient = new MockClient(TEST_RESPONSE);
        const tests = new APIWorkflowTests(mockClient);

        await tests.create({
            workflowId: "wf_abc123",
            target: { type: "block", block_id: "block_extract" },
            source: {
                type: "run_step",
                run_id: "wfrun_xyz",
                step_id: "block_extract",
            },
            assertion: {
                target: { output_handle_id: "output-json-0", path: "total" },
                condition: { kind: "exists" },
            },
        });

        const body = (mockClient.lastFetchParams as { body: Record<string, unknown> }).body;
        expect("name" in body).toBe(false);
        expect(body.source).toEqual({
            type: "run_step",
            run_id: "wfrun_xyz",
            step_id: "block_extract",
        });
    });
});

describe("workflows.tests CRUD URL shapes", () => {
    test("get() uses the test detail route", async () => {
        const mockClient = new MockClient(TEST_RESPONSE);
        const tests = new APIWorkflowTests(mockClient);

        await tests.get({ workflowId: "wf_abc123", testId: "wfnodetest_abc" });

        expect(mockClient.lastFetchParams).toMatchObject({
            url: "/workflows/wf_abc123/block-tests/wfnodetest_abc",
            method: "GET",
        });
    });

    test("list() uses the block-tests route with `target_block_id` filter", async () => {
        const mockClient = new MockClient({ tests: [] });
        const tests = new APIWorkflowTests(mockClient);

        await tests.list({
            workflowId: "wf_abc123",
            targetBlockId: "block_extract",
            limit: 25,
        });

        expect(mockClient.lastFetchParams).toMatchObject({
            url: "/workflows/wf_abc123/block-tests",
            method: "GET",
            params: { limit: 25, target_block_id: "block_extract" },
        });
    });

    test("list() omits target_block_id when undefined", async () => {
        // Passing `target_block_id: undefined` would have the backend
        // string-coerce to the literal "undefined" — make sure the SDK
        // omits the param entirely.
        const mockClient = new MockClient({ tests: [] });
        const tests = new APIWorkflowTests(mockClient);

        await tests.list({ workflowId: "wf_abc123" });

        const params = (mockClient.lastFetchParams as { params: Record<string, unknown> }).params;
        expect("target_block_id" in params).toBe(false);
    });

    test("delete() uses the test detail route with DELETE", async () => {
        // Body shape from the backend is empty (204) — the SDK doesn't
        // try to parse it.
        const mockClient = new MockClient(null);
        const tests = new APIWorkflowTests(mockClient);

        await tests.delete({
            workflowId: "wf_abc123",
            testId: "wfnodetest_abc",
        });

        expect(mockClient.lastFetchParams).toMatchObject({
            url: "/workflows/wf_abc123/block-tests/wfnodetest_abc",
            method: "DELETE",
        });
    });
});

describe("workflows.tests.update()", () => {
    test("PATCH body only carries the fields the caller passed", async () => {
        // Including a field with `undefined` would either trip Pydantic
        // validation or — worse — clear the field on the stored doc.
        // The SDK MUST omit, not nullify.
        const mockClient = new MockClient(TEST_RESPONSE);
        const tests = new APIWorkflowTests(mockClient);

        await tests.update({
            workflowId: "wf_abc123",
            testId: "wfnodetest_abc",
            name: "renamed",
        });

        expect(mockClient.lastFetchParams).toMatchObject({
            url: "/workflows/wf_abc123/block-tests/wfnodetest_abc",
            method: "PATCH",
            body: { name: "renamed" },
        });
        const body = (mockClient.lastFetchParams as { body: Record<string, unknown> }).body;
        expect(Object.keys(body)).toEqual(["name"]);
    });

    test("PATCH with assertion only includes assertion in the body", async () => {
        const mockClient = new MockClient(TEST_RESPONSE);
        const tests = new APIWorkflowTests(mockClient);

        await tests.update({
            workflowId: "wf_abc123",
            testId: "wfnodetest_abc",
            assertion: {
                target: { output_handle_id: "output-json-0", path: "vendor.name" },
                condition: { kind: "matches_regex", expected: "^Acme.*" },
            },
        });

        const body = (mockClient.lastFetchParams as { body: Record<string, unknown> }).body;
        expect(Object.keys(body)).toEqual(["assertion"]);
        expect(body.assertion).toMatchObject({
            condition: { kind: "matches_regex" },
        });
    });
});

describe("workflows.tests.execute()", () => {
    test("execute({testId}) posts only test_id", async () => {
        const mockClient = new MockClient(EXECUTE_RESPONSE);
        const tests = new APIWorkflowTests(mockClient);

        const response = await tests.execute({
            workflowId: "wf_abc123",
            testId: "wfnodetest_abc",
        });

        expect(mockClient.lastFetchParams).toMatchObject({
            url: "/workflows/wf_abc123/block-tests/execute",
            method: "POST",
            body: { test_id: "wfnodetest_abc" },
        });
        expect(response.status).toBe("queued");
    });

    test("execute({target, nConsensus}) posts target + n_consensus snake_case", async () => {
        const mockClient = new MockClient(EXECUTE_RESPONSE);
        const tests = new APIWorkflowTests(mockClient);

        await tests.execute({
            workflowId: "wf_abc123",
            target: { type: "block", block_id: "block_extract" },
            nConsensus: 5,
        });

        expect(mockClient.lastFetchParams).toMatchObject({
            body: {
                target: { type: "block", block_id: "block_extract" },
                n_consensus: 5,
            },
        });
    });

    test("execute({}) posts an empty body to run every test in the workflow", async () => {
        const mockClient = new MockClient(EXECUTE_RESPONSE);
        const tests = new APIWorkflowTests(mockClient);

        await tests.execute({ workflowId: "wf_abc123" });

        const body = (mockClient.lastFetchParams as { body: Record<string, unknown> }).body;
        expect(body).toEqual({});
    });
});

describe("workflows.tests.runs", () => {
    test("runs.list() uses the test runs route", async () => {
        const mockClient = new MockClient({ runs: [] });
        const tests = new APIWorkflowTests(mockClient);

        await tests.runs.list({
            workflowId: "wf_abc123",
            testId: "wfnodetest_abc",
            limit: 10,
        });

        expect(mockClient.lastFetchParams).toMatchObject({
            url: "/workflows/wf_abc123/block-tests/wfnodetest_abc/runs",
            method: "GET",
            params: { limit: 10 },
        });
    });

    test("runs.get() uses the run detail route and parses `outputs`", async () => {
        const mockClient = new MockClient(RUN_RESPONSE);
        const tests = new APIWorkflowTests(mockClient);

        const run = await tests.runs.get({
            workflowId: "wf_abc123",
            testId: "wfnodetest_abc",
            runId: "wfnodetestrun_abc",
        });

        expect(mockClient.lastFetchParams).toMatchObject({
            url: "/workflows/wf_abc123/block-tests/wfnodetest_abc/runs/wfnodetestrun_abc",
            method: "GET",
        });
        // The renamed `outputs` field (formerly `handle_outputs`) parses cleanly.
        expect(run.outputs).toEqual({ "output-json-0": { total: 1234.56 } });
        expect(run.status).toBe("passed");
    });
});

describe("status enum coverage", () => {
    test("Z* schemas accept all 7 BlockTestRunStatus values", async () => {
        // Pin: a future status addition (or accidental shrinkage) trips
        // this test before it ships. Without this, a `cancelled` run
        // record from the backend would runtime-fail in the SDK.
        const { ZBlockTestRunStatus } = await import(
            "../src/api/workflows/tests/types"
        );
        for (const value of [
            "queued",
            "running",
            "passed",
            "failed",
            "blocked",
            "error",
            "cancelled",
        ]) {
            expect(() => ZBlockTestRunStatus.parse(value)).not.toThrow();
        }
    });

    test("Z* batch counts schema declares all 7 status buckets", async () => {
        const { ZBlockTestBatchExecutionCounts } = await import(
            "../src/api/workflows/tests/types"
        );
        const parsed = ZBlockTestBatchExecutionCounts.parse({});
        for (const key of [
            "queued",
            "running",
            "passed",
            "failed",
            "blocked",
            "error",
            "cancelled",
        ]) {
            expect(parsed).toHaveProperty(key, 0);
        }
    });
});


// ---------------------------------------------------------------------------
// Round 2 — gap coverage caught in second-pass review
// ---------------------------------------------------------------------------


describe("workflows.tests parity bridge installation", () => {
    test("APIWorkflowTests has all 6 prepare_* aliases installed at runtime", () => {
        // The parity bridge auto-installs `prepare_create`, `prepare_get`, etc.
        // from `PYTHON_PUBLIC_PREPARE_METHODS` in `src/client.ts`. A typo in
        // either the registry key or the method name would silently disable
        // the bridge — pin the actual installed methods so a regression
        // breaks here, not in the python_parity test (which fails fast on
        // the FIRST violation and is harder to debug).
        const tests = new APIWorkflowTests(new MockClient({})) as unknown as Record<
            string,
            unknown
        >;
        for (const name of [
            "prepare_create",
            "prepare_get",
            "prepare_list",
            "prepare_update",
            "prepare_delete",
            "prepare_execute",
        ]) {
            expect(typeof tests[name]).toBe("function");
        }
    });

    test("APIWorkflowTestRuns has both prepare_* aliases installed", () => {
        const runs = new APIWorkflowTestRuns(new MockClient({})) as unknown as Record<
            string,
            unknown
        >;
        expect(typeof runs.prepare_list).toBe("function");
        expect(typeof runs.prepare_get).toBe("function");
    });
});


describe("workflows.tests response parsing edge cases", () => {
    test("ZWorkflowTest accepts `assertion: null` from a legacy doc", async () => {
        // Pre-rewrite tests in storage may have `assertion=None`. The SDK's
        // schema declares the field as `.nullable().optional()` to mirror
        // the backend's relaxed `WorkflowTest.assertion: AssertionSpec | None
        // = None`. Without this, listing a workflow with even one orphan
        // would crash the client at parse time.
        const { ZWorkflowTest } = await import(
            "../src/api/workflows/tests/types"
        );
        const legacy = { ...TEST_RESPONSE, assertion: null };
        const parsed = ZWorkflowTest.parse(legacy);
        expect(parsed.assertion).toBeNull();
    });

    test("runs.list() correctly parses a populated runs[] array", async () => {
        // The empty-array path is covered above. Pin the populated path so
        // a future schema drift (renaming `outputs` again, dropping
        // `verdict_summary`, etc.) breaks here — not in production with a
        // real run-history fetch.
        const populatedRunsResponse = {
            runs: [
                RUN_RESPONSE,
                { ...RUN_RESPONSE, id: "wfnodetestrun_b", status: "blocked" },
                {
                    ...RUN_RESPONSE,
                    id: "wfnodetestrun_c",
                    status: "cancelled",
                    outputs: null,
                },
            ],
        };
        const mockClient = new MockClient(populatedRunsResponse);
        const tests = new APIWorkflowTests(mockClient);

        const result = await tests.runs.list({
            workflowId: "wf_abc123",
            testId: "wfnodetest_abc",
        });

        expect(result.runs).toHaveLength(3);
        expect(result.runs[0]?.status).toBe("passed");
        expect(result.runs[1]?.status).toBe("blocked");
        expect(result.runs[2]?.status).toBe("cancelled");
        // The renamed `outputs` field (formerly `handle_outputs`) parses on
        // the populated record and tolerates `null` on the cancelled one.
        expect(result.runs[0]?.outputs).toEqual({ "output-json-0": { total: 1234.56 } });
        expect(result.runs[2]?.outputs).toBeNull();
    });
});


describe("workflows.tests param-merge precedence (sibling parity)", () => {
    test("list() lets caller-supplied options.params override the typed `limit`", async () => {
        // Matches APIWorkflowRuns.list precedence — the typed kwarg is the
        // default, but the `options.params` escape hatch wins. Allows
        // power users to pass through query params without forking the
        // SDK signature.
        const mockClient = new MockClient({ tests: [] });
        const tests = new APIWorkflowTests(mockClient);

        await tests.list(
            { workflowId: "wf_abc123", limit: 50 },
            { params: { limit: 100 } },
        );

        const params = (mockClient.lastFetchParams as { params: Record<string, unknown> })
            .params;
        expect(params.limit).toBe(100);
    });

    test("runs.list() also lets options.params override the typed `limit`", async () => {
        const mockClient = new MockClient({ runs: [] });
        const tests = new APIWorkflowTests(mockClient);

        await tests.runs.list(
            { workflowId: "wf_abc123", testId: "wfnodetest_abc", limit: 10 },
            { params: { limit: 25 } },
        );

        const params = (mockClient.lastFetchParams as { params: Record<string, unknown> })
            .params;
        expect(params.limit).toBe(25);
    });
});
