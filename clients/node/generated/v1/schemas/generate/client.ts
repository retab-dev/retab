import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZGenerateSchemaRequest, GenerateSchemaRequest, ZPartialSchema, PartialSchema } from "@/types";

export default class APIGenerate extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ idempotencyKey, ...body }: { idempotencyKey?: string | null } & GenerateSchemaRequest): Promise<PartialSchema> {
    let res = await this._fetch({
      url: `/v1/schemas/generate`,
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey },
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZPartialSchema.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
