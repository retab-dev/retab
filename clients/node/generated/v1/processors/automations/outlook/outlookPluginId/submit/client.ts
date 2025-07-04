import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { OutlookSubmitRequest, AutomationLog } from "@/types";

export default class APISubmit extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(outlookPluginId: string, { idempotencyKey, ...body }: { idempotencyKey?: string | null } & OutlookSubmitRequest): Promise<AutomationLog> {
    let res = await this._fetch({
      url: `/v1/processors/automations/outlook/${outlookPluginId}/submit`,
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey },
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
