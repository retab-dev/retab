import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { ProcessIterationRequest, RetabTypesEvaluationsIterationsIterationOutput } from "@/types";

export default class APIProcess extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(evaluationId: string, iterationId: string, { ...body }: ProcessIterationRequest): Promise<RetabTypesEvaluationsIterationsIterationOutput> {
    let res = await this._fetch({
      url: `/v1/evaluations/${evaluationId}/iterations/${iterationId}/process`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
