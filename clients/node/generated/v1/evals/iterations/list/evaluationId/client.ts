import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZMainServerServicesV1EvalsIterationsRoutesListIterationsResponse, MainServerServicesV1EvalsIterationsRoutesListIterationsResponse } from "@/types";

export default class APIEvaluationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(evaluationId: string, { model }: { model?: string | null } = {}): Promise<MainServerServicesV1EvalsIterationsRoutesListIterationsResponse> {
    let res = await this._fetch({
      url: `/v1/evals/iterations/list/${evaluationId}`,
      method: "GET",
      params: { "model": model },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZMainServerServicesV1EvalsIterationsRoutesListIterationsResponse.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
