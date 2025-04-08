import { AbstractClient, CompositionClient } from '@/client';
import { DocumentExtractRequest, UiParsedChatCompletionOutput } from "@/types";

export default class APIExtractions extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ IdempotencyKey, ...body }: { IdempotencyKey?: string | null } & DocumentExtractRequest): Promise<UiParsedChatCompletionOutput> {
    return this._fetch({
      url: `/v1/documents/extractions`,
      method: "POST",
      params: {  },
      headers: { "Idempotency-Key": IdempotencyKey },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
