import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZMailboxOutput, MailboxOutput, ZUpdateMailboxRequest, UpdateMailboxRequest } from "@/types";

export default class APIMailboxId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(mailboxId: string): Promise<MailboxOutput> {
    let res = await this._fetch({
      url: `/v1/processors/automations/mailboxes/${mailboxId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZMailboxOutput.parse(await res.json());
    throw new Error("Bad content type");
  }
  
  async put(mailboxId: string, { ...body }: UpdateMailboxRequest): Promise<MailboxOutput> {
    let res = await this._fetch({
      url: `/v1/processors/automations/mailboxes/${mailboxId}`,
      method: "PUT",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZMailboxOutput.parse(await res.json());
    throw new Error("Bad content type");
  }
  
  async delete(mailboxId: string): Promise<object> {
    let res = await this._fetch({
      url: `/v1/processors/automations/mailboxes/${mailboxId}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return z.object({}).parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
