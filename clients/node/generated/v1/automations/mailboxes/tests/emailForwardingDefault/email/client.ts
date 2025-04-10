import { AbstractClient, CompositionClient, streamResponse } from '@/client';

export default class APIEmail extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(email: string): Promise<object> {
    let res = await this._fetch({
      url: `/v1/automations/mailboxes/tests/email-forwarding-default/${email}`,
      method: "POST",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
