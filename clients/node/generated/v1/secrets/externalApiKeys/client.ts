import { AbstractClient, CompositionClient } from '@/client';
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
    return this._fetch({
      url: `/v1/secrets/external_api_keys`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
  async put({ ...body }: ExternalAPIKeyRequest): Promise<object> {
    return this._fetch({
      url: `/v1/secrets/external_api_keys`,
      method: "PUT",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
  async post({ ...body }: ExternalAPIKeyRequest): Promise<object> {
    return this._fetch({
      url: `/v1/secrets/external_api_keys`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
