import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZTimeRange, TimeRange, ZUsageTimeSeries, UsageTimeSeries } from "@/types";

export default class APITimeSeriesOrganization extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get({ timeRange }: { timeRange: TimeRange }): Promise<UsageTimeSeries> {
    let res = await this._fetch({
      url: `/v1/usage/time_series_organization`,
      method: "GET",
      params: { "time_range": timeRange },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZUsageTimeSeries.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
