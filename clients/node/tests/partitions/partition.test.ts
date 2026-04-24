import { describe, expect, test } from "bun:test";

import APIV1 from "../../src/api/client.js";
import { AbstractClient } from "../../src/client.js";

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
        if (params.url === "/partitions" && params.method === "GET") {
            return new Response(
                JSON.stringify({
                    data: [],
                    list_metadata: {
                        before: null,
                        after: null,
                    },
                }),
                {
                    status: 200,
                    headers: { "Content-Type": "application/json" },
                },
            );
        }
        if (params.method === "DELETE") {
            return new Response("{}", {
                status: 200,
                headers: { "Content-Type": "application/json" },
            });
        }
        return new Response(
            JSON.stringify({
                id: "prtn_123",
                file: {
                    id: "file_123",
                    filename: "invoice.pdf",
                    mime_type: "application/pdf",
                },
                model: "retab-small",
                key: "invoice_number",
                instructions: "Group all pages belonging to the same invoice number.",
                n_consensus: 3,
                output: [
                    {
                        key: "INV-001",
                        pages: [1, 2],
                    },
                ],
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

describe("Node SDK partitions resource", () => {
    test("partitions.create posts the partition resource contract", async () => {
        const mockClient = new MockClient();
        const client = new APIV1(mockClient);

        const response = await client.partitions.create({
            document: SAMPLE_DOCUMENT,
            key: "invoice_number",
            instructions: "Group all pages belonging to the same invoice number.",
            model: "retab-small",
            n_consensus: 3,
            bust_cache: false,
        });

        expect(mockClient.requests).toHaveLength(1);
        expect(mockClient.requests[0]).toEqual({
            url: "/partitions",
            method: "POST",
            body: {
                document: {
                    filename: "invoice.pdf",
                    url: "data:application/pdf;base64,JVBERi0xLjQK",
                },
                key: "invoice_number",
                instructions: "Group all pages belonging to the same invoice number.",
                model: "retab-small",
                n_consensus: 3,
            },
            params: undefined,
            headers: undefined,
        });

        expect(response).toEqual({
            id: "prtn_123",
            file: {
                id: "file_123",
                filename: "invoice.pdf",
                mime_type: "application/pdf",
            },
            model: "retab-small",
            key: "invoice_number",
            instructions: "Group all pages belonging to the same invoice number.",
            n_consensus: 3,
            output: [
                {
                    key: "INV-001",
                    pages: [1, 2],
                },
            ],
            consensus: {
                choices: [],
                likelihoods: null,
            },
            usage: {
                credits: 1.0,
            },
        });
    });

    test("partitions.list and delete use resource routes", async () => {
        const mockClient = new MockClient();
        const client = new APIV1(mockClient);

        const listed = await client.partitions.list({ limit: 5, filename: "invoice.pdf" });
        await client.partitions.delete("prtn_123");

        expect(listed.data).toEqual([]);
        expect(mockClient.requests[0]).toEqual({
            url: "/partitions",
            method: "GET",
            params: {
                limit: 5,
                order: "desc",
                filename: "invoice.pdf",
            },
            headers: undefined,
        });
        expect(mockClient.requests[1]).toEqual({
            url: "/partitions/prtn_123",
            method: "DELETE",
            params: undefined,
            headers: undefined,
        });
    });
});
