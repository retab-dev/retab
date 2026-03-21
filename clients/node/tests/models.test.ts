import { describe, expect, test } from "bun:test";

import APIV1 from "../src/api/client.js";
import { AbstractClient } from "../src/client.js";

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
});
