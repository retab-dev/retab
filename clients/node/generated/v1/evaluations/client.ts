import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIDistancesSub from "./distances/client";
import APIIoSub from "./io/client";
import APIProjectsSub from "./projects/client";
import APIEvaluationIdSub from "./evaluationId/client";
import { ZCreateEvaluation, CreateEvaluation, ZRetabTypesEvaluationsModelEvaluationOutput, RetabTypesEvaluationsModelEvaluationOutput, ZMainServerServicesV1EvaluationsRoutesListEvaluations, MainServerServicesV1EvaluationsRoutesListEvaluations } from "@/types";

export default class APIEvaluations extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  distances = new APIDistancesSub(this._client);
  io = new APIIoSub(this._client);
  projects = new APIProjectsSub(this._client);
  evaluationId = new APIEvaluationIdSub(this._client);

  async post({ ...body }: CreateEvaluation): Promise<RetabTypesEvaluationsModelEvaluationOutput> {
    let res = await this._fetch({
      url: `/v1/evaluations`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZRetabTypesEvaluationsModelEvaluationOutput.parse(await res.json());
    throw new Error("Bad content type");
  }
  
  async get({ before, after, limit, order, projectId, schemaId, schemaDataId }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", projectId?: string | null, schemaId?: string | null, schemaDataId?: string | null } = {}): Promise<MainServerServicesV1EvaluationsRoutesListEvaluations> {
    let res = await this._fetch({
      url: `/v1/evaluations`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "project_id": projectId, "schema_id": schemaId, "schema_data_id": schemaDataId },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZMainServerServicesV1EvaluationsRoutesListEvaluations.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
