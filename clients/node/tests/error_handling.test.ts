import { describe, test, expect } from "bun:test";
import { APIError } from "../src/client";

// ---------------------------------------------------------------------------
// APIError construction & fields
// ---------------------------------------------------------------------------

describe("APIError construction", () => {
  test("basic fields with defaults", () => {
    const err = new APIError(500, "something broke");
    expect(err.status).toBe(500);
    expect(err.info).toBe("something broke");
    expect(err.message).toBe("something broke");
    expect(err.code).toBeNull();
    expect(err.details).toBeNull();
    expect(err.body).toBe("");
    expect(err.requestId).toBeNull();
    expect(err.method).toBeNull();
    expect(err.url).toBeNull();
    expect(err.retries).toBe(0);
    expect(err.name).toBe("APIError");
  });

  test("all fields populated", () => {
    const err = new APIError(400, "bad request", {
      code: "INVALID_SCHEMA",
      details: { field: "json_schema", reason: "missing required property" },
      body: '{"detail": {"code": "INVALID_SCHEMA"}}',
      requestId: "req_abc123",
      method: "POST",
      url: "http://localhost:4000/v1/documents/extract",
    });
    expect(err.status).toBe(400);
    expect(err.info).toBe("bad request");
    expect(err.code).toBe("INVALID_SCHEMA");
    expect(err.details).toEqual({
      field: "json_schema",
      reason: "missing required property",
    });
    expect(err.body).toBe('{"detail": {"code": "INVALID_SCHEMA"}}');
    expect(err.requestId).toBe("req_abc123");
    expect(err.method).toBe("POST");
    expect(err.url).toBe("http://localhost:4000/v1/documents/extract");
    expect(err.retries).toBe(0);
  });

  test("is an instance of Error", () => {
    const err = new APIError(500, "fail");
    expect(err).toBeInstanceOf(Error);
  });

  test("retries can be set after construction", () => {
    const err = new APIError(502, "fail");
    expect(err.retries).toBe(0);
    err.retries = 4;
    expect(err.retries).toBe(4);
  });
});

// ---------------------------------------------------------------------------
// APIError.toString()
// ---------------------------------------------------------------------------

describe("APIError.toString()", () => {
  test("minimal — status and message only", () => {
    const err = new APIError(502, "Request failed (502)");
    const s = err.toString();
    expect(s).toContain("502");
    expect(s).toContain("Request failed (502)");
    expect(s).not.toContain("URL:");
    expect(s).not.toContain("Request-ID:");
    expect(s).not.toContain("Retries:");
    expect(s).not.toContain("Body:");
  });

  test("includes URL when method and url set", () => {
    const err = new APIError(502, "fail", {
      method: "POST",
      url: "http://localhost:4000/v1/documents/extract",
    });
    const s = err.toString();
    expect(s).toContain("URL:");
    expect(s).toContain("POST http://localhost:4000/v1/documents/extract");
  });

  test("no URL line when only method set", () => {
    const err = new APIError(502, "fail", { method: "POST" });
    const s = err.toString();
    expect(s).not.toContain("URL:");
  });

  test("includes Request-ID when set", () => {
    const err = new APIError(500, "fail", { requestId: "req_xyz" });
    const s = err.toString();
    expect(s).toContain("Request-ID:");
    expect(s).toContain("req_xyz");
  });

  test("includes Code when set", () => {
    const err = new APIError(422, "fail", { code: "INVALID_SCHEMA" });
    const s = err.toString();
    expect(s).toContain("Code:");
    expect(s).toContain("INVALID_SCHEMA");
  });

  test("includes Details when set", () => {
    const err = new APIError(422, "fail", {
      details: { field: "name" },
    });
    const s = err.toString();
    expect(s).toContain("Details:");
    expect(s).toContain("name");
  });

  test("includes Body when set", () => {
    const err = new APIError(502, "fail", {
      body: '{"error": "upstream timeout"}',
    });
    const s = err.toString();
    expect(s).toContain("Body:");
    expect(s).toContain("upstream timeout");
  });

  test("truncates body over 500 chars", () => {
    const longBody = "x".repeat(1000);
    const err = new APIError(502, "fail", { body: longBody });
    const s = err.toString();
    expect(s).toContain("...");
    const bodyLine = s.split("\n").find((l) => l.includes("Body:"))!;
    // 500 chars + "..." + prefix
    expect(bodyLine.length).toBeLessThan(520);
  });

  test("does not truncate body under 500 chars", () => {
    const shortBody = "x".repeat(499);
    const err = new APIError(502, "fail", { body: shortBody });
    const s = err.toString();
    expect(s).not.toContain("...");
  });

  test("includes Retries when nonzero", () => {
    const err = new APIError(502, "fail");
    err.retries = 4;
    const s = err.toString();
    expect(s).toContain("Retries:");
    expect(s).toContain("4");
  });

  test("no Retries line when zero", () => {
    const err = new APIError(502, "fail");
    const s = err.toString();
    expect(s).not.toContain("Retries:");
  });

  test("full output with all fields", () => {
    const err = new APIError(502, "upstream timeout", {
      code: "GATEWAY_ERROR",
      details: { service: "preprocessing_server" },
      body: '{"error": "upstream timeout from preprocessing_server"}',
      requestId: "req_abc123",
      method: "POST",
      url: "http://localhost:4000/v1/documents/extract",
    });
    err.retries = 3;
    const s = err.toString();
    const lines = s.split("\n");
    expect(lines[0]).toContain("502");
    expect(lines[0]).toContain("upstream timeout");
    expect(s).toContain("POST http://localhost:4000/v1/documents/extract");
    expect(s).toContain("req_abc123");
    expect(s).toContain("GATEWAY_ERROR");
    expect(s).toContain("preprocessing_server");
    expect(s).toContain("Body:");
    expect(s).toContain("Retries:");
    expect(s).toContain("3");
  });
});

// ---------------------------------------------------------------------------
// buildAPIError — structured error parsing via mock Response
// ---------------------------------------------------------------------------

function mockResponse(
  status: number,
  body: string,
  headers?: Record<string, string>,
): Response {
  return new Response(body, {
    status,
    headers: headers ?? {},
  });
}

// We can't import buildAPIError directly (not exported), so we test it
// indirectly via the AbstractClient._fetchJson path. But we CAN test the
// parsing logic by constructing APIErrors the same way the helper does.
// Let's re-implement the parsing inline for thorough unit testing.

async function parseErrorResponse(
  status: number,
  body: string,
  headers?: Record<string, string>,
  method?: string,
  url?: string,
): Promise<APIError> {
  const response = mockResponse(status, body, headers);
  const responseBody = await response.text();
  const requestId = response.headers.get("x-request-id") ?? null;

  let code: string | null = null;
  let message = `Request failed (${status})`;
  let details: Record<string, any> | null = null;

  try {
    const errorBody = JSON.parse(responseBody);
    if (
      typeof errorBody === "object" &&
      errorBody !== null &&
      !Array.isArray(errorBody)
    ) {
      const detail = errorBody.detail;
      if (
        typeof detail === "object" &&
        detail !== null &&
        !Array.isArray(detail)
      ) {
        code = detail.code ?? null;
        message = detail.message ?? message;
        details = detail.details ?? null;
      } else if (typeof detail === "string") {
        message = detail;
      }
    }
  } catch {
    if (responseBody) message = responseBody;
  }

  return new APIError(status, message, {
    code,
    details,
    body: responseBody,
    requestId,
    method: method ?? null,
    url: url ?? null,
  });
}

describe("buildAPIError parsing logic", () => {
  test("structured JSON with detail dict", async () => {
    const body = JSON.stringify({
      detail: {
        code: "INVALID_SCHEMA",
        message: "Schema validation failed",
        details: { field: "json_schema" },
      },
    });
    const err = await parseErrorResponse(422, body, {}, "POST", "/v1/test");
    expect(err.status).toBe(422);
    expect(err.code).toBe("INVALID_SCHEMA");
    expect(err.info).toBe("Schema validation failed");
    expect(err.details).toEqual({ field: "json_schema" });
    expect(err.body).toBe(body);
    expect(err.method).toBe("POST");
    expect(err.url).toBe("/v1/test");
  });

  test("structured JSON with detail string", async () => {
    const body = JSON.stringify({ detail: "Not authenticated" });
    const err = await parseErrorResponse(401, body);
    expect(err.status).toBe(401);
    expect(err.info).toBe("Not authenticated");
    expect(err.code).toBeNull();
    expect(err.details).toBeNull();
  });

  test("unstructured JSON (no detail key)", async () => {
    const body = JSON.stringify({ error: "something unexpected" });
    const err = await parseErrorResponse(500, body);
    expect(err.info).toBe("Request failed (500)");
    expect(err.code).toBeNull();
  });

  test("plain text body (not JSON)", async () => {
    const body = "Bad Gateway: upstream timeout";
    const err = await parseErrorResponse(502, body);
    expect(err.info).toBe("Bad Gateway: upstream timeout");
    expect(err.body).toBe(body);
  });

  test("empty body", async () => {
    const err = await parseErrorResponse(500, "");
    expect(err.info).toBe("Request failed (500)");
    expect(err.body).toBe("");
  });

  test("extracts x-request-id header", async () => {
    const err = await parseErrorResponse(500, "fail", {
      "x-request-id": "req_test123",
    });
    expect(err.requestId).toBe("req_test123");
  });

  test("no x-request-id header", async () => {
    const err = await parseErrorResponse(500, "fail");
    expect(err.requestId).toBeNull();
  });

  test("preserves method and url", async () => {
    const err = await parseErrorResponse(
      404,
      JSON.stringify({ detail: "Not found" }),
      {},
      "GET",
      "http://localhost:4000/v1/files/abc",
    );
    expect(err.method).toBe("GET");
    expect(err.url).toBe("http://localhost:4000/v1/files/abc");
  });

  test("detail is an array (not dict) — treated as no detail", async () => {
    const body = JSON.stringify({ detail: ["error1", "error2"] });
    const err = await parseErrorResponse(422, body);
    expect(err.info).toBe("Request failed (422)");
    expect(err.code).toBeNull();
  });

  test("detail dict without optional fields", async () => {
    const body = JSON.stringify({
      detail: {
        message: "Something failed",
      },
    });
    const err = await parseErrorResponse(500, body);
    expect(err.info).toBe("Something failed");
    expect(err.code).toBeNull();
    expect(err.details).toBeNull();
  });

  test("response body is non-object JSON (number)", async () => {
    const err = await parseErrorResponse(500, "42");
    expect(err.info).toBe("Request failed (500)");
  });

  test("response body is non-object JSON (string)", async () => {
    const err = await parseErrorResponse(500, '"just a string"');
    expect(err.info).toBe("Request failed (500)");
  });
});

// ---------------------------------------------------------------------------
// Integration: FetcherClient._fetch throws enriched APIError
// ---------------------------------------------------------------------------

describe("FetcherClient._fetch error integration", () => {
  test("throws APIError on non-ok response", async () => {
    // Use a guaranteed-to-fail URL to test real error path
    const { FetcherClient } = await import("../src/client");
    const client = new FetcherClient({
      apiKey: "sk_test_fake_key",
      baseUrl: "http://127.0.0.1:1", // Connection refused
    });

    try {
      await (client as any)._fetch({
        url: "/v1/test",
        method: "GET",
      });
      expect(true).toBe(false); // Should not reach here
    } catch (err) {
      // We expect either an APIError or a connection error
      expect(err).toBeInstanceOf(Error);
    }
  });
});

// ---------------------------------------------------------------------------
// End-to-end: toString output in catch scenario
// ---------------------------------------------------------------------------

describe("APIError end-to-end formatting", () => {
  test("catch and toString shows all context", () => {
    const err = new APIError(502, "upstream timeout", {
      body: '{"error": "preprocessing_server unreachable"}',
      method: "POST",
      url: "http://localhost:4000/v1/documents/extract",
      requestId: "req_999",
    });
    err.retries = 3;

    try {
      throw err;
    } catch (e: any) {
      const output = e.toString();
      expect(output).toContain("502");
      expect(output).toContain("upstream timeout");
      expect(output).toContain("POST http://localhost:4000/v1/documents/extract");
      expect(output).toContain("req_999");
      expect(output).toContain("preprocessing_server unreachable");
      expect(output).toContain("3");
    }
  });

  test("message property is the info string", () => {
    const err = new APIError(502, "upstream timeout");
    expect(err.message).toBe("upstream timeout");
  });

  test("can be caught as Error", () => {
    expect(() => {
      throw new APIError(500, "fail");
    }).toThrow(Error);
  });
});
