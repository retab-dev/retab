import * as z from "zod";
import * as crypto from "crypto";

type FetchParams = {
  url: string,
  method: string,
  params?: Record<string, any>,
  headers?: Record<string, any>,
  bodyMime?: "application/json" | "multipart/form-data",
  body?: Record<string, any> | unknown[],
  auth?: string[],
};

const PYTHON_PUBLIC_PREPARE_METHODS: Record<string, string[]> = {
  APIModels: ["list"],
  APISchemas: ["generate"],
  APIExtractions: ["create", "createStream", "list", "get", "sources", "delete"],
  APIFiles: ["upload", "list", "get", "get_download_link"],
  APIProjects: ["create", "get", "list", "delete", "publish", "extract", "split"],
  ProjectDatasets: ["create", "get", "list", "update", "delete", "duplicate", "add_document", "get_document", "list_documents", "update_document", "delete_document"],
  ProjectIterations: ["create", "get", "list", "update_draft", "delete", "finalize", "get_schema", "process_documents", "get_document", "list_documents", "update_document", "delete_document", "get_metrics"],
  APIWorkflows: ["get", "list", "create", "update", "delete", "publish", "duplicate", "get_entities"],
  APIWorkflowBlocks: ["list", "get", "create", "create_batch", "update", "delete"],
  APIWorkflowEdges: ["list", "get", "create", "create_batch", "delete", "delete_all"],
  APIWorkflowRuns: ["create", "get", "list", "delete", "cancel", "restart", "submit_hil_decision", "get_hil_decision"],
  APIWorkflowRunSteps: ["get", "list", "get_many"],
  APIEvalsExtract: ["process", "process_stream"],
  APIEvalsSplit: ["process"],
  APIEvalsClassify: ["process"],
};

async function* streamResponse<ZodSchema extends z.ZodType<any, any, any>>(schema: ZodSchema, response: Response): AsyncGenerator<z.output<ZodSchema>> {
  let body = "";
  let depth = 0;
  let inString = false;
  if (!response.body) {
    throw new APIError(response.status, "Response body is empty");
  }
  let reader = response.body.getReader();
  while (true) {
    let chunk = await reader.read();
    if (!chunk.value) break;
    let string = new TextDecoder().decode(chunk.value, { stream: true });
    let prevBodyLength = body.length;
    body += string;
    for (let i = 0; i < string.length; i++) {
      let char = string[i];
      if (char === '"') {
        inString = !inString;
      }
      if (inString) {
        if (char === "\\") {
          i++;
        }
      } else {
        if (char === "{") {
          depth++;
        } else if (char === "}") {
          depth--;
          if (depth === 0) {
            yield schema.parseAsync(JSON.parse(body.slice(0, prevBodyLength + i + 1)));
            body = body.slice(prevBodyLength + i + 1);
            prevBodyLength = -i - 1;
          }
        }
      }
    }
    if (chunk.done) break;
  }
}
export class AbstractClient {
  protected _fetch(_: FetchParams): Promise<Response> {
    throw new Error("Method not implemented");
  }
  protected async _fetchJson(params: FetchParams): Promise<void>;
  protected async _fetchJson<ZodSchema extends z.ZodType<any, any, any>>(bodyType: ZodSchema, params: FetchParams): Promise<z.output<ZodSchema>>;
  protected async _fetchJson<ZodSchema extends z.ZodType<any, any, any>>(_bodyType: ZodSchema | FetchParams, _params?: FetchParams): Promise<z.output<ZodSchema> | void> {
    let params = _params || (_bodyType as FetchParams);
    let bodyType = _params ? _bodyType as ZodSchema : undefined;
    let response = await this._fetch(params);
    if (!response.ok) {
      throw await buildAPIError(response, params.method, params.url);
    }
    if (!bodyType) {
      return;
    }
    if (!response.headers.get("Content-Type")?.startsWith("application/json")) throw new APIError(response.status, "Response is not JSON");
    return bodyType.parseAsync(await response.json());
  }
  protected async _fetchStream<ZodSchema extends z.ZodType<any, any, any>>(schema: ZodSchema, params: FetchParams): Promise<AsyncGenerator<z.output<ZodSchema>>> {
    let response = await this._fetch(params);
    if (!response.ok) {
      throw await buildAPIError(response, params.method, params.url);
    }
    if (!response.headers.get("Content-Type")?.startsWith("application/stream+json")) throw new APIError(response.status, "Response is not stream JSON");
    return streamResponse(schema, response);
  }
}

export class CompositionClient extends AbstractClient {
  protected _client: AbstractClient;
  constructor(client: AbstractClient) {
    super();
    this._client = client;
    this._installPrepareMethods();
  }
  protected _fetch(params: FetchParams): Promise<Response> {
    return this._client["_fetch"](params);
  }

  private _installPrepareMethods(): void {
    const allowedMethods = PYTHON_PUBLIC_PREPARE_METHODS[this.constructor.name] ?? [];
    if (allowedMethods.length === 0) {
      return;
    }

    for (const methodName of allowedMethods) {
      const method = (this as Record<string, unknown>)[methodName];
      const prepareMethodName = `prepare_${methodName}`;
      if (typeof method !== "function" || prepareMethodName in this) {
        continue;
      }

      Object.defineProperty(this, prepareMethodName, {
        configurable: true,
        enumerable: false,
        writable: false,
        value: (...args: unknown[]) => this._capturePreparedRequest(methodName, args),
      });
    }
  }

  protected async _capturePreparedRequest(methodName: string, args: unknown[]): Promise<FetchParams> {
    let capturedRequest: FetchParams | undefined;
    const self = this as Record<string, any>;
    const originalFetch = self._fetch;
    const originalFetchJson = self._fetchJson;
    const originalFetchStream = self._fetchStream;

    self._fetch = async (params: FetchParams) => {
      capturedRequest = params;
      return new Response("{}", {
        status: 200,
        headers: {
          "Content-Type": "application/json",
        },
      });
    };

    self._fetchJson = async (...fetchArgs: unknown[]) => {
      capturedRequest = (fetchArgs.length === 2 ? fetchArgs[1] : fetchArgs[0]) as FetchParams;
      return {};
    };

    self._fetchStream = async (...fetchArgs: unknown[]) => {
      capturedRequest = fetchArgs[1] as FetchParams;
      return (async function* emptyStream() {})();
    };

    try {
      await self[methodName](...args);
    } finally {
      self._fetch = originalFetch;
      self._fetchJson = originalFetchJson;
      self._fetchStream = originalFetchStream;
    }

    if (!capturedRequest) {
      throw new Error(`Unable to capture prepared request for ${methodName}`);
    }

    return capturedRequest;
  }
}

export class APIError extends Error {
  status: number;
  info: string;
  code: string | null;
  details: Record<string, any> | null;
  body: string;
  requestId: string | null;
  method: string | null;
  url: string | null;
  retries: number;

  constructor(
    status: number,
    info: string,
    options?: {
      code?: string | null;
      details?: Record<string, any> | null;
      body?: string;
      requestId?: string | null;
      method?: string | null;
      url?: string | null;
    },
  ) {
    super(info);
    this.name = "APIError";
    this.status = status;
    this.info = info;
    this.code = options?.code ?? null;
    this.details = options?.details ?? null;
    this.body = options?.body ?? "";
    this.requestId = options?.requestId ?? null;
    this.method = options?.method ?? null;
    this.url = options?.url ?? null;
    this.retries = 0;
  }

  override toString(): string {
    const lines = [`${this.status} — ${this.info}`];
    if (this.method && this.url) {
      lines.push(`  URL:        ${this.method} ${this.url}`);
    }
    if (this.requestId) {
      lines.push(`  Request-ID: ${this.requestId}`);
    }
    if (this.code) {
      lines.push(`  Code:       ${this.code}`);
    }
    if (this.details) {
      lines.push(`  Details:    ${JSON.stringify(this.details)}`);
    }
    if (this.body) {
      const truncated = this.body.length > 500 ? this.body.slice(0, 500) + "..." : this.body;
      lines.push(`  Body:       ${truncated}`);
    }
    if (this.retries > 0) {
      lines.push(`  Retries:    ${this.retries}`);
    }
    return lines.join("\n");
  }
}

async function buildAPIError(response: Response, method?: string, url?: string): Promise<APIError> {
  const body = await response.text();
  const requestId = response.headers.get("x-request-id") ?? null;

  let code: string | null = null;
  let message = `Request failed (${response.status})`;
  let details: Record<string, any> | null = null;

  try {
    const errorBody = JSON.parse(body);
    if (typeof errorBody === "object" && errorBody !== null && !Array.isArray(errorBody)) {
      const detail = errorBody.detail;
      if (typeof detail === "object" && detail !== null && !Array.isArray(detail)) {
        code = detail.code ?? null;
        message = detail.message ?? message;
        details = detail.details ?? null;
      } else if (typeof detail === "string") {
        message = detail;
      }
    }
  } catch {
    if (body) message = body;
  }

  return new APIError(response.status, message, {
    code,
    details,
    body,
    requestId,
    method: method ?? null,
    url: url ?? null,
  });
}

export class SignatureVerificationError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "SignatureVerificationError";
  }
}

export const DateOrISO = z.union([
  z.date(),
  z.string().refine(val => !isNaN(Date.parse(val)), {
    message: "Invalid date string",
  }).transform(val => new Date(val)),
]);

type AuthTypes = { apiKey: string } | {};
export type ClientOptions = {
  baseUrl?: string,
  timeout?: number,
} & AuthTypes;

export type RequestOptions = {
  params?: Record<string, any>,
  headers?: Record<string, any>,
  body?: Record<string, any>,
};

export class FetcherClient extends AbstractClient {
  options: ClientOptions;
  timeout: number;
  constructor(options?: ClientOptions) {
    super();
    this.options = options || {};
    this.timeout = this.options.timeout ?? 1800000; // Default 1800 seconds (30 minutes) (in milliseconds)

    // Validate that at least one authentication method is provided
    const apiKey = "apiKey" in this.options ? this.options.apiKey : process.env["RETAB_API_KEY"];

    if (!apiKey) {
      throw new Error(
        "Authentication required: Please provide an API key. " +
        "You can pass it in the constructor options or set the RETAB_API_KEY environment variable."
      );
    }
  }

  async _fetch(params: {
    url: string;
    method: string;
    params?: Record<string, any>;
    headers?: Record<string, any>;
    bodyMime?: "application/json" | "multipart/form-data";
    body?: Record<string, any> | unknown[];
  }): Promise<Response> {
    let query = "";
    if (params.params) {
      query = "?" + new URLSearchParams(
        Object.fromEntries(
          Object.entries(params.params).filter(([_, v]) => v !== undefined)
        )
      ).toString();
    }
    let url = (this.options.baseUrl || "https://api.retab.com/v1") + params.url + query;
    let headers = params.headers || {};
    let init: RequestInit = {
      method: params.method,
    };
    if (params.method !== "GET") {
      if (params.bodyMime === "multipart/form-data") {
        let formData: FormData = new FormData();
        const multipartBody = (params.body || {}) as Record<string, any>;
        for (const key of Object.keys(multipartBody)) {
          let value = multipartBody[key];
          if (Array.isArray(value)) {
            for (const item of value) {
              formData.append(key, item);
            }
            continue;
          }
          formData.append(key, value);
        }
        init.body = formData;
        // Don't set Content-Type for multipart/form-data - let FormData set it with boundary
      } else {
        headers["Content-Type"] = params.bodyMime || "application/json";
        // Default to empty object if body is undefined to ensure valid JSON is sent
        init.body = JSON.stringify(params.body ?? {});
      }
    }
    const apiKey = "apiKey" in this.options ? this.options.apiKey : process.env["RETAB_API_KEY"];
    headers["Api-Key"] = apiKey;
    init.headers = headers;
    init.signal = AbortSignal.timeout(this.timeout);
    let res = await fetch(url, init);
    if (!res.ok) {
      throw await buildAPIError(res, params.method, url);
    }
    return res;
  }

  /**
   * Verify the signature of a webhook event.
   * 
   * @param eventBody - The raw request body as a string or Buffer
   * @param eventSignature - The signature from the request header (x-retab-signature)
   * @param secret - The webhook secret key used for signing
   * @returns The parsed event payload (JSON)
   * @throws {SignatureVerificationError} If the signature verification fails
   * 
   * @example
   * ```typescript
   * import { FetcherClient } from './client';
   * 
   * // In your webhook handler
   * const secret = "your_webhook_secret";
   * const body = req.body; // Raw string or Buffer
   * const signature = req.headers['x-retab-signature'];
   * 
   * try {
   *   const event = FetcherClient.verifyEvent(body, signature, secret);
   *   console.log(`Verified event: ${event}`);
   * } catch (error) {
   *   console.log("Invalid signature!");
   * }
   * ```
   */
  static verifyEvent(eventBody: string | Buffer, eventSignature: string, secret: string): any {
    const bodyBuffer = typeof eventBody === 'string' ? Buffer.from(eventBody, 'utf-8') : eventBody;
    const expectedSignature = crypto
      .createHmac('sha256', secret)
      .update(bodyBuffer)
      .digest('hex');

    // Use constant-time comparison to prevent timing attacks
    if (!crypto.timingSafeEqual(Buffer.from(eventSignature), Buffer.from(expectedSignature))) {
      throw new SignatureVerificationError("Invalid signature");
    }

    return JSON.parse(bodyBuffer.toString('utf-8'));
  }
}
