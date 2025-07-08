import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APILlmAnnotateSub from "./llmAnnotate/client";
import { ZPatchEvaluationDocumentRequest, PatchEvaluationDocumentRequest, ZEvaluationDocumentOutput, EvaluationDocumentOutput } from "@/types";

export default class APIDocumentId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  llmAnnotate = new APILlmAnnotateSub(this._client);

  async patch(evaluationId: string, documentId: string, { ...body }: PatchEvaluationDocumentRequest): Promise<EvaluationDocumentOutput> {
    let res = await this._fetch({
      url: `/v1/evaluations/${evaluationId}/documents/${documentId}`,
      method: "PATCH",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZEvaluationDocumentOutput.parse(await res.json());
    throw new Error("Bad content type");
  }
  
  async delete(evaluationId: string, documentId: string): Promise<object> {
    let res = await this._fetch({
      url: `/v1/evaluations/${evaluationId}/documents/${documentId}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return z.object({}).parse(await res.json());
    throw new Error("Bad content type");
  }
  
  async get(evaluationId: string, documentId: string): Promise<EvaluationDocumentOutput> {
    let res = await this._fetch({
      url: `/v1/evaluations/${evaluationId}/documents/${documentId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZEvaluationDocumentOutput.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
