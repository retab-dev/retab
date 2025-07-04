import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APICreditsSub from "./credits/client";
import { TimeRange, UsageTimeSeries } from "@/types";

export default class APIProcessorId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  credits = new APICreditsSub(this._client);

  async get(processorId: string, { timeRange }: { timeRange: TimeRange }): Promise<UsageTimeSeries> {
    let res = await this._fetch({
      url: `/v1/usage/time_series/processor_id/${processorId}`,
      method: "GET",
      params: { "time_range": timeRange },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
