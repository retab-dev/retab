import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZLLMAnnotateDocumentRequest, LLMAnnotateDocumentRequest } from "@/types";

export default class APIStream extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(evaluationId: string, documentId: string, { ...body }: LLMAnnotateDocumentRequest): Promise<AsyncGenerator<string>> {
    let res = await this._fetch({
      url: `/v1/evaluations/${evaluationId}/documents/${documentId}/llm-annotate/stream`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/stream+json") return streamResponse(res, z.string());
    throw new Error("Bad content type");
  }
  
}
