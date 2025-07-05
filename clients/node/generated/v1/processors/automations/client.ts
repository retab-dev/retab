import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIMailboxesSub from "./mailboxes/client";
import APILinksSub from "./links/client";
import APIOutlookSub from "./outlook/client";
import APIEndpointsSub from "./endpoints/client";
import APILogsSub from "./logs/client";
import APITestsSub from "./tests/client";
import APIAutomationIdSub from "./automationId/client";
import { ListAutomations } from "@/types";

export default class APIAutomations extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  mailboxes = new APIMailboxesSub(this._client);
  links = new APILinksSub(this._client);
  outlook = new APIOutlookSub(this._client);
  endpoints = new APIEndpointsSub(this._client);
  logs = new APILogsSub(this._client);
  tests = new APITestsSub(this._client);
  automationId = new APIAutomationIdSub(this._client);

  async get({ before, after, limit, order, id, webhookUrl, processorId, name }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", id?: string | null, webhookUrl?: string | null, processorId?: string | null, name?: string | null } = {}): Promise<ListAutomations> {
    let res = await this._fetch({
      url: `/v1/processors/automations`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "id": id, "webhook_url": webhookUrl, "processor_id": processorId, "name": name },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
