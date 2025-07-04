import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { ComputeDictSimilarityRequest, ComputeDictSimilarityResponse } from "@/types";

export default class APISimilarity extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: ComputeDictSimilarityRequest): Promise<ComputeDictSimilarityResponse> {
    let res = await this._fetch({
      url: `/v1/consensus/similarity`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
