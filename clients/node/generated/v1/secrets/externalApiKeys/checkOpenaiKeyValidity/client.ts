import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { ExternalAPIKeyRequest, KeyValidationResponse } from "@/types";

export default class APICheckOpenaiKeyValidity extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: ExternalAPIKeyRequest): Promise<KeyValidationResponse> {
    let res = await this._fetch({
      url: `/v1/secrets/external_api_keys/check_openai_key_validity`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
