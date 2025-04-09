import { AbstractClient, CompositionClient } from '@/client';
import APIFieldMetricsSub from "./fieldMetrics/client";
import APIAggregateSub from "./aggregate/client";
import { ListEvals } from "@/types";

export default class APIEvals extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  fieldMetrics = new APIFieldMetricsSub(this._client);
  aggregate = new APIAggregateSub(this._client);

  async get({ before, after, limit, order, fileId, evalId, schemaId, sortBy }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", fileId?: string | null, evalId?: string | null, schemaId?: string | null, sortBy?: string } = {}): Promise<ListEvals> {
    return this._fetch({
      url: `/v1/db/evals`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "file_id": fileId, "eval_id": evalId, "schema_id": schemaId, "sort_by": sortBy },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
