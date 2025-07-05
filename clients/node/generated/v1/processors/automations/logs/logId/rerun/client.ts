import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { ExternalRequestLog } from "@/types";

export default class APIRerun extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(logId: string): Promise<ExternalRequestLog> {
    let res = await this._fetch({
      url: `/v1/processors/automations/logs/${logId}/rerun`,
      method: "POST",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
