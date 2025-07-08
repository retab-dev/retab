import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZCredits, Credits } from "@/types";

export default class APICredits extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(automationId: string, { startDate, endDate }: { startDate?: Date | null, endDate?: Date | null } = {}): Promise<Credits> {
    let res = await this._fetch({
      url: `/v1/usage/automation/${automationId}/credits`,
      method: "GET",
      params: { "start_date": startDate, "end_date": endDate },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZCredits.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
