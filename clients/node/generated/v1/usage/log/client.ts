import { AbstractClient, CompositionClient } from '@/client';
import { LogCompletionRequest, AutomationLog } from "@/types";

export default class APILog extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: LogCompletionRequest): Promise<AutomationLog> {
    return this._fetch({
      url: `/v1/usage/log`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
