import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZIterationDocumentStatusResponse, IterationDocumentStatusResponse } from "@/types";

export default class APIStatus extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(evaluationId: string, iterationId: string): Promise<IterationDocumentStatusResponse> {
    let res = await this._fetch({
      url: `/v1/evaluations/${evaluationId}/iterations/${iterationId}/status`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZIterationDocumentStatusResponse.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
