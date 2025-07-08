import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIDocumentIdSub from "./documentId/client";
import { ZListEvaluationDocumentsResponse, ListEvaluationDocumentsResponse, ZDocumentItem, DocumentItem, ZEvaluationDocumentOutput, EvaluationDocumentOutput } from "@/types";

export default class APIDocuments extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  documentId = new APIDocumentIdSub(this._client);

  async get(evaluationId: string, { filename }: { filename?: string | null } = {}): Promise<ListEvaluationDocumentsResponse> {
    let res = await this._fetch({
      url: `/v1/evaluations/${evaluationId}/documents`,
      method: "GET",
      params: { "filename": filename },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZListEvaluationDocumentsResponse.parse(await res.json());
    throw new Error("Bad content type");
  }
  
  async post(evaluationId: string, { ...body }: DocumentItem): Promise<EvaluationDocumentOutput> {
    let res = await this._fetch({
      url: `/v1/evaluations/${evaluationId}/documents`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZEvaluationDocumentOutput.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
