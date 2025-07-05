import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIKeyIdSub from "./keyId/client";
import APIWebhookSignatureSub from "./webhookSignature/client";
import APIGetOneApiKeySub from "./getOneApiKey/client";
import { APIKeyInfo, APIKeyCreate, APIKeyResponse } from "@/types";

export default class APIApiKeys extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  keyId = new APIKeyIdSub(this._client);
  webhookSignature = new APIWebhookSignatureSub(this._client);
  getOneApiKey = new APIGetOneApiKeySub(this._client);

  async get(): Promise<APIKeyInfo[]> {
    let res = await this._fetch({
      url: `/v1/secrets/api_keys`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
  async post({ ...body }: APIKeyCreate): Promise<APIKeyResponse> {
    let res = await this._fetch({
      url: `/v1/secrets/api_keys`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
