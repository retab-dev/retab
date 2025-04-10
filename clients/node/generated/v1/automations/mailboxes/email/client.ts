import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { MailboxOutput, UpdateMailboxRequest, MailboxOutput } from "@/types";

export default class APIEmail extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(email: string): Promise<MailboxOutput> {
    let res = await this._fetch({
      url: `/v1/automations/mailboxes/${email}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async put(email: string, { ...body }: UpdateMailboxRequest): Promise<MailboxOutput> {
    let res = await this._fetch({
      url: `/v1/automations/mailboxes/${email}`,
      method: "PUT",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async delete(email: string): Promise<object> {
    let res = await this._fetch({
      url: `/v1/automations/mailboxes/${email}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
