import { AbstractClient, CompositionClient } from '@/client';
import { AutomationLog } from "@/types";

export default class APILogId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(logId: string): Promise<AutomationLog> {
    return this._fetch({
      url: `/v1/automations/mailboxes/logs/${logId}`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
