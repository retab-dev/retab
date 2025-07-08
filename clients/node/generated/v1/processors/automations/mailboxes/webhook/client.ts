import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';

export default class APIWebhook extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(): Promise<any> {
    let res = await this._fetch({
      url: `/v1/processors/automations/mailboxes/webhook`,
      method: "POST",
    });
    if (res.headers.get("Content-Type") === "application/json") return z.any().parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
