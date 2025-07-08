import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIDistancesSub from "./distances/client";
import APIIoSub from "./io/client";
import APIIterationsSub from "./iterations/client";
import APIProjectsSub from "./projects/client";
import APIEvaluationIdSub from "./evaluationId/client";
import { ZMainServerServicesV1EvalsRoutesListEvaluations, MainServerServicesV1EvalsRoutesListEvaluations, ZRetabTypesEvalsEvaluationInput, RetabTypesEvalsEvaluationInput, ZRetabTypesEvalsEvaluationOutput, RetabTypesEvalsEvaluationOutput } from "@/types";

export default class APIEvals extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  distances = new APIDistancesSub(this._client);
  io = new APIIoSub(this._client);
  iterations = new APIIterationsSub(this._client);
  projects = new APIProjectsSub(this._client);
  evaluationId = new APIEvaluationIdSub(this._client);

  async get({ before, after, limit, order, projectId, evaluationId, name, schemaId, schemaDataId }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", projectId?: string | null, evaluationId?: string | null, name?: string | null, schemaId?: string | null, schemaDataId?: string | null } = {}): Promise<MainServerServicesV1EvalsRoutesListEvaluations> {
    let res = await this._fetch({
      url: `/v1/evals`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "project_id": projectId, "evaluation_id": evaluationId, "name": name, "schema_id": schemaId, "schema_data_id": schemaDataId },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZMainServerServicesV1EvalsRoutesListEvaluations.parse(await res.json());
    throw new Error("Bad content type");
  }
  
  async post({ ...body }: RetabTypesEvalsEvaluationInput): Promise<RetabTypesEvalsEvaluationOutput> {
    let res = await this._fetch({
      url: `/v1/evals`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZRetabTypesEvalsEvaluationOutput.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
