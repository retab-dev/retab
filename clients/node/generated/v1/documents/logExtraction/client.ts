import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZLogExtractionRequest, LogExtractionRequest, ZLogExtractionResponse, LogExtractionResponse } from "@/types";

export default class APILogExtraction extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ idempotencyKey, ...body }: { idempotencyKey?: string | null } & LogExtractionRequest): Promise<LogExtractionResponse> {
    let res = await this._fetch({
      url: `/v1/documents/log_extraction`,
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey },
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZLogExtractionResponse.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
