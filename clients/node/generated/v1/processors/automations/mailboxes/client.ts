import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APITestsSub from "./tests/client";
import APIMailboxIdSub from "./mailboxId/client";
import APIWebhookSub from "./webhook/client";
import { MailboxInput, MailboxOutput, PaginatedList } from "@/types";

export default class APIMailboxes extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  tests = new APITestsSub(this._client);
  mailboxId = new APIMailboxIdSub(this._client);
  webhook = new APIWebhookSub(this._client);

  async post({ ...body }: MailboxInput): Promise<MailboxOutput> {
    let res = await this._fetch({
      url: `/v1/processors/automations/mailboxes`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
  async get({ processorId, before, after, limit, order, email, webhookUrl, schemaId, schemaDataId }: { processorId: string, before?: string | null, after?: string | null, limit?: number | null, order?: "asc" | "desc" | null, email?: string | null, webhookUrl?: string | null, schemaId?: string | null, schemaDataId?: string | null }): Promise<PaginatedList> {
    let res = await this._fetch({
      url: `/v1/processors/automations/mailboxes`,
      method: "GET",
      params: { "processor_id": processorId, "before": before, "after": after, "limit": limit, "order": order, "email": email, "webhook_url": webhookUrl, "schema_id": schemaId, "schema_data_id": schemaDataId },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
