import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { BodySubmitToProcessorV1ProcessorsProcessorIdSubmitStreamPost } from "@/types";

export default class APIStream extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(processorId: string, { idempotencyKey, idempotencyForceRefresh, openAIApiKey, anthropicApiKey, geminiApiKey, xAIApiKey, ...body }: { idempotencyKey?: string | null, idempotencyForceRefresh?: boolean, openAIApiKey?: string | null, anthropicApiKey?: string | null, geminiApiKey?: string | null, xAIApiKey?: string | null } & BodySubmitToProcessorV1ProcessorsProcessorIdSubmitStreamPost): Promise<AsyncGenerator<string>> {
    let res = await this._fetch({
      url: `/v1/processors/${processorId}/submit/stream`,
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey, "Idempotency-ForceRefresh": idempotencyForceRefresh, "OpenAI-Api-Key": openAIApiKey, "Anthropic-Api-Key": anthropicApiKey, "Gemini-Api-Key": geminiApiKey, "XAI-Api-Key": xAIApiKey },
      body: body,
      bodyMime: "multipart/form-data",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/stream+json") return streamResponse(res);
    throw new Error("Bad content type");
  }
  
}
