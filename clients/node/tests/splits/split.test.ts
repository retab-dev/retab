import { describe, expect, test } from "bun:test";

import APIV1 from "../../src/api/client.js";
import { AbstractClient } from "../../src/client.js";

type RecordedRequest = {
    url: string;
    method: string;
    params?: Record<string, unknown>;
    headers?: Record<string, unknown>;
    body?: Record<string, unknown>;
};

class MockClient extends AbstractClient {
    public requests: RecordedRequest[] = [];

    protected async _fetch(params: RecordedRequest): Promise<Response> {
        this.requests.push(params);
        return new Response(
            JSON.stringify({
                id: "split_123",
                file: {
                    id: "file_123",
                    filename: "invoice.pdf",
                    mime_type: "application/pdf",
                },
                model: "retab-small",
                subdocuments: [
                    {
                        name: "invoice",
                        description: "invoice pages",
                        partition_key: "invoice_number",
                        allow_overlap: true,
                        allow_multiple_instances: false,
                    },
                ],
                n_consensus: 1,
                output: [{ name: "invoice", pages: [1, 2] }],
                consensus: {
                    choices: [],
                    likelihoods: null,
                },
                usage: {
                    credits: 1.0,
                },
            }),
            {
                status: 200,
                headers: { "Content-Type": "application/json" },
            },
        );
    }
}

const SAMPLE_DOCUMENT = {
    filename: "invoice.pdf",
    url: "data:application/pdf;base64,JVBERi0xLjQK",
    mime_type: "application/pdf",
};

describe("Node SDK splits resource", () => {
    test("splits.create forwards partition subdocument controls", async () => {
        const mockClient = new MockClient();
        const client = new APIV1(mockClient);

        await client.splits.create({
            document: SAMPLE_DOCUMENT,
            model: "retab-small",
            subdocuments: [
                {
                    name: "invoice",
                    description: "invoice pages",
                    partition_key: "invoice_number",
                    allow_overlap: true,
                },
            ],
        });

        expect(mockClient.requests).toHaveLength(1);
        expect(mockClient.requests[0]).toEqual({
            url: "/splits",
            method: "POST",
            body: {
                document: {
                    filename: "invoice.pdf",
                    url: "data:application/pdf;base64,JVBERi0xLjQK",
                },
                subdocuments: [
                    {
                        name: "invoice",
                        description: "invoice pages",
                        partition_key: "invoice_number",
                        allow_overlap: true,
                        allow_multiple_instances: false,
                    },
                ],
                model: "retab-small",
            },
            params: undefined,
            headers: undefined,
        });
    });
});
