import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZRetabChatCompletionsRequest, RetabChatCompletionsRequest, ZRetabParsedChatCompletionOutput, RetabParsedChatCompletionOutput } from "@/types";

export default class APICompletions extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ idempotencyKey, openAIApiKey, anthropicApiKey, geminiApiKey, xAIApiKey, ...body }: { idempotencyKey?: string | null, openAIApiKey?: string | null, anthropicApiKey?: string | null, geminiApiKey?: string | null, xAIApiKey?: string | null } & RetabChatCompletionsRequest): Promise<RetabParsedChatCompletionOutput> {
    let res = await this._fetch({
      url: `/v1/consensus/completions`,
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey, "OpenAI-Api-Key": openAIApiKey, "Anthropic-Api-Key": anthropicApiKey, "Gemini-Api-Key": geminiApiKey, "XAI-Api-Key": xAIApiKey },
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZRetabParsedChatCompletionOutput.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
