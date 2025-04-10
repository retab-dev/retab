import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { SchemaCost } from "@/types";

export default class APICost extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get({ startDate, endDate }: { startDate?: Date | null, endDate?: Date | null } = {}): Promise<SchemaCost[]> {
    let res = await this._fetch({
      url: `/v1/analytics/schemas/cost`,
      method: "GET",
      params: { "start_date": startDate, "end_date": endDate },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
