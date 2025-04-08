import { AbstractClient, CompositionClient } from '@/client';
import APIModel from "./model/client";
import { PaginatedList } from "@/types";

export default class APIFinetunedModels extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  model = new APIModel(this);

  async get({ before, after, limit, order, schemaId, model, schemaDataId }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", schemaId?: string | null, model?: string | null, schemaDataId?: string | null }): Promise<PaginatedList> {
    return this._fetch({
      url: `/v1/finetuned_models`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "schema_id": schemaId, "model": model, "schema_data_id": schemaDataId },
      headers: {  },
    });
  }
  
}
