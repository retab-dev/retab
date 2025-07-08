import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZRetabTypesEvalsCreateIterationRequest, RetabTypesEvalsCreateIterationRequest, ZRetabTypesEvalsIterationOutput, RetabTypesEvalsIterationOutput } from "@/types";

export default class APIEvaluationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(evaluationId: string, { ...body }: RetabTypesEvalsCreateIterationRequest): Promise<RetabTypesEvalsIterationOutput> {
    let res = await this._fetch({
      url: `/v1/evals/iterations/create/${evaluationId}`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZRetabTypesEvalsIterationOutput.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
