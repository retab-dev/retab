import { AbstractClient, CompositionClient } from '@/client';
import { AutomationLog } from "@/types";

export default class APILinkId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(linkId: string): Promise<AutomationLog> {
    return this._fetch({
      url: `/v1/automations/links/tests/webhook/${linkId}`,
      method: "POST",
      params: {  },
      headers: {  },
    });
  }
  
}
