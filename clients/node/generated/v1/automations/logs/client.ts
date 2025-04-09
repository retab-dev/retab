import { AbstractClient, CompositionClient } from '@/client';
import { ListLogs } from "@/types";

export default class APILogs extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get({ before, after, limit, order, automationId, webhookUrl, schemaId, schemaDataId }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", automationId?: string | null, webhookUrl?: string | null, schemaId?: string | null, schemaDataId?: string | null } = {}): Promise<ListLogs> {
    return this._fetch({
      url: `/v1/automations/logs`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "automation_id": automationId, "webhook_url": webhookUrl, "schema_id": schemaId, "schema_data_id": schemaDataId },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
