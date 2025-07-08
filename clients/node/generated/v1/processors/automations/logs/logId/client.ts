import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIRerunSub from "./rerun/client";
import { ZAutomationLog, AutomationLog } from "@/types";

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
    if (res.headers.get("Content-Type") === "application/json") return ZAutomationLog.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
