import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIModelIdSub from "./modelId/client";
import { ListFinetunedModels } from "@/types";

export default class APIFinetuning extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  modelId = new APIModelIdSub(this._client);

  async get({ before, after, limit, order, schemaId, schemaDataId, sortBy }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", schemaId?: string | null, schemaDataId?: string | null, sortBy?: string } = {}): Promise<ListFinetunedModels> {
    let res = await this._fetch({
      url: `/v1/models/finetuning`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "schema_id": schemaId, "schema_data_id": schemaDataId, "sort_by": sortBy },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
