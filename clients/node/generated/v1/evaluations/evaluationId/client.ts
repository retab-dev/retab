import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIDocumentsSub from "./documents/client";
import APIIterationsSub from "./iterations/client";
import APIDuplicateSub from "./duplicate/client";
import { RetabTypesEvaluationsModelEvaluationOutput, RetabTypesEvaluationsModelPatchEvaluationRequest } from "@/types";

export default class APIEvaluationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  documents = new APIDocumentsSub(this._client);
  iterations = new APIIterationsSub(this._client);
  duplicate = new APIDuplicateSub(this._client);

  async get(evaluationId: string): Promise<RetabTypesEvaluationsModelEvaluationOutput> {
    let res = await this._fetch({
      url: `/v1/evaluations/${evaluationId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
  async patch(evaluationId: string, { ...body }: RetabTypesEvaluationsModelPatchEvaluationRequest): Promise<RetabTypesEvaluationsModelEvaluationOutput> {
    let res = await this._fetch({
      url: `/v1/evaluations/${evaluationId}`,
      method: "PATCH",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
  async delete(evaluationId: string): Promise<object> {
    let res = await this._fetch({
      url: `/v1/evaluations/${evaluationId}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
