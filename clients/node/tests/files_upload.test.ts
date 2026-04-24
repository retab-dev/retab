import { describe, expect, test } from "bun:test";

import APIV1 from "../src/api/client.js";
import { AbstractClient } from "../src/client.js";
import { ZMIMEData } from "../src/types.js";

type RecordedRequest = {
  url: string;
  method: string;
  params?: Record<string, unknown>;
  headers?: Record<string, unknown>;
  bodyMime?: "application/json" | "multipart/form-data";
  body?: Record<string, unknown>;
};

class RecordingClient extends AbstractClient {
  public request: RecordedRequest | null = null;

  protected async _fetch(params: RecordedRequest): Promise<Response> {
    this.request = params;
    return new Response(JSON.stringify({
      fileId: "file_123",
      filename: "invoice.pdf",
    }), {
      status: 200,
      headers: {
        "Content-Type": "application/json",
      },
    });
  }
}

describe("Node SDK files upload request", () => {
  test("accepts signed bucket URLs without base64 inlining", async () => {
    const signedUrl = "https://storage.googleapis.com/uiform-eu-multiregion/test/invoice.pdf?X-Goog-Signature=abc";
    const recorder = new RecordingClient();
    const api = new APIV1(recorder);

    const response = await api.files.upload(signedUrl);

    expect(response.file_id).toBe("file_123");
    expect(recorder.request?.url).toBe("/files/upload");
    expect(recorder.request?.method).toBe("POST");
    expect(recorder.request?.body).toEqual({
      mimeData: {
        filename: "invoice.pdf",
        url: signedUrl,
      },
    });
  });

  test("preserves signed bucket URLs for resource route MIME inputs", async () => {
    const signedUrl = "https://storage.googleapis.com/uiform-eu-multiregion/test/invoice.pdf?X-Goog-Signature=abc";

    const mimeData = await ZMIMEData.parseAsync(signedUrl);

    expect(mimeData).toEqual({
      filename: "invoice.pdf",
      url: signedUrl,
    });
  });
});
