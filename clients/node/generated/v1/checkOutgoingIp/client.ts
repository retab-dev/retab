import { AbstractClient, CompositionClient } from '@/client';

export default class APICheckOutgoingIp extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(): Promise<object> {
    return this._fetch({
      url: `/v1/check-outgoing-ip`,
      method: "GET",
    });
  }
  
}
