import { AbstractClient, CompositionClient } from '@/client';
import { WebhookSignature, WebhookSignature } from "@/types";

export default class APIWebhookSignature extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(): Promise<object> {
    return this._fetch({
      url: `/v1/secrets/api_keys/webhook-signature`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
  async put(): Promise<WebhookSignature> {
    return this._fetch({
      url: `/v1/secrets/api_keys/webhook-signature`,
      method: "PUT",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
  async post(): Promise<WebhookSignature> {
    return this._fetch({
      url: `/v1/secrets/api_keys/webhook-signature`,
      method: "POST",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
