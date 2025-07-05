import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { Credits } from "@/types";

export default class APICredits extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(processorId: string, { startDate, endDate }: { startDate?: Date | null, endDate?: Date | null } = {}): Promise<Credits> {
    let res = await this._fetch({
      url: `/v1/usage/processor/${processorId}/credits`,
      method: "GET",
      params: { "start_date": startDate, "end_date": endDate },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
