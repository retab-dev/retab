import { AbstractClient, CompositionClient } from '@/client';

export default class APICreateMessagesMock extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(): Promise<object> {
    return this._fetch({
      url: `/v1/documents/create_messagesMock`,
      method: "POST",
    });
  }
  
}
