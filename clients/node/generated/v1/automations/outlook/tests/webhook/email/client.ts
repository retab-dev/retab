import { AbstractClient, CompositionClient } from '@/client';
import { AutomationLog } from "@/types";

export default class APIEmail extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(email: string): Promise<AutomationLog> {
    return this._fetch({
      url: `/v1/automations/outlook/tests/webhook/${email}`,
      method: "POST",
    });
  }
  
}
