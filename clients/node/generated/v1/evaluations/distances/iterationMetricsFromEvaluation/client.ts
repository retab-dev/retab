import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { MainServerServicesV1EvaluationsDistancesRoutesIterationMetricsFromEvaluationRequest, OptimizedIterationMetrics } from "@/types";

export default class APIIterationMetricsFromEvaluation extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: MainServerServicesV1EvaluationsDistancesRoutesIterationMetricsFromEvaluationRequest): Promise<OptimizedIterationMetrics> {
    let res = await this._fetch({
      url: `/v1/evaluations/distances/iteration_metrics_from_evaluation`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
