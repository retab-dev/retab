import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZGenerateSystemPromptRequest, GenerateSystemPromptRequest } from "@/types";

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
    if (res.headers.get("Content-Type") === "application/json") return z.any().parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
