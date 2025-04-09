import { AbstractClient, CompositionClient } from '@/client';
import { MonthlyUsageResponseContent } from "@/types";

export default class APIMonthlyUsage extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(): Promise<MonthlyUsageResponseContent> {
    return this._fetch({
      url: `/v1/documents/monthly-usage`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
