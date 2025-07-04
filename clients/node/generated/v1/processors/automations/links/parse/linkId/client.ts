import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { BodyHandleLinkWebhookV1ProcessorsAutomationsLinksParseLinkIdPost, AutomationLog } from "@/types";

export default class APILinkId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(linkId: string, { idempotencyKey, ...body }: { idempotencyKey?: string | null } & BodyHandleLinkWebhookV1ProcessorsAutomationsLinksParseLinkIdPost): Promise<AutomationLog> {
    let res = await this._fetch({
      url: `/v1/processors/automations/links/parse/${linkId}`,
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey },
      body: body,
      bodyMime: "multipart/form-data",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
