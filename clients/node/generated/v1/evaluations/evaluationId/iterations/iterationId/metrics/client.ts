import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { OptimizedIterationMetrics } from "@/types";

export default class APIMetrics extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(evaluationId: string, iterationId: string): Promise<OptimizedIterationMetrics> {
    let res = await this._fetch({
      url: `/v1/evaluations/${evaluationId}/iterations/${iterationId}/metrics`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
