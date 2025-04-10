import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { Amount } from "@/types";

export default class APISchemaId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(schemaId: string, { startDate, endDate }: { startDate?: Date | null, endDate?: Date | null } = {}): Promise<Amount> {
    let res = await this._fetch({
      url: `/v1/usageschema_id/${schemaId}`,
      method: "GET",
      params: { "start_date": startDate, "end_date": endDate },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
