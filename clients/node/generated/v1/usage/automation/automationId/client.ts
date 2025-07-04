import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APICreditsSub from "./credits/client";
import { Amount } from "@/types";

export default class APIAutomationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  credits = new APICreditsSub(this._client);

  async get(automationId: string, { startDate, endDate }: { startDate?: Date | null, endDate?: Date | null } = {}): Promise<Amount> {
    let res = await this._fetch({
      url: `/v1/usage/automation/${automationId}`,
      method: "GET",
      params: { "start_date": startDate, "end_date": endDate },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
