import * as z from "zod";

type FetchParams = {
  url: string,
  method: string,
  params?: Record<string, any>,
  headers?: Record<string, any>,
  bodyMime?: "application/json" | "multipart/form-data",
  body?: Record<string, any>,
  auth?: string[],
};
export class AbstractClient {
  protected _fetch(_: FetchParams): Promise<Response> {
    throw new Error("Method not implemented");
  }
  protected async _fetchJson<ZodSchema extends z.ZodType<any, any, any>>(bodyType: ZodSchema, params: FetchParams): Promise<z.output<ZodSchema>> {
    let response = await this._fetch(params);
    if (!response.ok) {
      throw new APIError(response.status, await response.text());
    }
    if (response.headers.get("Content-Type") !== "application/json") throw new APIError(response.status, "Response is not JSON");
    return bodyType.parse(await response.json());
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

export async function* streamResponse<ZodSchema extends z.ZodType<any, any, any>>(response: Response, schema: ZodSchema): AsyncGenerator<z.output<ZodSchema>> {
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
    let string = new TextDecoder().decode(chunk.value, {stream: true});
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
            yield schema.parse(JSON.parse(body.slice(0, prevBodyLength + i + 1)));
            body = body.slice(prevBodyLength + i + 1);
            prevBodyLength = -i - 1;
          }
        }
      }
    }
    if (chunk.done) break;
  }
}

export const DateOrISO = z.union([
  z.date(),
  z.string().refine(val => !isNaN(Date.parse(val)), {
    message: "Invalid date string",
  }).transform(val => new Date(val)),
]);

type AuthTypes = { "bearer": string } | { "masterKey": string } | { "apiKey": string } | {};
export type ClientOptions = {
  baseUrl?: string,
} & AuthTypes;

export class FetcherClient extends AbstractClient {
  options: ClientOptions;
  constructor(options?: ClientOptions) {
    super();
    this.options = options || {};
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
      headers["Content-Type"] = params.bodyMime || "application/json";
      if (params.bodyMime === "multipart/form-data") {
        let formData: FormData = new FormData();
        for (const key of Object.keys(params.body || {})) {
          formData.append(key, params.body![key]);
        }
        init.body = formData;
      } else {
        init.body = JSON.stringify(params.body);
      }
    }
    let bearerToken = "bearer" in this.options ? this.options.bearer : process.env["RETAB_BEARER_TOKEN"];
    let masterKey = "masterKey" in this.options ? this.options.masterKey : process.env["RETAB_MASTER_KEY"];
    let apiKey = "apiKey" in this.options ? this.options.apiKey : process.env["RETAB_API_KEY"];

    if (bearerToken) {
      headers["Authorization"] = `Bearer ${bearerToken}`;
    } else if (masterKey) {
      headers["Master-Key"] = masterKey;
    } else if (apiKey) {
      headers["Api-Key"] = apiKey;
    }
    init.headers = headers;
    let res = await fetch(url, init);
    if (!res.ok) {
      throw new APIError(res.status, await res.text());
    }
    return res;
  }
}
