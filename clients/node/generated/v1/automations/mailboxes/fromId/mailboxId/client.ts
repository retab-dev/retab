import { AbstractClient, CompositionClient } from '@/client';
import { MailboxOutput } from "@/types";

export default class APIMailboxId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(mailboxId: string): Promise<MailboxOutput> {
    return this._fetch({
      url: `/v1/automations/mailboxes/from_id/${mailboxId}`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
