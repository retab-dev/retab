import { AbstractClient, CompositionClient } from '@/client';
import { SubscriptionStatus } from "@/types";

export default class APISubscriptionStatus extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(): Promise<SubscriptionStatus> {
    return this._fetch({
      url: `/v1/iam/payments/subscription-status`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
