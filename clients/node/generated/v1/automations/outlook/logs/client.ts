import { AbstractClient, CompositionClient } from '@/client';
import APILogId from "./logId/client";
import { ListLogs } from "@/types";

export default class APILogs extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  logId = new APILogId(this);

  async get({ before, after, limit, order, outlookPluginId, name, webhookUrl, schemaId, schemaDataId }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", outlookPluginId?: string | null, name?: string | null, webhookUrl?: string | null, schemaId?: string | null, schemaDataId?: string | null }): Promise<ListLogs> {
    return this._fetch({
      url: `/v1/automations/outlook/logs`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "outlook_plugin_id": outlookPluginId, "name": name, "webhook_url": webhookUrl, "schema_id": schemaId, "schema_data_id": schemaDataId },
      headers: {  },
    });
  }
  
}
