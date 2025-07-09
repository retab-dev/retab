import { SyncAPIResource, AsyncAPIResource } from '../../resource.js';
import { ExternalAPIKeys, AsyncExternalAPIKeys } from './external_api_keys.js';
import { Webhooks, AsyncWebhooks } from './webhooks.js';

export class Secrets extends SyncAPIResource {
  public external_api_keys: ExternalAPIKeys;
  public webhooks: Webhooks;

  constructor(client: any) {
    super(client);
    this.external_api_keys = new ExternalAPIKeys(client);
    this.webhooks = new Webhooks(client);
  }
}

export class AsyncSecrets extends AsyncAPIResource {
  public external_api_keys: AsyncExternalAPIKeys;
  public webhooks: AsyncWebhooks;

  constructor(client: any) {
    super(client);
    this.external_api_keys = new AsyncExternalAPIKeys(client);
    this.webhooks = new AsyncWebhooks(client);
  }
}