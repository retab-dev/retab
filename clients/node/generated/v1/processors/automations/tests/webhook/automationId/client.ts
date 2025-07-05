import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { AutomationLog } from "@/types";

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
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
