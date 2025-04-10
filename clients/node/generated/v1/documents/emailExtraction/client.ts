import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { EmailExtractRequest, UiParsedChatCompletionOutput } from "@/types";

export default class APIEmailExtraction extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ idempotencyKey, ...body }: { idempotencyKey?: string | null } & EmailExtractRequest): Promise<UiParsedChatCompletionOutput> {
    let res = await this._fetch({
      url: `/v1/documents/email_extraction`,
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
