import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIStreamSub from "./stream/client";
import { ZLLMAnnotateDocumentRequest, LLMAnnotateDocumentRequest, ZRetabParsedChatCompletionOutput, RetabParsedChatCompletionOutput } from "@/types";

export default class APILlmAnnotate extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  stream = new APIStreamSub(this._client);

  async post(evaluationId: string, documentId: string, { ...body }: LLMAnnotateDocumentRequest): Promise<RetabParsedChatCompletionOutput> {
    let res = await this._fetch({
      url: `/v1/evaluations/${evaluationId}/documents/${documentId}/llm-annotate`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZRetabParsedChatCompletionOutput.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
