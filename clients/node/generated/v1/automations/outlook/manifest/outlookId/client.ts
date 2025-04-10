import { AbstractClient, CompositionClient, streamResponse } from '@/client';

export default class APIOutlookId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(outlookId: string): Promise<any> {
    let res = await this._fetch({
      url: `/v1/automations/outlook/manifest/${outlookId}`,
      method: "GET",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
