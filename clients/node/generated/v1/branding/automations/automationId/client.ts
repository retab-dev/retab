import { AbstractClient, CompositionClient } from '@/client';
import { Branding } from "@/types";

export default class APIAutomationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(automationId: string): Promise<Branding> {
    return this._fetch({
      url: `/v1/branding/automations/${automationId}`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
