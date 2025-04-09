import { AbstractClient, CompositionClient } from '@/client';
import { ExtractionCount } from "@/types";

export default class APISchemaId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(schemaId: string): Promise<ExtractionCount> {
    return this._fetch({
      url: `/v1/schemas/extraction_count/${schemaId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
