import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { ListFieldMetrics } from "@/types";

export default class APIFieldPath extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(fieldPath: string, { before, after, limit, order, fileId, evalId, schemaId, sortBy }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", fileId?: string | null, evalId?: string | null, schemaId?: string | null, sortBy?: string } = {}): Promise<ListFieldMetrics> {
    let res = await this._fetch({
      url: `/v1/db/evals/field_metrics/${fieldPath}`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "file_id": fileId, "eval_id": evalId, "schema_id": schemaId, "sort_by": sortBy },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
