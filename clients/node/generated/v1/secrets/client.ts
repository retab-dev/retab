import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIApiKeysSub from "./apiKeys/client";
import APIExternalApiKeysSub from "./externalApiKeys/client";

export default class APISecrets extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  apiKeys = new APIApiKeysSub(this._client);
  externalApiKeys = new APIExternalApiKeysSub(this._client);

}
