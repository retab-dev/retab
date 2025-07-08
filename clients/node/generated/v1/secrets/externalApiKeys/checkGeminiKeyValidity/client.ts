import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZExternalAPIKeyRequest, ExternalAPIKeyRequest, ZKeyValidationResponse, KeyValidationResponse } from "@/types";

export default class APICheckGeminiKeyValidity extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: ExternalAPIKeyRequest): Promise<KeyValidationResponse> {
    let res = await this._fetch({
      url: `/v1/secrets/external_api_keys/check_gemini_key_validity`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZKeyValidationResponse.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
