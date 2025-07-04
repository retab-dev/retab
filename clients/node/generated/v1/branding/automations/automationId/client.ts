import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { Branding } from "@/types";

export default class APIAutomationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(automationId: string): Promise<Branding> {
    let res = await this._fetch({
      url: `/v1/branding/automations/${automationId}`,
      method: "GET",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
