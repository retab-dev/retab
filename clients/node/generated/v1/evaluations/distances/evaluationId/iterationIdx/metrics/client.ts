import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { OptimizedIterationMetrics } from "@/types";

export default class APIMetrics extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(evaluationId: string, iterationIdx: number): Promise<OptimizedIterationMetrics> {
    let res = await this._fetch({
      url: `/v1/evaluations/distances/${evaluationId}/${iterationIdx}/metrics`,
      method: "GET",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
