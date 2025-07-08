import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZTimeRange, TimeRange, ZCreditsTimeSeries, CreditsTimeSeries } from "@/types";

export default class APITimeSeriesPreprocessing extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get({ timeRange }: { timeRange: TimeRange }): Promise<CreditsTimeSeries> {
    let res = await this._fetch({
      url: `/v1/usage/time_series_preprocessing`,
      method: "GET",
      params: { "time_range": timeRange },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZCreditsTimeSeries.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
