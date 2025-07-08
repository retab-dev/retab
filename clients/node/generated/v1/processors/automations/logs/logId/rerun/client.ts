import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZExternalRequestLog, ExternalRequestLog } from "@/types";

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
    if (res.headers.get("Content-Type") === "application/json") return ZExternalRequestLog.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
