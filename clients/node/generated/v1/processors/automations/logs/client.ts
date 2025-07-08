import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APILogIdSub from "./logId/client";
import { ZListLogs, ListLogs } from "@/types";

export default class APILogs extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  logId = new APILogIdSub(this._client);

  async get({ before, after, limit, order, automationId, processorId, webhookUrl, schemaId, schemaDataId, statusCode, statusClass }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", automationId?: string | null, processorId?: string | null, webhookUrl?: string | null, schemaId?: string | null, schemaDataId?: string | null, statusCode?: number | null, statusClass?: "2xx" | "3xx" | "4xx" | "5xx" | null } = {}): Promise<ListLogs> {
    let res = await this._fetch({
      url: `/v1/processors/automations/logs`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "automation_id": automationId, "processor_id": processorId, "webhook_url": webhookUrl, "schema_id": schemaId, "schema_data_id": schemaDataId, "status_code": statusCode, "status_class": statusClass },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZListLogs.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
