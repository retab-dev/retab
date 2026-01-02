import * as z from "zod";
import * as crypto from "crypto";

type FetchParams = {
  url: string,
  method: string,
  params?: Record<string, any>,
  headers?: Record<string, any>,
  bodyMime?: "application/json" | "multipart/form-data",
  body?: Record<string, any>,
  auth?: string[],
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
      throw new APIError(response.status, await response.text());
    }
    if (!bodyType) {
      return;
    }
    if (response.headers.get("Content-Type") !== "application/json") throw new APIError(response.status, "Response is not JSON");
    return bodyType.parseAsync(await response.json());
  }
  protected async _fetchStream<ZodSchema extends z.ZodType<any, any, any>>(schema: ZodSchema, params: FetchParams): Promise<AsyncGenerator<z.output<ZodSchema>>> {
    let response = await this._fetch(params);
    if (!response.ok) {
      throw new APIError(response.status, await response.text());
    }
    if (response.headers.get("Content-Type") !== "application/stream+json") throw new APIError(response.status, "Response is not stream JSON");
    return streamResponse(schema, response);
  }
}

export class CompositionClient extends AbstractClient {
  protected _client: AbstractClient;
  constructor(client: AbstractClient) {
    super();
    this._client = client;
  }
  protected _fetch(params: FetchParams): Promise<Response> {
    return this._client["_fetch"](params);
  }
}

export class APIError extends Error {
  status: number;
  info: string;
  constructor(status: number, info: string) {
    super(`API Error ${status}: ${info}`);
    this.status = status;
    this.info = info;
  }
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
    body?: Record<string, any>,
  }): Promise<Response> {
    let query = "";
    if (params.params) {
      query = "?" + new URLSearchParams(
        Object.fromEntries(
          Object.entries(params.params).filter(([_, v]) => v !== undefined)
        )
      ).toString();
    }
    let url = (this.options.baseUrl || "https://api.retab.com") + params.url + query;
    let headers = params.headers || {};
    let init: RequestInit = {
      method: params.method,
    };
    if (params.method !== "GET") {
      if (params.bodyMime === "multipart/form-data") {
        let formData: FormData = new FormData();
        for (const key of Object.keys(params.body || {})) {
          let value = params.body![key];
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
      throw new APIError(res.status, await res.text());
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
