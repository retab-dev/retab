import { describe, expect, test } from "bun:test";
import fs from "fs";
import os from "os";
import path from "path";

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
  public requests: RecordedRequest[] = [];

  protected async _fetch(params: RecordedRequest): Promise<Response> {
    this.request = params;
    this.requests.push(params);
    const body = params.url === "/files/upload"
      ? {
        fileId: "file_123",
        filename: "invoice.pdf",
        uploadUrl: "https://storage.googleapis.com/signed-upload",
        uploadMethod: "PUT",
        uploadHeaders: { "Content-Type": "application/pdf" },
        storageUrl: "https://storage.retab.com/file_123",
        expiresAt: "2026-04-24T12:00:00Z",
      }
      : {
        fileId: "file_123",
        filename: "invoice.pdf",
        storageUrl: "https://storage.retab.com/file_123",
      };
    return new Response(JSON.stringify(body), {
      status: 200,
      headers: {
        "Content-Type": "application/json",
      },
    });
  }
}

describe("Node SDK files upload request", () => {
  test("rejects signed bucket URLs for files.upload", async () => {
    const signedUrl = "https://storage.googleapis.com/uiform-eu-multiregion/test/invoice.pdf?X-Goog-Signature=abc";
    const recorder = new RecordingClient();
    const api = new APIV1(recorder);

    await expect(api.files.upload(signedUrl)).rejects.toThrow("local file paths");
    expect(recorder.request).toBeNull();
  });

  test("rejects HTTP URLs for files.upload", async () => {
    const recorder = new RecordingClient();
    const api = new APIV1(recorder);

    await expect(api.files.upload("http://example.com/invoice.pdf")).rejects.toThrow("local file paths");
    expect(recorder.request).toBeNull();
  });

  test("preserves signed bucket URLs for resource route MIME inputs", async () => {
    const signedUrl = "https://storage.googleapis.com/uiform-eu-multiregion/test/invoice.pdf?X-Goog-Signature=abc";

    const mimeData = await ZMIMEData.parseAsync(signedUrl);

    expect(mimeData).toEqual({
      filename: "invoice.pdf",
      url: signedUrl,
    });
  });

  test("prepare helpers expose the upload session and completion contracts", async () => {
    const api = new APIV1(new RecordingClient());

    expect(api.files.prepare_upload("invoice.pdf", "application/pdf", 8, "abc123")).toEqual({
      url: "/files/upload",
      method: "POST",
      body: {
        filename: "invoice.pdf",
        content_type: "application/pdf",
        size_bytes: 8,
        sha256: "abc123",
      },
    });
    expect(api.files.prepare_complete_upload("file_123", "abc123")).toEqual({
      url: "/files/upload/file_123/complete",
      method: "POST",
      body: {
        sha256: "abc123",
      },
    });
    expect(api.files.prepare_complete_upload("file_123")).toEqual({
      url: "/files/upload/file_123/complete",
      method: "POST",
      body: {},
    });
  });

  test("uses direct storage upload for local paths", async () => {
    const tempDir = await fs.promises.mkdtemp(path.join(os.tmpdir(), "retab-upload-"));
    const filePath = path.join(tempDir, "invoice.pdf");
    await fs.promises.writeFile(filePath, "%PDF-1.4");
    const recorder = new RecordingClient();
    const api = new APIV1(recorder);

    const originalFetch = globalThis.fetch;
    const directUploads: Array<{ url: string; init?: RequestInit }> = [];
    globalThis.fetch = (async (url: string | URL | Request, init?: RequestInit) => {
      directUploads.push({ url: String(url), init });
      return new Response("", { status: 200 });
    }) as typeof fetch;

    try {
      const response = await api.files.upload(filePath);

      expect(response.file_id).toBe("file_123");
    } finally {
      globalThis.fetch = originalFetch;
    }

    expect(recorder.requests.map((request) => request.url)).toEqual([
      "/files/upload",
      "/files/upload/file_123/complete",
    ]);
    expect(recorder.requests[0].body).toEqual({
      filename: "invoice.pdf",
      content_type: "application/pdf",
      size_bytes: 8,
      sha256: "e16fa5d9b51928755db85b917f0297babaf22c7a47e97d9212adab56e61ba04e",
    });
    expect(recorder.requests[1].body).toEqual({
      sha256: "e16fa5d9b51928755db85b917f0297babaf22c7a47e97d9212adab56e61ba04e",
    });
    expect(directUploads).toHaveLength(1);
    expect(directUploads[0].url).toBe("https://storage.googleapis.com/signed-upload");
    expect(directUploads[0].init?.method).toBe("PUT");
    expect(directUploads[0].init?.headers).toEqual({ "Content-Type": "application/pdf" });
  });

  test("falls back to octet-stream for unknown file extensions", async () => {
    const tempDir = await fs.promises.mkdtemp(path.join(os.tmpdir(), "retab-upload-"));
    const filePath = path.join(tempDir, "payload.unknownext");
    await fs.promises.writeFile(filePath, "payload");
    const recorder = new RecordingClient();
    const api = new APIV1(recorder);

    const originalFetch = globalThis.fetch;
    globalThis.fetch = (async () => new Response("", { status: 200 })) as typeof fetch;

    try {
      await api.files.upload(filePath);
    } finally {
      globalThis.fetch = originalFetch;
    }

    expect(recorder.requests[0].body).toEqual({
      filename: "payload.unknownext",
      content_type: "application/octet-stream",
      size_bytes: 7,
      sha256: "239f59ed55e737c77147cf55ad0c1b030b6d7ee748a7426952f9b852d5a935e5",
    });
  });

  test("does not complete when direct storage upload fails", async () => {
    const tempDir = await fs.promises.mkdtemp(path.join(os.tmpdir(), "retab-upload-"));
    const filePath = path.join(tempDir, "invoice.pdf");
    await fs.promises.writeFile(filePath, "%PDF-1.4");
    const recorder = new RecordingClient();
    const api = new APIV1(recorder);

    const originalFetch = globalThis.fetch;
    globalThis.fetch = (async () => new Response("signature mismatch", { status: 403 })) as typeof fetch;

    try {
      await expect(api.files.upload(filePath)).rejects.toThrow("Direct file upload failed: 403 signature mismatch");
    } finally {
      globalThis.fetch = originalFetch;
    }

    expect(recorder.requests.map((request) => request.url)).toEqual(["/files/upload"]);
  });
});
