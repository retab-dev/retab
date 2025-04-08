import { AbstractClient, CompositionClient } from '@/client';

export default class APIFields extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(): Promise<object> {
    return this._fetch({
      url: `/v1/extractions_logs/fields`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
