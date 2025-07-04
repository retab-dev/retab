import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { ExternalAPIKey } from "@/types";

export default class APIProvider extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(provider: "OpenAI" | "Anthropic" | "Gemini" | "xAI" | "Retab", { includeFallbackApiKeys }: { includeFallbackApiKeys?: boolean } = {}): Promise<ExternalAPIKey> {
    let res = await this._fetch({
      url: `/v1/secrets/external_api_keys/${provider}`,
      method: "GET",
      params: { "include_fallback_api_keys": includeFallbackApiKeys },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async delete(provider: "OpenAI" | "Anthropic" | "Gemini" | "xAI" | "Retab"): Promise<object> {
    let res = await this._fetch({
      url: `/v1/secrets/external_api_keys/${provider}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
