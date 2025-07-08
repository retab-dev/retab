import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIIterationIdSub from "./iterationId/client";
import { ZMainServerServicesV1EvaluationsIterationsRoutesListIterationsResponse, MainServerServicesV1EvaluationsIterationsRoutesListIterationsResponse, ZRetabTypesEvaluationsIterationsCreateIterationRequest, RetabTypesEvaluationsIterationsCreateIterationRequest, ZRetabTypesEvaluationsIterationsIterationOutput, RetabTypesEvaluationsIterationsIterationOutput } from "@/types";

export default class APIIterations extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  iterationId = new APIIterationIdSub(this._client);

  async get(evaluationId: string, { model }: { model?: string | null } = {}): Promise<MainServerServicesV1EvaluationsIterationsRoutesListIterationsResponse> {
    let res = await this._fetch({
      url: `/v1/evaluations/${evaluationId}/iterations`,
      method: "GET",
      params: { "model": model },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZMainServerServicesV1EvaluationsIterationsRoutesListIterationsResponse.parse(await res.json());
    throw new Error("Bad content type");
  }
  
  async post(evaluationId: string, { ...body }: RetabTypesEvaluationsIterationsCreateIterationRequest): Promise<RetabTypesEvaluationsIterationsIterationOutput> {
    let res = await this._fetch({
      url: `/v1/evaluations/${evaluationId}/iterations`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZRetabTypesEvaluationsIterationsIterationOutput.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
