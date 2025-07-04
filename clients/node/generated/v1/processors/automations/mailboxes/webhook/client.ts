import { AbstractClient, CompositionClient, streamResponse } from '@/client';

export default class APIWebhook extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(): Promise<any> {
    let res = await this._fetch({
      url: `/v1/processors/automations/mailboxes/webhook`,
      method: "POST",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
