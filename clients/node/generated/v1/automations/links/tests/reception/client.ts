import { AbstractClient, CompositionClient, streamResponse } from '@/client';

export default class APIReception extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(): Promise<object> {
    let res = await this._fetch({
      url: `/v1/automations/links/tests/reception`,
      method: "POST",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
