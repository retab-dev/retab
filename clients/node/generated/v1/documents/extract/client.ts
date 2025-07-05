import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { DocumentExtractRequest, RetabParsedChatCompletionOutput } from "@/types";

export default class APIExtract extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ idempotencyKey, idempotencyForceRefresh, ...body }: { idempotencyKey?: string | null, idempotencyForceRefresh?: boolean } & DocumentExtractRequest): Promise<RetabParsedChatCompletionOutput> {
    let res = await this._fetch({
      url: `/v1/documents/extract`,
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey, "Idempotency-ForceRefresh": idempotencyForceRefresh },
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
