import { AbstractClient, CompositionClient, streamResponse } from '@/client';

export default class APICreateMessagesMock extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(): Promise<object> {
    let res = await this._fetch({
      url: `/v1/documents/create_messagesMock`,
      method: "POST",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
