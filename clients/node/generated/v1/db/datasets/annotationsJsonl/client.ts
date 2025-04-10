import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { ClusteringExtractionRequest, ClusteringExtractionResponse } from "@/types";

export default class APIAnnotationsJsonl extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: ClusteringExtractionRequest): Promise<ClusteringExtractionResponse> {
    let res = await this._fetch({
      url: `/v1/db/datasets/annotations-jsonl`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
