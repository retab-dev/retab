import { AbstractClient, CompositionClient } from '@/client';
import { BodyHandleLinkWebhookV1AutomationsLinksParseLinkIdPost, AutomationLog } from "@/types";

export default class APILinkId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(linkId: string, { IdempotencyKey, ...body }: { IdempotencyKey?: string | null } & BodyHandleLinkWebhookV1AutomationsLinksParseLinkIdPost): Promise<AutomationLog> {
    return this._fetch({
      url: `/v1/automations/links/parse/${linkId}`,
      method: "POST",
      params: {  },
      headers: { "Idempotency-Key": IdempotencyKey },
      body: body,
      bodyMime: "multipart/form-data",
    });
  }
  
}
