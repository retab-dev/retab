import { AbstractClient, CompositionClient, streamResponse } from '@/client';

export default class APIDownload extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get({ order, sourceDotId, schemaId, schemaDataId, status, validationState, fromDate, toDate, model }: { order?: "asc" | "desc", sourceDotId?: string | null, schemaId?: string | null, schemaDataId?: string | null, status?: string | null, validationState?: string | null, fromDate?: string | null, toDate?: string | null, model?: string | null } = {}): Promise<object> {
    let res = await this._fetch({
      url: `/v1/extractions/download`,
      method: "GET",
      params: { "order": order, "source_dot_id": sourceDotId, "schema_id": schemaId, "schema_data_id": schemaDataId, "status": status, "validation_state": validationState, "from_date": fromDate, "to_date": toDate, "model": model },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
