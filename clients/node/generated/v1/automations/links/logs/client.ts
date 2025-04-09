import { AbstractClient, CompositionClient } from '@/client';
import APILogIdSub from "./logId/client";
import { ListLogs } from "@/types";

export default class APILogs extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  logId = new APILogIdSub(this._client);

  async get({ before, after, limit, order, linkId, name, webhookUrl, schemaId, schemaDataId, sortBy }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", linkId?: string | null, name?: string | null, webhookUrl?: string | null, schemaId?: string | null, schemaDataId?: string | null, sortBy?: string } = {}): Promise<ListLogs> {
    return this._fetch({
      url: `/v1/automations/links/logs`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "link_id": linkId, "name": name, "webhook_url": webhookUrl, "schema_id": schemaId, "schema_data_id": schemaDataId, "sort_by": sortBy },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
