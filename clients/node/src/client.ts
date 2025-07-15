import * as z from "zod";

export class AbstractClient {
  protected _fetch(params: {
    url: string,
    method: string,
    params?: Record<string, any>,
    headers?: Record<string, any>,
    bodyMime?: "application/json" | "multipart/form-data",
    body?: Record<string, any>,
    auth?: string[],
  }): Promise<Response> {
    throw new Error("Method not implemented");
  }
}

export class CompositionClient extends AbstractClient {
  protected _client: AbstractClient;
  constructor(client: AbstractClient) {
    super();
    this._client = client;
  }
  protected _fetch(params: {
    url: string,
    method: string,
    params?: Record<string, any>,
    headers?: Record<string, any>,
    bodyMime?: "application/json" | "multipart/form-data",
    body?: Record<string, any>,
    auth?: string[],
  }): Promise<Response> {
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
type UiFormClientOptions = {
  baseUrl?: string,
  basicAuth?: { username: string, password: string },
} & AuthTypes;

class UiFormClientFetcher extends AbstractClient {
  options: UiFormClientOptions;
  constructor(options?: UiFormClientOptions) {
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
    auth?: string[];
  }): Promise<Response> {
    let query = "";
    if (params.params) {
      query = "?" + new URLSearchParams(
        Object.fromEntries(
          Object.entries(params.params).filter(([k, v]) => v !== undefined)
        )
      ).toString();
    }
    let url = (this.options.baseUrl || "https://api.uiform.com") + params.url + query;
    let headers = params.headers || {};
    let init: RequestInit = {
      method: params.method,
    };
    if (params.method !== "GET") {
      headers["Content-Type"] = params.bodyMime;
      if (params.bodyMime === "application/json") {
        init.body = JSON.stringify(params.body);
      } else {
        let formData: FormData = new FormData();
        for (const key of Object.keys(params.body || {})) {
          formData.append(key, params.body![key]);
        }
        init.body = formData;
      }
    }
    if (params.auth) {
      let bearerToken = "bearer" in this.options ? this.options.bearer : process.env["UIFORM_BEARER_TOKEN"];
      let masterKey = "masterKey" in this.options ? this.options.masterKey : process.env["UIFORM_MASTER_KEY"];
      let apiKey = "apiKey" in this.options ? this.options.apiKey : process.env["UIFORM_API_KEY"];
      let basicUsername = this.options.basicAuth ? this.options.basicAuth.username : process.env["UIFORM_BASIC_USERNAME"];
      let basicPassword = this.options.basicAuth ? this.options.basicAuth.password : process.env["UIFORM_BASIC_PASSWORD"];

      if (params.auth.includes("HTTPBearer") && bearerToken) {
        headers["Authorization"] = `Bearer ${bearerToken}`;
      } else if (params.auth.includes("Master Key") && masterKey) {
        headers["Master-Key"] = masterKey;
      } else if (params.auth.includes("API Key") && apiKey) {
        headers["Api-Key"] = apiKey;
      } else if (params.auth.includes("HTTPBasic") && basicUsername && basicPassword) {
        headers["Authorization"] = `Basic ${btoa(`${basicUsername}:${basicPassword}`)}`;
      }
    }
    init.headers = headers;
    let res = await fetch(url, init);
    if (!res.ok) {
      throw new APIError(res.status, await res.text());
    }
    return res;
  }
}

export class UiFormClient extends APIGenerated {
  constructor(options?: UiFormClientOptions) {
    super(new UiFormClientFetcher(options));
  }
}
