import { AbstractClient, CompositionClient } from '@/client';

export default class APIGetAuthUrl extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(): Promise<object> {
    return this._fetch({
      url: `/v1/integrations/oauth/google/get-auth-url`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
