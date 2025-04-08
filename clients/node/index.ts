import APIGenerated from "@/generated/client";
import { AbstractClient, APIError } from "@/client";
export * from "@/client";

type UiFormClientOptions = {
  baseUrl?: string;
  auth?: { "bearer": string } | { "masterKey": string } | { "apiKey": string };
}

class UiFormClientFetcher implements AbstractClient {
  options: UiFormClientOptions;
  constructor(options: UiFormClientOptions) {
    this.options = options;
  }

  async _fetch<T>(params: {
    url: string;
    method: string;
    params?: Record<string, any>;
    headers?: Record<string, any>;
    bodyMime?: "application/json" | "multipart/form-data";
    body?: Record<string, any>
  }): Promise<T> {
    let query = "";
    if (params.params) {
      query = "?" + new URLSearchParams(params.params).toString();
    }
    let url = (this.options.baseUrl || "https://uiform.com") + params.url + query;
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
    init.headers = headers;
    let res = await fetch(url, init);
    if (res.status >= 200 && res.status < 300) {
      return res.json();
    }
    throw new APIError(res.status, await res.json());
  }
}

export class UiFormClient extends APIGenerated {
  constructor(options: UiFormClientOptions) {
    super(new UiFormClientFetcher(options));
  }
}
