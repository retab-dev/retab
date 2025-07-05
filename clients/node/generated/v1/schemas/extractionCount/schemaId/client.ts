import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { ExtractionCount } from "@/types";

export default class APISchemaId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(schemaId: string): Promise<ExtractionCount> {
    let res = await this._fetch({
      url: `/v1/schemas/extraction_count/${schemaId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
