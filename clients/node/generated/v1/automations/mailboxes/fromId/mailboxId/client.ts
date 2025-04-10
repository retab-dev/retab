import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { MailboxOutput } from "@/types";

export default class APIMailboxId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(mailboxId: string): Promise<MailboxOutput> {
    let res = await this._fetch({
      url: `/v1/automations/mailboxes/from_id/${mailboxId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
