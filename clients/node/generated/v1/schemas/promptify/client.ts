import { AbstractClient, CompositionClient } from '@/client';
import { PromptifyRequest, PartialSchema } from "@/types";

export default class APIPromptify extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ idempotencyKey, ...body }: { idempotencyKey?: string | null } & PromptifyRequest): Promise<PartialSchema> {
    return this._fetch({
      url: `/v1/schemas/promptify`,
      method: "POST",
      params: { "idempotency_key": idempotencyKey },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
