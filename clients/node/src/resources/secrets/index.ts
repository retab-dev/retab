import { SyncAPIResource, AsyncAPIResource } from '../../resource.js';
import { ExternalAPIKeys, AsyncExternalAPIKeys } from './external_api_keys.js';

export class Secrets extends SyncAPIResource {
  public external_api_keys: ExternalAPIKeys;

  constructor(client: any) {
    super(client);
    this.external_api_keys = new ExternalAPIKeys(client);
  }
}

export class AsyncSecrets extends AsyncAPIResource {
  public external_api_keys: AsyncExternalAPIKeys;

  constructor(client: any) {
    super(client);
    this.external_api_keys = new AsyncExternalAPIKeys(client);
  }
}