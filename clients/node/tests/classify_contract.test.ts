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
    constructor(private readonly responsePayload: unknown) {
        super();
    }

    protected async _fetch(_params: RecordedRequest): Promise<Response> {
        return new Response(JSON.stringify(this.responsePayload), {
            status: 200,
            headers: {
                "Content-Type": "application/json",
            },
        });
    }
}

const LEGACY_CLASSIFY_PAYLOAD = {
    result: {
        reasoning: "Detected invoice keywords",
        classification: "invoice",
    },
    likelihood: 0.67,
    votes: [
        { classification: "invoice", reasoning: "invoice vote one" },
        { classification: "receipt", reasoning: "receipt alternative" },
    ],
    usage: {
        credits: 0.12,
    },
};

describe("Node SDK classify response normalization", () => {
    test("documents.classify accepts legacy classify envelopes", async () => {
        const api = new APIV1(new MockClient(LEGACY_CLASSIFY_PAYLOAD));

        const response = await api.documents.classify({
            document: {
                filename: "invoice.txt",
                url: "data:text/plain;base64,SW52b2ljZQ==",
            },
            categories: [{ name: "invoice", description: "Invoice documents" }],
            model: "retab-small",
        });

        expect(response.classification).toEqual({
            reasoning: "Detected invoice keywords",
            category: "invoice",
        });
        expect(response.consensus.likelihood).toBe(0.67);
        expect(response.consensus.choices.map((choice) => choice.classification.category)).toEqual([
            "invoice",
            "receipt",
        ]);
    });

    test("evals.classify.process accepts legacy classify envelopes", async () => {
        const api = new APIV1(new MockClient(LEGACY_CLASSIFY_PAYLOAD));

        const response = await api.evals.classify.process({
            eval_id: "eval_123",
            document: {
                filename: "invoice.txt",
                url: "data:text/plain;base64,SW52b2ljZQ==",
            },
        });

        expect(response.classification.category).toBe("invoice");
        expect(response.consensus.choices).toHaveLength(2);
    });
});
