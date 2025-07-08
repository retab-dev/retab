import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APICreditsSub from "./credits/client";
import { ZTimeRange, TimeRange, ZUsageTimeSeries, UsageTimeSeries } from "@/types";

export default class APIAutomationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  credits = new APICreditsSub(this._client);

  async get(automationId: string, { timeRange }: { timeRange: TimeRange }): Promise<UsageTimeSeries> {
    let res = await this._fetch({
      url: `/v1/usage/time_series/automation_id/${automationId}`,
      method: "GET",
      params: { "time_range": timeRange },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZUsageTimeSeries.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
