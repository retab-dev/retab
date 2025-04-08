import { AbstractClient, CompositionClient } from '@/client';

export default class APIReception extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(): Promise<object> {
    return this._fetch({
      url: `/v1/automations/links/tests/reception`,
      method: "POST",
      params: {  },
      headers: {  },
    });
  }
  
}
