import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';

export default class APIMailboxId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(mailboxId: string): Promise<object> {
    let res = await this._fetch({
      url: `/v1/processors/automations/outlook/tests/email-forwarding-default/${mailboxId}`,
      method: "POST",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return z.object({}).parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
