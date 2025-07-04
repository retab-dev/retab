import { AbstractClient, CompositionClient, streamResponse } from '@/client';

export default class APIApplicationName extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(applicationName: string): Promise<object> {
    let res = await this._fetch({
      url: `/v1/integrations/check_oauth_token/${applicationName}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
