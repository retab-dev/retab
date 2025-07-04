import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { AutomationLog } from "@/types";

export default class APIAutomationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(automationId: string): Promise<AutomationLog> {
    let res = await this._fetch({
      url: `/v1/automations/tests/webhook/${automationId}`,
      method: "POST",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
