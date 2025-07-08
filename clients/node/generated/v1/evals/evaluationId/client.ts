import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIDuplicateSub from "./duplicate/client";
import { ZRetabTypesEvalsEvaluationOutput, RetabTypesEvalsEvaluationOutput, ZRetabTypesEvalsEvaluationInput, RetabTypesEvalsEvaluationInput, ZMainServerServicesV1EvalsRoutesPatchEvaluationRequest, MainServerServicesV1EvalsRoutesPatchEvaluationRequest } from "@/types";

export default class APIEvaluationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  duplicate = new APIDuplicateSub(this._client);

  async get(evaluationId: string): Promise<RetabTypesEvalsEvaluationOutput> {
    let res = await this._fetch({
      url: `/v1/evals/${evaluationId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZRetabTypesEvalsEvaluationOutput.parse(await res.json());
    throw new Error("Bad content type");
  }
  
  async put(evaluationId: string, { newFilesIds, ...body }: { newFilesIds?: string[] | null } & RetabTypesEvalsEvaluationInput): Promise<RetabTypesEvalsEvaluationOutput> {
    let res = await this._fetch({
      url: `/v1/evals/${evaluationId}`,
      method: "PUT",
      params: { "new_files_ids": newFilesIds },
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZRetabTypesEvalsEvaluationOutput.parse(await res.json());
    throw new Error("Bad content type");
  }
  
  async patch(evaluationId: string, { ...body }: MainServerServicesV1EvalsRoutesPatchEvaluationRequest): Promise<RetabTypesEvalsEvaluationOutput> {
    let res = await this._fetch({
      url: `/v1/evals/${evaluationId}`,
      method: "PATCH",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZRetabTypesEvalsEvaluationOutput.parse(await res.json());
    throw new Error("Bad content type");
  }
  
  async delete(evaluationId: string): Promise<object> {
    let res = await this._fetch({
      url: `/v1/evals/${evaluationId}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return z.object({}).parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
