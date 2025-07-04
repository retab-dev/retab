import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIIterationMetricsFromEvaluationSub from "./iterationMetricsFromEvaluation/client";
import APIEvaluationIdSub from "./evaluationId/client";

export default class APIDistances extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  iterationMetricsFromEvaluation = new APIIterationMetricsFromEvaluationSub(this._client);
  evaluationId = new APIEvaluationIdSub(this._client);

}
