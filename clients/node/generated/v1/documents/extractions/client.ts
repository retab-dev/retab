import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIStreamSub from "./stream/client";
import { DocumentExtractRequest, UiParsedChatCompletionOutput } from "@/types";

export default class APIExtractions extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  stream = new APIStreamSub(this._client);

  async post({ idempotencyKey, ...body }: { idempotencyKey?: string | null } & DocumentExtractRequest): Promise<UiParsedChatCompletionOutput> {
    let res = await this._fetch({
      url: `/v1/documents/extractions`,
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey },
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
