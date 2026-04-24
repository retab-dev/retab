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

class RecordingClient extends AbstractClient {
  public requests: RecordedRequest[] = [];

  protected async _fetch(params: RecordedRequest): Promise<Response> {
    this.requests.push(params);
    return new Response(JSON.stringify({
      data: [],
      list_metadata: {
        before: null,
        after: null,
      },
    }), {
      status: 200,
      headers: {
        "Content-Type": "application/json",
      },
    });
  }
}

describe("Node SDK resource list filters", () => {
  test("includes filename filters for persisted processing resources", async () => {
    const recorder = new RecordingClient();
    const api = new APIV1(recorder);

    await api.classifications.list({ limit: 5, filename: "invoice.pdf" });
    await api.parses.list({ limit: 6, filename: "invoice.pdf" });
    await api.splits.list({ limit: 7, filename: "invoice.pdf" });

    expect(recorder.requests.map((request) => ({
      url: request.url,
      params: request.params,
    }))).toEqual([
      {
        url: "/classifications",
        params: {
          limit: 5,
          order: "desc",
          filename: "invoice.pdf",
        },
      },
      {
        url: "/parses",
        params: {
          limit: 6,
          order: "desc",
          filename: "invoice.pdf",
        },
      },
      {
        url: "/splits",
        params: {
          limit: 7,
          order: "desc",
          filename: "invoice.pdf",
        },
      },
    ]);
  });
});
