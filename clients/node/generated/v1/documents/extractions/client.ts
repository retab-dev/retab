import { AbstractClient, CompositionClient } from '@/client';
import { DocumentExtractRequest, UiParsedChatCompletionOutput } from "@/types";

export default class APIExtractions extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ idempotencyKey, ...body }: { idempotencyKey?: string | null } & DocumentExtractRequest): Promise<UiParsedChatCompletionOutput> {
    return this._fetch({
      url: `/v1/documents/extractions`,
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey },
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
