import { AbstractClient, CompositionClient } from '@/client';
import APIExternalApiKeys from "./externalApiKeys/client";
import APIApiKeys from "./apiKeys/client";

export default class APISecrets extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  externalApiKeys = new APIExternalApiKeys(this);
  apiKeys = new APIApiKeys(this);

}
