import { AbstractClient, CompositionClient } from '@/client';
import { LogExtractionRequest, LogExtractionResponse } from "@/types";

export default class APILogExtraction extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ idempotencyKey, ...body }: { idempotencyKey?: string | null } & LogExtractionRequest): Promise<LogExtractionResponse> {
    return this._fetch({
      url: `/v1/documents/log_extraction`,
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey },
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
