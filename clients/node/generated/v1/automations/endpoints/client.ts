import { AbstractClient, CompositionClient } from '@/client';
import APILogs from "./logs/client";
import APIOpen from "./open/client";
import APIEndpointId from "./endpointId/client";
import APIProcess from "./process/client";
import { EndpointInput, EndpointOutput, ListEndpoints } from "@/types";

export default class APIEndpoints extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  logs = new APILogs(this);
  open = new APIOpen(this);
  endpointId = new APIEndpointId(this);
  process = new APIProcess(this);

  async post({ ...body }: EndpointInput): Promise<EndpointOutput> {
    return this._fetch({
      url: `/v1/automations/endpoints`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
  async get({ before, after, limit, order, endpointId, name, webhookUrl, schemaId, schemaDataId }: { before?: string | null, after?: string | null, limit?: number | null, order?: "asc" | "desc" | null, endpointId?: string | null, name?: string | null, webhookUrl?: string | null, schemaId?: string | null, schemaDataId?: string | null }): Promise<ListEndpoints> {
    return this._fetch({
      url: `/v1/automations/endpoints`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "endpoint_id": endpointId, "name": name, "webhook_url": webhookUrl, "schema_id": schemaId, "schema_data_id": schemaDataId },
      headers: {  },
    });
  }
  
}
