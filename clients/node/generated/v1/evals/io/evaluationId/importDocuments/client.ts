import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZBodyImportDocumentsV1EvalsIoEvaluationIdImportDocumentsPost, BodyImportDocumentsV1EvalsIoEvaluationIdImportDocumentsPost } from "@/types";

export default class APIImportDocuments extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(evaluationId: string, { ...body }: BodyImportDocumentsV1EvalsIoEvaluationIdImportDocumentsPost): Promise<object> {
    let res = await this._fetch({
      url: `/v1/evals/io/${evaluationId}/import_documents`,
      method: "POST",
      body: body,
      bodyMime: "multipart/form-data",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return z.object({}).parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
