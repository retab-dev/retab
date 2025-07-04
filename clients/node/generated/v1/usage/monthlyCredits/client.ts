import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { MonthlyUsageResponseContent } from "@/types";

export default class APIMonthlyCredits extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(): Promise<MonthlyUsageResponseContent> {
    let res = await this._fetch({
      url: `/v1/usage/monthly_credits`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
