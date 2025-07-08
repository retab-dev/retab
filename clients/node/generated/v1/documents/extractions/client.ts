import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIStreamSub from "./stream/client";
import { ZDocumentExtractRequest, DocumentExtractRequest, ZRetabParsedChatCompletionOutput, RetabParsedChatCompletionOutput } from "@/types";

export default class APIExtractions extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  stream = new APIStreamSub(this._client);

  async post({ idempotencyKey, idempotencyForceRefresh, ...body }: { idempotencyKey?: string | null, idempotencyForceRefresh?: boolean } & DocumentExtractRequest): Promise<RetabParsedChatCompletionOutput> {
    let res = await this._fetch({
      url: `/v1/documents/extractions`,
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey, "Idempotency-ForceRefresh": idempotencyForceRefresh },
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZRetabParsedChatCompletionOutput.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
