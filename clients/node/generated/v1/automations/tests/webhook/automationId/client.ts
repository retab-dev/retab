import { AbstractClient, CompositionClient } from '@/client';
import { AutomationLog } from "@/types";

export default class APIAutomationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(automationId: string): Promise<AutomationLog> {
    return this._fetch({
      url: `/v1/automations/tests/webhook/${automationId}`,
      method: "POST",
    });
  }
  
}
