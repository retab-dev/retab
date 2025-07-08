import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZDistancesResult, DistancesResult } from "@/types";

export default class APIDocumentId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(iterationId: string, documentId: string, { metricType }: { metricType?: "levenshtein" | "jaccard" | "hamming" } = {}): Promise<DistancesResult> {
    let res = await this._fetch({
      url: `/v1/evals/iterations/${iterationId}/distances/${documentId}`,
      method: "GET",
      params: { "metric_type": metricType },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZDistancesResult.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
