import { AbstractClient, CompositionClient } from '@/client';

export default class APIGetOneApiKey extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(): Promise<object> {
    return this._fetch({
      url: `/v1/secrets/api_keys/get_one_api_key`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
