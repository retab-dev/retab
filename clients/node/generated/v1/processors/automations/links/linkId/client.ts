import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { LinkOutput, UpdateLinkRequest, LinkOutput } from "@/types";

export default class APILinkId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(linkId: string): Promise<LinkOutput> {
    let res = await this._fetch({
      url: `/v1/processors/automations/links/${linkId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async delete(linkId: string): Promise<object> {
    let res = await this._fetch({
      url: `/v1/processors/automations/links/${linkId}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async put(linkId: string, { ...body }: UpdateLinkRequest): Promise<LinkOutput> {
    let res = await this._fetch({
      url: `/v1/processors/automations/links/${linkId}`,
      method: "PUT",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
