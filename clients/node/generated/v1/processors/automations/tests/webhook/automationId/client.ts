import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZAutomationLog, AutomationLog } from "@/types";

export default class APIAutomationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(automationId: string): Promise<AutomationLog> {
    let res = await this._fetch({
      url: `/v1/processors/automations/tests/webhook/${automationId}`,
      method: "POST",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZAutomationLog.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
