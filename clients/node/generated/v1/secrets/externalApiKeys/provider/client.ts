import { AbstractClient, CompositionClient } from '@/client';
import { ExternalAPIKey } from "@/types";

export default class APIProvider extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(provider: "OpenAI" | "Gemini"): Promise<ExternalAPIKey> {
    return this._fetch({
      url: `/v1/secrets/external_api_keys/${provider}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
  async delete(provider: "OpenAI" | "Gemini"): Promise<object> {
    return this._fetch({
      url: `/v1/secrets/external_api_keys/${provider}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
