import { AbstractClient, CompositionClient } from '@/client';
import { EmailExtractRequest, UiParsedChatCompletionOutput } from "@/types";

export default class APIEmailExtraction extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ IdempotencyKey, ...body }: { IdempotencyKey?: string | null } & EmailExtractRequest): Promise<UiParsedChatCompletionOutput> {
    return this._fetch({
      url: `/v1/documents/email_extraction`,
      method: "POST",
      params: {  },
      headers: { "Idempotency-Key": IdempotencyKey },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
