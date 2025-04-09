import { AbstractClient, CompositionClient } from '@/client';
import { GenerateSchemaRequest, PartialSchema } from "@/types";

export default class APIGenerate extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ idempotencyKey, ...body }: { idempotencyKey?: string | null } & GenerateSchemaRequest): Promise<PartialSchema> {
    return this._fetch({
      url: `/v1/schemas/generate`,
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey },
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
