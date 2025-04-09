import { AbstractClient, CompositionClient } from '@/client';

export default class APIGetToken extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(): Promise<object> {
    return this._fetch({
      url: `/v1/integrations/oauth/google/get-token`,
      method: "POST",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
