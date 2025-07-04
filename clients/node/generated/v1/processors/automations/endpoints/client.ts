import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIEndpointIdSub from "./endpointId/client";
import APIProcessSub from "./process/client";
import { EndpointInput, EndpointOutput, ListEndpoints } from "@/types";

export default class APIEndpoints extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  endpointId = new APIEndpointIdSub(this._client);
  process = new APIProcessSub(this._client);

  async post({ ...body }: EndpointInput): Promise<EndpointOutput> {
    let res = await this._fetch({
      url: `/v1/processors/automations/endpoints`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async get({ before, after, limit, order, id, name, webhookUrl, processorId }: { before?: string | null, after?: string | null, limit?: number | null, order?: "asc" | "desc" | null, id?: string | null, name?: string | null, webhookUrl?: string | null, processorId?: string | null } = {}): Promise<ListEndpoints> {
    let res = await this._fetch({
      url: `/v1/processors/automations/endpoints`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "id": id, "name": name, "webhook_url": webhookUrl, "processor_id": processorId },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
