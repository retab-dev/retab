import { AbstractClient, CompositionClient, streamResponse } from '@/client';

export default class APICheckOutgoingIp extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(): Promise<object> {
    let res = await this._fetch({
      url: `/v1/check-outgoing-ip`,
      method: "GET",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
