import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { SubscriptionStatus } from "@/types";

export default class APISubscriptionStatus extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(): Promise<SubscriptionStatus> {
    let res = await this._fetch({
      url: `/v1/iam/payments/subscription-status`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
