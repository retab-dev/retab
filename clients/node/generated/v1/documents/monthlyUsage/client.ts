import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { MonthlyUsageResponseContent } from "@/types";

export default class APIMonthlyUsage extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(): Promise<MonthlyUsageResponseContent> {
    let res = await this._fetch({
      url: `/v1/documents/monthly-usage`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
