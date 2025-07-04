import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { AutomationLog } from "@/types";

export default class APIEmail extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(email: string): Promise<AutomationLog> {
    let res = await this._fetch({
      url: `/v1/automations/mailboxes/tests/webhook/${email}`,
      method: "POST",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
