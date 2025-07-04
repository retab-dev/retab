import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { AutomationLog } from "@/types";

export default class APILogId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(logId: string): Promise<AutomationLog> {
    let res = await this._fetch({
      url: `/v1/automations/endpoints/logs/${logId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
