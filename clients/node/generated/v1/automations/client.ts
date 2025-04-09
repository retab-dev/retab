import { AbstractClient, CompositionClient } from '@/client';
import APIMailboxesSub from "./mailboxes/client";
import APILinksSub from "./links/client";
import APIOutlookSub from "./outlook/client";
import APITestsSub from "./tests/client";
import APIEndpointsSub from "./endpoints/client";
import APILogsSub from "./logs/client";
import APIAutomationIdSub from "./automationId/client";
import APIReviewExtractionSub from "./reviewExtraction/client";
import { ListAutomations } from "@/types";

export default class APIAutomations extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  mailboxes = new APIMailboxesSub(this._client);
  links = new APILinksSub(this._client);
  outlook = new APIOutlookSub(this._client);
  tests = new APITestsSub(this._client);
  endpoints = new APIEndpointsSub(this._client);
  logs = new APILogsSub(this._client);
  automationId = new APIAutomationIdSub(this._client);
  reviewExtraction = new APIReviewExtractionSub(this._client);

  async get({ before, after, limit, order, automationId, webhookUrl, model, schemaId, schemaDataId }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", automationId?: string | null, webhookUrl?: string | null, model?: string | null, schemaId?: string | null, schemaDataId?: string | null } = {}): Promise<ListAutomations> {
    return this._fetch({
      url: `/v1/automations`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "automation_id": automationId, "webhook_url": webhookUrl, "model": model, "schema_id": schemaId, "schema_data_id": schemaDataId },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
