import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { LinkOutput } from "@/types";

export default class APILinkId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(linkId: string): Promise<LinkOutput> {
    let res = await this._fetch({
      url: `/v1/automations/links/open/${linkId}`,
      method: "GET",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
