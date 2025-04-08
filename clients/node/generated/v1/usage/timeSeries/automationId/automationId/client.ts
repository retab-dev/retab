import { AbstractClient, CompositionClient } from '@/client';
import { TimeRange, UsageTimeSeries } from "@/types";

export default class APIAutomationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(automationId: string, { timeRange }: { timeRange: TimeRange }): Promise<UsageTimeSeries> {
    return this._fetch({
      url: `/v1/usage/time_series/automation_id/${automationId}`,
      method: "GET",
      params: { "time_range": timeRange },
      headers: {  },
    });
  }
  
}
