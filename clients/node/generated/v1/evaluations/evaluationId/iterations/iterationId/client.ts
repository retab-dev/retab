import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIDocumentsSub from "./documents/client";
import APIStatusSub from "./status/client";
import APIProcessSub from "./process/client";
import APIMetricsSub from "./metrics/client";
import { RetabTypesEvaluationsIterationsIterationOutput, PatchIterationRequest, RetabTypesEvaluationsIterationsIterationOutput } from "@/types";

export default class APIIterationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  documents = new APIDocumentsSub(this._client);
  status = new APIStatusSub(this._client);
  process = new APIProcessSub(this._client);
  metrics = new APIMetricsSub(this._client);

  async get(evaluationId: string, iterationId: string): Promise<RetabTypesEvaluationsIterationsIterationOutput> {
    let res = await this._fetch({
      url: `/v1/evaluations/${evaluationId}/iterations/${iterationId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async patch(evaluationId: string, iterationId: string, { ...body }: PatchIterationRequest): Promise<RetabTypesEvaluationsIterationsIterationOutput> {
    let res = await this._fetch({
      url: `/v1/evaluations/${evaluationId}/iterations/${iterationId}`,
      method: "PATCH",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async delete(evaluationId: string, iterationId: string): Promise<object> {
    let res = await this._fetch({
      url: `/v1/evaluations/${evaluationId}/iterations/${iterationId}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
