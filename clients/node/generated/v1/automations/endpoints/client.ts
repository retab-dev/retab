import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APILogsSub from "./logs/client";
import APIOpenSub from "./open/client";
import APIEndpointIdSub from "./endpointId/client";
import APIProcessSub from "./process/client";
import { EndpointInput, EndpointOutput, ListEndpoints } from "@/types";

export default class APIEndpoints extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  logs = new APILogsSub(this._client);
  open = new APIOpenSub(this._client);
  endpointId = new APIEndpointIdSub(this._client);
  process = new APIProcessSub(this._client);

  async post({ ...body }: EndpointInput): Promise<EndpointOutput> {
    let res = await this._fetch({
      url: `/v1/automations/endpoints`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async get({ before, after, limit, order, endpointId, name, webhookUrl, schemaId, schemaDataId }: { before?: string | null, after?: string | null, limit?: number | null, order?: "asc" | "desc" | null, endpointId?: string | null, name?: string | null, webhookUrl?: string | null, schemaId?: string | null, schemaDataId?: string | null } = {}): Promise<ListEndpoints> {
    let res = await this._fetch({
      url: `/v1/automations/endpoints`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "endpoint_id": endpointId, "name": name, "webhook_url": webhookUrl, "schema_id": schemaId, "schema_data_id": schemaDataId },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
