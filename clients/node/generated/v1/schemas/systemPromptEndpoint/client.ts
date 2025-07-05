import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { GenerateSystemPromptRequest } from "@/types";

export default class APISystemPromptEndpoint extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: GenerateSystemPromptRequest): Promise<any> {
    let res = await this._fetch({
      url: `/v1/schemas/system_prompt_endpoint`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
