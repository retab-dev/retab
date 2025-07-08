import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZVectorSearchRequest, VectorSearchRequest, ZVectorSearchResponse, VectorSearchResponse } from "@/types";

export default class APIVectorSearch extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: VectorSearchRequest): Promise<VectorSearchResponse> {
    let res = await this._fetch({
      url: `/v1/processors/automations/outlook/vector_search/vector_search`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZVectorSearchResponse.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
