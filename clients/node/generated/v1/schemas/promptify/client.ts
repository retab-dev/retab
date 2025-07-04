import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { PromptifyRequest, PartialSchema } from "@/types";

export default class APIPromptify extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ idempotencyKey, ...body }: { idempotencyKey?: string | null } & PromptifyRequest): Promise<PartialSchema> {
    let res = await this._fetch({
      url: `/v1/schemas/promptify`,
      method: "POST",
      params: { "idempotency_key": idempotencyKey },
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
