import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZOpenAITierResponse, OpenAITierResponse } from "@/types";

export default class APIOpenaiTier extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(): Promise<OpenAITierResponse> {
    let res = await this._fetch({
      url: `/v1/models/openai_tier`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZOpenAITierResponse.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
