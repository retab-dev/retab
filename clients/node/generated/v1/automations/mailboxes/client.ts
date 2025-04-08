import { AbstractClient, CompositionClient } from '@/client';
import APITests from "./tests/client";
import APILogs from "./logs/client";
import APIEmail from "./email/client";
import APIFromId from "./fromId/client";
import APIWebhook from "./webhook/client";
import { MailboxInput, MailboxOutput, PaginatedList } from "@/types";

export default class APIMailboxes extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  tests = new APITests(this);
  logs = new APILogs(this);
  email = new APIEmail(this);
  fromId = new APIFromId(this);
  webhook = new APIWebhook(this);

  async post({ ...body }: MailboxInput): Promise<MailboxOutput> {
    return this._fetch({
      url: `/v1/automations/mailboxes`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
  async get({ before, after, limit, order, email, webhookUrl, schemaId, schemaDataId }: { before?: string | null, after?: string | null, limit?: number | null, order?: "asc" | "desc" | null, email?: string | null, webhookUrl?: string | null, schemaId?: string | null, schemaDataId?: string | null }): Promise<PaginatedList> {
    return this._fetch({
      url: `/v1/automations/mailboxes`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "email": email, "webhook_url": webhookUrl, "schema_id": schemaId, "schema_data_id": schemaDataId },
      headers: {  },
    });
  }
  
}
