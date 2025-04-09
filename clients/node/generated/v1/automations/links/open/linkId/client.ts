import { AbstractClient, CompositionClient } from '@/client';
import { LinkOutput } from "@/types";

export default class APILinkId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(linkId: string): Promise<LinkOutput> {
    return this._fetch({
      url: `/v1/automations/links/open/${linkId}`,
      method: "GET",
    });
  }
  
}
