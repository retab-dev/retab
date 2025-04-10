import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { ExternalAPIKey } from "@/types";

export default class APIProvider extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(provider: "OpenAI" | "Gemini"): Promise<ExternalAPIKey> {
    let res = await this._fetch({
      url: `/v1/secrets/external_api_keys/${provider}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async delete(provider: "OpenAI" | "Gemini"): Promise<object> {
    let res = await this._fetch({
      url: `/v1/secrets/external_api_keys/${provider}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
