import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZDuplicateEvaluationRequest, DuplicateEvaluationRequest, ZRetabTypesEvaluationsModelEvaluationOutput, RetabTypesEvaluationsModelEvaluationOutput } from "@/types";

export default class APIDuplicate extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(evaluationId: string, { ...body }: DuplicateEvaluationRequest): Promise<RetabTypesEvaluationsModelEvaluationOutput> {
    let res = await this._fetch({
      url: `/v1/evaluations/${evaluationId}/duplicate`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZRetabTypesEvaluationsModelEvaluationOutput.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
