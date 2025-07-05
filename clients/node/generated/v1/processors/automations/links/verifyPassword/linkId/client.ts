import { AbstractClient, CompositionClient, streamResponse } from '@/client';

export default class APILinkId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(linkId: string, { ...body }: object): Promise<object> {
    let res = await this._fetch({
      url: `/v1/processors/automations/links/verify-password/${linkId}`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
