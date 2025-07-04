import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { UpdateEmailDataRequest, EmailDataOutput } from "@/types";

export default class APIUpdateEmailData extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: UpdateEmailDataRequest): Promise<EmailDataOutput> {
    let res = await this._fetch({
      url: `/v1/processors/automations/outlook/update_email_data`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
