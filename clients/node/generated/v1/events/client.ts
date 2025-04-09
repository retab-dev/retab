import { AbstractClient, CompositionClient } from '@/client';
import APIEventIdSub from "./eventId/client";
import { PaginatedList } from "@/types";

export default class APIEvents extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  eventId = new APIEventIdSub(this._client);

  async get({ before, after, limit, order, id, name, schemaId, schemaDataId }: { before?: string | null, after?: string | null, limit?: number | null, order?: "asc" | "desc" | null, id?: string | null, name?: string | null, schemaId?: string | null, schemaDataId?: string | null } = {}): Promise<PaginatedList> {
    return this._fetch({
      url: `/v1/events`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "id": id, "name": name, "schema_id": schemaId, "schema_data_id": schemaDataId },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
