import { AbstractClient, CompositionClient } from '@/client';

export default class APIEmail extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(email: string): Promise<object> {
    return this._fetch({
      url: `/v1/automations/outlook/tests/email-forwarding-default/${email}`,
      method: "POST",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
