import APIGenerated from "@/generated/client";
import { AbstractClient, APIError } from "@/client";
export * from "@/client";

type UiFormClientOptions = {
  baseUrl?: string,
  auth?: { "bearer": string } | { "masterKey": string } | { "apiKey": string },
  basicAuth?: { username: string, password: string },
}

class UiFormClientFetcher implements AbstractClient {
  options: UiFormClientOptions;
  constructor(options?: UiFormClientOptions) {
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
      let bearerToken = this.options.auth?.["bearer"] || process.env["UIFORM_BEARER_TOKEN"];
      let masterKey = this.options.auth?.["masterKey"] || process.env["UIFORM_MASTER_KEY"];
      let apiKey = this.options.auth?.["apiKey"] || process.env["UIFORM_API_KEY"];
      let basicUsername = this.options.basicAuth?.username || process.env["UIFORM_BASIC_USERNAME"];
      let basicPassword = this.options.basicAuth?.password || process.env["UIFORM_BASIC_PASSWORD"];

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
