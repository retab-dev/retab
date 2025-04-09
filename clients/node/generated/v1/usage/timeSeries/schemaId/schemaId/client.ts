import { AbstractClient, CompositionClient } from '@/client';
import { TimeRange, UsageTimeSeries } from "@/types";

export default class APISchemaId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(schemaId: string, { timeRange }: { timeRange: TimeRange }): Promise<UsageTimeSeries> {
    return this._fetch({
      url: `/v1/usage/time_series/schema_id/${schemaId}`,
      method: "GET",
      params: { "time_range": timeRange },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
