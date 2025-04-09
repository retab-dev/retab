import { AbstractClient, CompositionClient } from '@/client';
import { VectorSearchRequest, VectorSearchResponse } from "@/types";

export default class APIVectorSearch extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: VectorSearchRequest): Promise<VectorSearchResponse> {
    return this._fetch({
      url: `/v1/automations/outlook/vector_search/vector_search`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
