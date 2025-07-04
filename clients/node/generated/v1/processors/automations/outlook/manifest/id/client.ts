import { AbstractClient, CompositionClient, streamResponse } from '@/client';

export default class APIId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(id: string): Promise<any> {
    let res = await this._fetch({
      url: `/v1/processors/automations/outlook/manifest/${id}`,
      method: "GET",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
