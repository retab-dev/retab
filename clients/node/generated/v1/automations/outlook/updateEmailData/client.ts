import { AbstractClient, CompositionClient } from '@/client';
import { UpdateEmailDataRequest, EmailDataOutput } from "@/types";

export default class APIUpdateEmailData extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: UpdateEmailDataRequest): Promise<EmailDataOutput> {
    return this._fetch({
      url: `/v1/automations/outlook/update_email_data`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
