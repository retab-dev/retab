import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIDownloadSub from "./download/client";
import APIExtractionIdSub from "./extractionId/client";
import APIFieldsSub from "./fields/client";
import { PaginatedList } from "@/types";

export default class APIExtractions extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  download = new APIDownloadSub(this._client);
  extractionId = new APIExtractionIdSub(this._client);
  fields = new APIFieldsSub(this._client);

  async get({ before, after, limit, order, sourceDotId, completionDotId, schemaId, schemaDataId, status, validationState, fromDate, toDate, model }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", sourceDotId?: string | null, completionDotId?: string | null, schemaId?: string | null, schemaDataId?: string | null, status?: string | null, validationState?: string | null, fromDate?: string | null, toDate?: string | null, model?: string | null } = {}): Promise<PaginatedList> {
    let res = await this._fetch({
      url: `/v1/extractions`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "source_dot_id": sourceDotId, "completion_dot_id": completionDotId, "schema_id": schemaId, "schema_data_id": schemaDataId, "status": status, "validation_state": validationState, "from_date": fromDate, "to_date": toDate, "model": model },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
