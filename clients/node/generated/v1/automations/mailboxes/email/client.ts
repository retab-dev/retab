import { AbstractClient, CompositionClient } from '@/client';
import { MailboxOutput, UpdateMailboxRequest, MailboxOutput } from "@/types";

export default class APIEmail extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(email: string): Promise<MailboxOutput> {
    return this._fetch({
      url: `/v1/automations/mailboxes/${email}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
  async put(email: string, { ...body }: UpdateMailboxRequest): Promise<MailboxOutput> {
    return this._fetch({
      url: `/v1/automations/mailboxes/${email}`,
      method: "PUT",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
  async delete(email: string): Promise<object> {
    return this._fetch({
      url: `/v1/automations/mailboxes/${email}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
