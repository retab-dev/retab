import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { RetabChatResponseCreateRequest, RetabParsedChatCompletionOutput } from "@/types";

export default class APIResponses extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ idempotencyKey, openAIApiKey, anthropicApiKey, geminiApiKey, xAIApiKey, ...body }: { idempotencyKey?: string | null, openAIApiKey?: string | null, anthropicApiKey?: string | null, geminiApiKey?: string | null, xAIApiKey?: string | null } & RetabChatResponseCreateRequest): Promise<RetabParsedChatCompletionOutput> {
    let res = await this._fetch({
      url: `/v1/consensus/responses`,
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey, "OpenAI-Api-Key": openAIApiKey, "Anthropic-Api-Key": anthropicApiKey, "Gemini-Api-Key": geminiApiKey, "XAI-Api-Key": xAIApiKey },
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
