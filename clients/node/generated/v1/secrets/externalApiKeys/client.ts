import { AbstractClient, CompositionClient } from '@/client';
import APIProvider from "./provider/client";
import APICheckOpenaiKeyValidity from "./checkOpenaiKeyValidity/client";
import { ExternalAPIKey, ExternalAPIKeyRequest, ExternalAPIKeyRequest } from "@/types";

export default class APIExternalApiKeys extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  provider = new APIProvider(this);
  checkOpenaiKeyValidity = new APICheckOpenaiKeyValidity(this);

  async get(): Promise<ExternalAPIKey[]> {
    return this._fetch({
      url: `/v1/secrets/external_api_keys`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
  async put({ ...body }: ExternalAPIKeyRequest): Promise<object> {
    return this._fetch({
      url: `/v1/secrets/external_api_keys`,
      method: "PUT",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
  async post({ ...body }: ExternalAPIKeyRequest): Promise<object> {
    return this._fetch({
      url: `/v1/secrets/external_api_keys`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
