import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZOptimizedIterationMetrics, OptimizedIterationMetrics } from "@/types";

export default class APIMetrics extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(evaluationId: string, iterationIdx: number): Promise<OptimizedIterationMetrics> {
    let res = await this._fetch({
      url: `/v1/evaluations/distances/${evaluationId}/${iterationIdx}/metrics`,
      method: "GET",
    });
    if (res.headers.get("Content-Type") === "application/json") return ZOptimizedIterationMetrics.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
