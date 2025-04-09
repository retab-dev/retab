import { AbstractClient, CompositionClient } from '@/client';

export default class APIApplicationName extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(applicationName: string): Promise<object> {
    return this._fetch({
      url: `/v1/integrations/check_oauth_token/${applicationName}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
