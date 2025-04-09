import { AbstractClient, CompositionClient } from '@/client';

export default class APIKeyId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async delete(keyId: string): Promise<object> {
    return this._fetch({
      url: `/v1/secrets/api_keys/${keyId}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
