import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { Amount } from "@/types";

export default class APISchemaDataId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(schemaDataId: string, { startDate, endDate }: { startDate?: Date | null, endDate?: Date | null } = {}): Promise<Amount> {
    let res = await this._fetch({
      url: `/v1/usage/schema_data_id/${schemaDataId}`,
      method: "GET",
      params: { "start_date": startDate, "end_date": endDate },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
