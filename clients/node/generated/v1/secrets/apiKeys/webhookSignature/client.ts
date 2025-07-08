import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZWebhookSignature, WebhookSignature } from "@/types";

export default class APIWebhookSignature extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(): Promise<object> {
    let res = await this._fetch({
      url: `/v1/secrets/api_keys/webhook-signature`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return z.object({}).parse(await res.json());
    throw new Error("Bad content type");
  }
  
  async put(): Promise<WebhookSignature> {
    let res = await this._fetch({
      url: `/v1/secrets/api_keys/webhook-signature`,
      method: "PUT",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZWebhookSignature.parse(await res.json());
    throw new Error("Bad content type");
  }
  
  async post(): Promise<WebhookSignature> {
    let res = await this._fetch({
      url: `/v1/secrets/api_keys/webhook-signature`,
      method: "POST",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZWebhookSignature.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
