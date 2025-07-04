import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { LinkOutput } from "@/types";

export default class APILinkId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(linkId: string): Promise<LinkOutput> {
    let res = await this._fetch({
      url: `/v1/automations/links/${linkId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async delete(linkId: string): Promise<object> {
    let res = await this._fetch({
      url: `/v1/automations/links/${linkId}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
