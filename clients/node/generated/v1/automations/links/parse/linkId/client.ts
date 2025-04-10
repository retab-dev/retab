import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { BodyHandleLinkWebhookV1AutomationsLinksParseLinkIdPost, AutomationLog } from "@/types";

export default class APILinkId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(linkId: string, { idempotencyKey, ...body }: { idempotencyKey?: string | null } & BodyHandleLinkWebhookV1AutomationsLinksParseLinkIdPost): Promise<AutomationLog> {
    let res = await this._fetch({
      url: `/v1/automations/links/parse/${linkId}`,
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey },
      body: body,
      bodyMime: "multipart/form-data",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
