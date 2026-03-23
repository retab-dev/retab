import { describe, expect, test } from "bun:test";
import path from "path";

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

        if (params.url === "/documents/extractions") {
            return new Response(
                JSON.stringify({
                    id: "chunk-123",
                    object: "chat.completion.chunk",
                    created: 1710000000,
                    model: "retab-micro",
                    choices: [
                        {
                            index: 0,
                            delta: {
                                content: "{}",
                                is_valid_json: true,
                                flat_parsed: {},
                                flat_likelihoods: {},
                                flat_deleted_keys: [],
                            },
                        },
                    ],
                }),
                {
                    status: 200,
                    headers: {
                        "Content-Type": "application/stream+json",
                    },
                },
            );
        }

        return new Response(
            JSON.stringify({
                id: "chatcmpl-123",
                object: "chat.completion",
                created: 1710000000,
                model: "retab-micro",
                choices: [
                    {
                        index: 0,
                        finish_reason: "stop",
                        message: {
                            role: "assistant",
                            content: "{}",
                            parsed: {},
                        },
                    },
                ],
            }),
            {
                status: 200,
                headers: {
                    "Content-Type": "application/json",
                },
            },
        );
    }
}

const bookingConfirmationPath = path.resolve(
    import.meta.dirname,
    "../data/freight/booking_confirmation_1.jpg",
);

describe("Retab SDK Extract Request Tests", () => {
    test("extract preserves the original filename for file path documents", async () => {
        const mockClient = new MockClient();
        const client = new APIV1(mockClient);

        const additionalMessages = [
            { role: "developer" as const, content: "Use the document filename when building context." },
            { role: "user" as const, content: "Extract the booking confirmation data." },
        ];

        await client.documents.extract({
            json_schema: {
                type: "object",
                properties: {
                    booking_reference: { type: "string" },
                },
            },
            document: bookingConfirmationPath,
            model: "retab-micro",
            additional_messages: additionalMessages,
        });

        const request = mockClient.requests[0];
        expect(request?.url).toBe("/documents/extract");
        expect((request?.body?.document as { filename?: string }).filename).toBe("booking_confirmation_1.jpg");
        expect(request?.body?.additional_messages).toEqual(additionalMessages);
    });

    test("extract_stream uses the streaming extractions endpoint", async () => {
        const mockClient = new MockClient();
        const client = new APIV1(mockClient);

        await client.documents.extract_stream({
            json_schema: {
                type: "object",
                properties: {
                    booking_reference: { type: "string" },
                },
            },
            document: bookingConfirmationPath,
            model: "retab-micro",
        });

        const request = mockClient.requests[0];
        expect(request?.url).toBe("/documents/extractions");
        expect(request?.body?.stream).toBe(true);
        expect((request?.body?.document as { filename?: string }).filename).toBe("booking_confirmation_1.jpg");
    });
});
