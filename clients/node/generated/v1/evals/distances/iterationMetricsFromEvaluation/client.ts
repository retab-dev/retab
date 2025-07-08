import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZMainServerServicesV1EvalsDistancesRoutesIterationMetricsFromEvaluationRequest, MainServerServicesV1EvalsDistancesRoutesIterationMetricsFromEvaluationRequest, ZOptimizedIterationMetrics, OptimizedIterationMetrics } from "@/types";

export default class APIIterationMetricsFromEvaluation extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: MainServerServicesV1EvalsDistancesRoutesIterationMetricsFromEvaluationRequest): Promise<OptimizedIterationMetrics> {
    let res = await this._fetch({
      url: `/v1/evals/distances/iteration_metrics_from_evaluation`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
    });
    if (res.headers.get("Content-Type") === "application/json") return ZOptimizedIterationMetrics.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
