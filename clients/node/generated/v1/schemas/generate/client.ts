import { AbstractClient, CompositionClient } from '@/client';
import { GenerateSchemaRequest, PartialSchema } from "@/types";

export default class APIGenerate extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ IdempotencyKey, ...body }: { IdempotencyKey?: string | null } & GenerateSchemaRequest): Promise<PartialSchema> {
    return this._fetch({
      url: `/v1/schemas/generate`,
      method: "POST",
      params: {  },
      headers: { "Idempotency-Key": IdempotencyKey },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
