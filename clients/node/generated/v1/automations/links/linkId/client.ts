import { AbstractClient, CompositionClient } from '@/client';
import { LinkOutput } from "@/types";

export default class APILinkId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(linkId: string): Promise<LinkOutput> {
    return this._fetch({
      url: `/v1/automations/links/${linkId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
  async delete(linkId: string): Promise<object> {
    return this._fetch({
      url: `/v1/automations/links/${linkId}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
