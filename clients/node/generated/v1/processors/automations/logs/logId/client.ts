import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIRerunSub from "./rerun/client";
import { AutomationLog } from "@/types";

export default class APILogId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  rerun = new APIRerunSub(this._client);

  async get(logId: string): Promise<AutomationLog> {
    let res = await this._fetch({
      url: `/v1/processors/automations/logs/${logId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
