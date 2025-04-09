import { AbstractClient, CompositionClient } from '@/client';
import { Amount } from "@/types";

export default class APIAutomationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(automationId: string, { startDate, endDate }: { startDate?: Date | null, endDate?: Date | null } = {}): Promise<Amount> {
    return this._fetch({
      url: `/v1/usage/automation/${automationId}`,
      method: "GET",
      params: { "start_date": startDate, "end_date": endDate },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
