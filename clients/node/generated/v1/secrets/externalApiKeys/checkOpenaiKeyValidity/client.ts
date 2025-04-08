import { AbstractClient, CompositionClient } from '@/client';
import { ExternalAPIKeyRequest, OpenAIKeyValidationResponse } from "@/types";

export default class APICheckOpenaiKeyValidity extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: ExternalAPIKeyRequest): Promise<OpenAIKeyValidationResponse> {
    return this._fetch({
      url: `/v1/secrets/external_api_keys/check_openai_key_validity`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
