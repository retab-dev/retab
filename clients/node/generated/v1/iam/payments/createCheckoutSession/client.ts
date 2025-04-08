import { AbstractClient, CompositionClient } from '@/client';

export default class APICreateCheckoutSession extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(): Promise<object> {
    return this._fetch({
      url: `/v1/iam/payments/create-checkout-session`,
      method: "POST",
      params: {  },
      headers: {  },
    });
  }
  
}
