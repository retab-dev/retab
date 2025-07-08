import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZExtractionCount, ExtractionCount } from "@/types";

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
    if (res.headers.get("Content-Type") === "application/json") return ZExtractionCount.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
