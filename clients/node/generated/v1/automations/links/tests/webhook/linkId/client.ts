import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { AutomationLog } from "@/types";

export default class APILinkId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(linkId: string): Promise<AutomationLog> {
    let res = await this._fetch({
      url: `/v1/automations/links/tests/webhook/${linkId}`,
      method: "POST",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
