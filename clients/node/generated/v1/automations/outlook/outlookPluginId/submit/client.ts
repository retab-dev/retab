import { AbstractClient, CompositionClient } from '@/client';
import { OutlookSubmitRequest, AutomationLog } from "@/types";

export default class APISubmit extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(outlookPluginId: string, { IdempotencyKey, ...body }: { IdempotencyKey?: string | null } & OutlookSubmitRequest): Promise<AutomationLog> {
    return this._fetch({
      url: `/v1/automations/outlook/${outlookPluginId}/submit`,
      method: "POST",
      params: {  },
      headers: { "Idempotency-Key": IdempotencyKey },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
