import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { PromptifyRequest } from "@/types";

export default class APISystemPrompt extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: PromptifyRequest): Promise<any> {
    let res = await this._fetch({
      url: `/v1/schemas/system_prompt`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
