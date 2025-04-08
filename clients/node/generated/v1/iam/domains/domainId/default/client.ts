import { AbstractClient, CompositionClient } from '@/client';

export default class APIDefault extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async put(domainId: string): Promise<any> {
    return this._fetch({
      url: `/v1/iam/domains/${domainId}/default`,
      method: "PUT",
      params: {  },
      headers: {  },
    });
  }
  
}
