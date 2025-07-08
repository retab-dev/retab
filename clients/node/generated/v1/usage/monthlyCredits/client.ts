import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZMonthlyUsageResponseContent, MonthlyUsageResponseContent } from "@/types";

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
    if (res.headers.get("Content-Type") === "application/json") return ZMonthlyUsageResponseContent.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
