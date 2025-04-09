import { AbstractClient, CompositionClient } from '@/client';
import { PromptifyRequest } from "@/types";

export default class APISystemPrompt extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: PromptifyRequest): Promise<any> {
    return this._fetch({
      url: `/v1/schemas/system_prompt`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
