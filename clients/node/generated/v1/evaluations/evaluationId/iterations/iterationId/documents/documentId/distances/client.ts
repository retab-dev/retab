import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { ComputeDictSimilarityResponse } from "@/types";

export default class APIDistances extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(evaluationId: string, iterationId: string, documentId: string, { metricType }: { metricType?: "levenshtein" | "jaccard" | "hamming" } = {}): Promise<ComputeDictSimilarityResponse> {
    let res = await this._fetch({
      url: `/v1/evaluations/${evaluationId}/iterations/${iterationId}/documents/${documentId}/distances`,
      method: "GET",
      params: { "metric_type": metricType },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
