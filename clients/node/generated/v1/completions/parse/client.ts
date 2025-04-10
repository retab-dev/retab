import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { UiChatCompletionsParseRequest, UiParsedChatCompletionOutput } from "@/types";

export default class APIParse extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ idempotencyKey, ...body }: { idempotencyKey?: string | null } & UiChatCompletionsParseRequest): Promise<UiParsedChatCompletionOutput> {
    let res = await this._fetch({
      url: `/v1/completions/parse`,
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
