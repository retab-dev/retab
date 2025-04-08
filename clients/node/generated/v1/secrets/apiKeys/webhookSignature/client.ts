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
      params: {  },
      headers: {  },
    });
  }
  
  async put(): Promise<WebhookSignature> {
    return this._fetch({
      url: `/v1/secrets/api_keys/webhook-signature`,
      method: "PUT",
      params: {  },
      headers: {  },
    });
  }
  
  async post(): Promise<WebhookSignature> {
    return this._fetch({
      url: `/v1/secrets/api_keys/webhook-signature`,
      method: "POST",
      params: {  },
      headers: {  },
    });
  }
  
}
