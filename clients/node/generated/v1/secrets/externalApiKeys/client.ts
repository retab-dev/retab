import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIProviderSub from "./provider/client";
import APICheckOpenaiKeyValiditySub from "./checkOpenaiKeyValidity/client";
import { ExternalAPIKey, ExternalAPIKeyRequest, ExternalAPIKeyRequest } from "@/types";

export default class APIExternalApiKeys extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  provider = new APIProviderSub(this._client);
  checkOpenaiKeyValidity = new APICheckOpenaiKeyValiditySub(this._client);

  async get(): Promise<ExternalAPIKey[]> {
    let res = await this._fetch({
      url: `/v1/secrets/external_api_keys`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async put({ ...body }: ExternalAPIKeyRequest): Promise<object> {
    let res = await this._fetch({
      url: `/v1/secrets/external_api_keys`,
      method: "PUT",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async post({ ...body }: ExternalAPIKeyRequest): Promise<object> {
    let res = await this._fetch({
      url: `/v1/secrets/external_api_keys`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
