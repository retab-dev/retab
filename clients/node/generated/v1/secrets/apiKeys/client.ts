import { AbstractClient, CompositionClient } from '@/client';
import APIKeyId from "./keyId/client";
import APIWebhookSignature from "./webhookSignature/client";
import APIGetOneApiKey from "./getOneApiKey/client";
import { APIKeyInfo, APIKeyCreate, APIKeyResponse } from "@/types";

export default class APIApiKeys extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  keyId = new APIKeyId(this);
  webhookSignature = new APIWebhookSignature(this);
  getOneApiKey = new APIGetOneApiKey(this);

  async get(): Promise<APIKeyInfo[]> {
    return this._fetch({
      url: `/v1/secrets/api_keys`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
  async post({ ...body }: APIKeyCreate): Promise<APIKeyResponse> {
    return this._fetch({
      url: `/v1/secrets/api_keys`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
