import { AbstractClient, CompositionClient } from '@/client';

export default class APIWebhook extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(): Promise<object> {
    return this._fetch({
      url: `/v1/iam/payments/webhook`,
      method: "POST",
      params: {  },
      headers: {  },
    });
  }
  
}
