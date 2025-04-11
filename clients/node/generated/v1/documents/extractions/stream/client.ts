import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { DocumentExtractRequest, UiParsedChatCompletionChunk } from "@/types";

export default class APIStream extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ idempotencyKey, ...body }: { idempotencyKey?: string | null } & DocumentExtractRequest): Promise<AsyncGenerator<UiParsedChatCompletionChunk>> {
    let res = await this._fetch({
      url: `/v1/documents/extractions/stream`,
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey },
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/stream+json") return streamResponse(res);
    throw new Error("Bad content type");
  }
  
}
