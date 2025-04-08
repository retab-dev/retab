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
      params: {  },
      headers: {  },
    });
  }
  
  async put(email: string, { ...body }: UpdateMailboxRequest): Promise<MailboxOutput> {
    return this._fetch({
      url: `/v1/automations/mailboxes/${email}`,
      method: "PUT",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
  async delete(email: string): Promise<object> {
    return this._fetch({
      url: `/v1/automations/mailboxes/${email}`,
      method: "DELETE",
      params: {  },
      headers: {  },
    });
  }
  
}
