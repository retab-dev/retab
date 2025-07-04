import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIStreamSub from "./stream/client";
import { BodySubmitToProcessorV1ProcessorsProcessorIdSubmitPost, RetabParsedChatCompletionOutput } from "@/types";

export default class APISubmit extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  stream = new APIStreamSub(this._client);

  async post(processorId: string, { idempotencyKey, idempotencyForceRefresh, openAIApiKey, anthropicApiKey, geminiApiKey, xAIApiKey, ...body }: { idempotencyKey?: string | null, idempotencyForceRefresh?: boolean, openAIApiKey?: string | null, anthropicApiKey?: string | null, geminiApiKey?: string | null, xAIApiKey?: string | null } & BodySubmitToProcessorV1ProcessorsProcessorIdSubmitPost): Promise<RetabParsedChatCompletionOutput> {
    let res = await this._fetch({
      url: `/v1/processors/${processorId}/submit`,
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey, "Idempotency-ForceRefresh": idempotencyForceRefresh, "OpenAI-Api-Key": openAIApiKey, "Anthropic-Api-Key": anthropicApiKey, "Gemini-Api-Key": geminiApiKey, "XAI-Api-Key": xAIApiKey },
      body: body,
      bodyMime: "multipart/form-data",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
