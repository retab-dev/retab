import { AbstractClient, CompositionClient } from '@/client';
import APIMailboxes from "./mailboxes/client";
import APILinks from "./links/client";
import APIOutlook from "./outlook/client";
import APITests from "./tests/client";
import APIEndpoints from "./endpoints/client";
import APILogs from "./logs/client";
import APIAutomationId from "./automationId/client";
import APIReviewExtraction from "./reviewExtraction/client";
import { ListAutomations } from "@/types";

export default class APIAutomations extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  mailboxes = new APIMailboxes(this);
  links = new APILinks(this);
  outlook = new APIOutlook(this);
  tests = new APITests(this);
  endpoints = new APIEndpoints(this);
  logs = new APILogs(this);
  automationId = new APIAutomationId(this);
  reviewExtraction = new APIReviewExtraction(this);

  async get({ before, after, limit, order, automationId, webhookUrl, model, schemaId, schemaDataId }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", automationId?: string | null, webhookUrl?: string | null, model?: string | null, schemaId?: string | null, schemaDataId?: string | null }): Promise<ListAutomations> {
    return this._fetch({
      url: `/v1/automations`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "automation_id": automationId, "webhook_url": webhookUrl, "model": model, "schema_id": schemaId, "schema_data_id": schemaDataId },
      headers: {  },
    });
  }
  
}
