import { describe, expect, test } from "bun:test";

import APIV1 from "../src/api/client.js";
import { AbstractClient } from "../src/client.js";
import { ZClassification, ZExtractionV2, ZPartition, ZSplit } from "../src/types.js";

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
        return new Response(JSON.stringify({
            data: [
                {
                    id: "retab-small",
                    object: "model",
                    created: 1710000000,
                    owned_by: "retab",
                },
            ],
        }), {
            status: 200,
            headers: {
                "Content-Type": "application/json",
            },
        });
    }
}

describe("models resource", () => {
    test("list() is exposed on the top-level client and mirrors the python endpoint", async () => {
        const mockClient = new MockClient();
        const api = new APIV1(mockClient);

        const models = await api.models.list({
            supports_finetuning: true,
            supports_image: true,
            include_finetuned_models: false,
        });

        expect(models).toHaveLength(1);
        expect(models[0]?.id).toBe("retab-small");
        expect(mockClient.requests[0]).toEqual({
            url: "/v1/models",
            method: "GET",
            params: {
                supports_finetuning: true,
                supports_image: true,
                include_finetuned_models: false,
            },
        });
    });

    test("processing resource schemas accept canonical normalized consensus payloads", () => {
        const classification = ZClassification.parse({
            id: "clss_123",
            file: {
                id: "file_123",
                filename: "doc.pdf",
                mime_type: "application/pdf",
            },
            model: "retab-small",
            categories: [{ name: "invoice", description: "Invoice" }],
            n_consensus: 1,
            output: { reasoning: "", category: "invoice" },
            consensus: { choices: [], likelihood: null },
        });

        const extraction = ZExtractionV2.parse({
            id: "extr_123",
            file: {
                id: "file_123",
                filename: "doc.pdf",
                mime_type: "application/pdf",
            },
            model: "retab-small",
            json_schema: { type: "object" },
            n_consensus: 1,
            image_resolution_dpi: 192,
            output: { invoice_number: "INV-001" },
            consensus: { choices: [], likelihoods: null },
            metadata: {},
        });

        const split = ZSplit.parse({
            id: "splt_123",
            file: {
                id: "file_123",
                filename: "doc.pdf",
                mime_type: "application/pdf",
            },
            model: "retab-small",
            subdocuments: [{ name: "invoice", description: "Invoice" }],
            n_consensus: 1,
            output: [{ name: "invoice", pages: [1] }],
            consensus: null,
        });

        const partition = ZPartition.parse({
            id: "prtn_123",
            file: {
                id: "file_123",
                filename: "doc.pdf",
                mime_type: "application/pdf",
            },
            model: "retab-small",
            key: "invoice_number",
            instructions: "Group by invoice number.",
            n_consensus: 1,
            output: [{ key: "INV-001", pages: [1] }],
            consensus: { choices: [], likelihoods: null },
        });

        expect(classification.consensus.choices).toEqual([]);
        expect(extraction.consensus.choices).toEqual([]);
        expect(split.consensus).toBeNull();
        expect(partition.consensus.choices).toEqual([]);
    });
});
