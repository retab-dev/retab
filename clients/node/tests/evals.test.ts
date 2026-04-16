import { fileURLToPath } from "node:url";

import { describe, expect, test } from "bun:test";

import APIV1 from "../src/api/client.js";
import { AbstractClient } from "../src/client.js";

type RecordedRequest = {
    url: string;
    method: string;
    params?: Record<string, unknown>;
    headers?: Record<string, unknown>;
    bodyMime?: "application/json" | "multipart/form-data";
    body?: Record<string, unknown>;
};

class MockClient extends AbstractClient {
    public requests: RecordedRequest[] = [];

    protected async _fetch(params: RecordedRequest): Promise<Response> {
        this.requests.push(params);

        if (params.url.endsWith("/stream")) {
            const chunks = [
                JSON.stringify({
                    id: "chatcmpl-1",
                    object: "chat.completion.chunk",
                    created: 1710000000,
                    model: "retab-small",
                    choices: [
                        {
                            index: 0,
                            finish_reason: null,
                            delta: {
                                role: "assistant",
                                content: '{"status":"stream"}',
                                flat_parsed: { status: "stream" },
                                flat_likelihoods: {},
                                flat_deleted_keys: [],
                                is_valid_json: true,
                            },
                        },
                    ],
                    extraction_id: "ext-1",
                }),
            ].join("");

            return new Response(chunks, {
                status: 200,
                headers: { "Content-Type": "application/stream+json" },
            });
        }

        return new Response(JSON.stringify(buildResponse(params)), {
            status: 200,
            headers: {
                "Content-Type": "application/json",
            },
        });
    }
}

function buildResponse(params: RecordedRequest): unknown {
    if (params.url.startsWith("/evals/extract/extract/")) {
        return {
            id: "chatcmpl-1",
            object: "chat.completion",
            created: 1710000000,
            model: "retab-small",
            choices: [
                {
                    index: 0,
                    finish_reason: "stop",
                    message: {
                        role: "assistant",
                        content: '{"status":"extract"}',
                        parsed: { status: "extract" },
                    },
                },
            ],
            extraction_id: "ext-1",
        };
    }

    if (params.url.startsWith("/evals/split/extract/")) {
        return {
            id: "chatcmpl-2",
            object: "chat.completion",
            created: 1710000000,
            model: "retab-small",
            choices: [
                {
                    index: 0,
                    finish_reason: "stop",
                    message: {
                        role: "assistant",
                        content: '{"status":"split"}',
                        parsed: { status: "split" },
                    },
                },
            ],
            extraction_id: "ext-2",
        };
    }

    if (params.url.startsWith("/evals/classify/extract/")) {
        return {
            classification: {
                reasoning: "Detected invoice keywords",
                category: "invoice",
            },
            consensus: {
                choices: [],
                likelihood: null,
            },
            usage: {
                credits: 0.12,
            },
        };
    }

    return {};
}

const payslipPath = fileURLToPath(new URL("./data/payslip/payslip.pdf", import.meta.url));

describe("Node SDK evals", () => {
    test("exposes only eval processing resources", () => {
        const api = new APIV1(new MockClient());

        expect(api.evals.extract).toBeDefined();
        expect(api.evals.split).toBeDefined();
        expect(api.evals.classify).toBeDefined();

        expect((api.evals.extract as Record<string, unknown>).datasets).toBeUndefined();
        expect((api.evals.extract as Record<string, unknown>).templates).toBeUndefined();
        expect((api.evals.extract as Record<string, unknown>).create).toBeUndefined();
        expect((api.evals.extract as Record<string, unknown>).extract).toBeUndefined();
        expect(typeof api.evals.extract.process).toBe("function");
        expect(typeof api.evals.extract.processStream).toBe("function");

        expect((api.evals.split as Record<string, unknown>).datasets).toBeUndefined();
        expect((api.evals.split as Record<string, unknown>).templates).toBeUndefined();
        expect((api.evals.split as Record<string, unknown>).create).toBeUndefined();
        expect(typeof api.evals.split.process).toBe("function");
        expect((api.evals.split as Record<string, unknown>).processStream).toBeUndefined();

        expect((api.evals.classify as Record<string, unknown>).datasets).toBeUndefined();
        expect((api.evals.classify as Record<string, unknown>).templates).toBeUndefined();
        expect((api.evals.classify as Record<string, unknown>).create).toBeUndefined();
        expect(typeof api.evals.classify.process).toBe("function");
        expect((api.evals.classify as Record<string, unknown>).processStream).toBeUndefined();
    });

    test("builds process requests for extract, split, and classify", async () => {
        const mockClient = new MockClient();
        const api = new APIV1(mockClient);

        const extractResponse = await api.evals.extract.process({
            eval_id: "eval-extract",
            iteration_id: "iter-extract",
            document: payslipPath,
            metadata: { source: "unit-test" },
        });
        const splitResponse = await api.evals.split.process({
            eval_id: "eval-split",
            document: payslipPath,
            model: "retab-small",
        });
        const classifyResponse = await api.evals.classify.process({
            eval_id: "eval-classify",
            iteration_id: "iter-classify",
            document: payslipPath,
        });

        expect(extractResponse.choices[0].message.parsed).toEqual({ status: "extract" });
        expect(splitResponse.choices[0].message.parsed).toEqual({ status: "split" });
        expect(classifyResponse.classification.category).toBe("invoice");

        expect(mockClient.requests.map((request) => request.url)).toEqual([
            "/evals/extract/extract/eval-extract/iter-extract",
            "/evals/split/extract/eval-split",
            "/evals/classify/extract/eval-classify/iter-classify",
        ]);
        expect(mockClient.requests[0]?.bodyMime).toBe("multipart/form-data");
        expect(mockClient.requests[0]?.body?.metadata).toBe(JSON.stringify({ source: "unit-test" }));
    });

    test("streams extract eval responses", async () => {
        const mockClient = new MockClient();
        const api = new APIV1(mockClient);

        const stream = await api.evals.extract.processStream({
            eval_id: "eval-extract",
            document: payslipPath,
        });
        const chunks = [];
        for await (const chunk of stream) {
            chunks.push(chunk);
        }

        expect(chunks).toHaveLength(1);
        expect(chunks[0]?.choices[0]?.delta?.flat_parsed).toEqual({ status: "stream" });
        expect(mockClient.requests[0]?.url).toBe("/evals/extract/extract/eval-extract/stream");
    });

    test("rejects multi-document extract requests", async () => {
        const api = new APIV1(new MockClient());

        await expect(
            api.evals.extract.process({
                eval_id: "eval-extract",
                document: payslipPath,
                documents: [payslipPath],
            } as unknown as Parameters<typeof api.evals.extract.process>[0]),
        ).rejects.toThrow("accepts only 'document'");
    });
});
