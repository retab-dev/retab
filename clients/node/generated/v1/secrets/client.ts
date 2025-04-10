import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIExternalApiKeysSub from "./externalApiKeys/client";
import APIApiKeysSub from "./apiKeys/client";

export default class APISecrets extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  externalApiKeys = new APIExternalApiKeysSub(this._client);
  apiKeys = new APIApiKeysSub(this._client);

}
