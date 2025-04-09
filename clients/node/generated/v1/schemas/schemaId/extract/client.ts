import { AbstractClient, CompositionClient } from '@/client';
import { SchemaExtractionRequest, AutomationLog } from "@/types";

export default class APIExtract extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(schemaId: string, { ...body }: SchemaExtractionRequest): Promise<AutomationLog> {
    return this._fetch({
      url: `/v1/schemas/${schemaId}/extract`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
