import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZDocumentExtractRequest, DocumentExtractRequest, ZRetabParsedChatCompletionChunk, RetabParsedChatCompletionChunk } from "@/types";

export default class APIStream extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ idempotencyKey, idempotencyForceRefresh, ...body }: { idempotencyKey?: string | null, idempotencyForceRefresh?: boolean } & DocumentExtractRequest): Promise<AsyncGenerator<RetabParsedChatCompletionChunk>> {
    let res = await this._fetch({
      url: `/v1/documents/extractions/stream`,
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey, "Idempotency-ForceRefresh": idempotencyForceRefresh },
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/stream+json") return streamResponse(res, ZRetabParsedChatCompletionChunk);
    throw new Error("Bad content type");
  }
  
}
