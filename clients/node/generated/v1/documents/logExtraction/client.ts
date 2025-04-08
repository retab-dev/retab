import { AbstractClient, CompositionClient } from '@/client';
import { LogExtractionRequest, LogExtractionResponse } from "@/types";

export default class APILogExtraction extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ IdempotencyKey, ...body }: { IdempotencyKey?: string | null } & LogExtractionRequest): Promise<LogExtractionResponse> {
    return this._fetch({
      url: `/v1/documents/log_extraction`,
      method: "POST",
      params: {  },
      headers: { "Idempotency-Key": IdempotencyKey },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
