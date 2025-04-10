import { AbstractClient, CompositionClient, streamResponse } from '@/client';

export default class APIGetToken extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(): Promise<object> {
    let res = await this._fetch({
      url: `/v1/integrations/oauth/google/get-token`,
      method: "POST",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
