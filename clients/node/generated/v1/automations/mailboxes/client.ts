import { AbstractClient, CompositionClient } from '@/client';
import APITestsSub from "./tests/client";
import APILogsSub from "./logs/client";
import APIEmailSub from "./email/client";
import APIFromIdSub from "./fromId/client";
import APIWebhookSub from "./webhook/client";
import { MailboxInput, MailboxOutput, PaginatedList } from "@/types";

export default class APIMailboxes extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  tests = new APITestsSub(this._client);
  logs = new APILogsSub(this._client);
  email = new APIEmailSub(this._client);
  fromId = new APIFromIdSub(this._client);
  webhook = new APIWebhookSub(this._client);

  async post({ ...body }: MailboxInput): Promise<MailboxOutput> {
    return this._fetch({
      url: `/v1/automations/mailboxes`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
  async get({ before, after, limit, order, email, webhookUrl, schemaId, schemaDataId }: { before?: string | null, after?: string | null, limit?: number | null, order?: "asc" | "desc" | null, email?: string | null, webhookUrl?: string | null, schemaId?: string | null, schemaDataId?: string | null } = {}): Promise<PaginatedList> {
    return this._fetch({
      url: `/v1/automations/mailboxes`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "email": email, "webhook_url": webhookUrl, "schema_id": schemaId, "schema_data_id": schemaDataId },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
