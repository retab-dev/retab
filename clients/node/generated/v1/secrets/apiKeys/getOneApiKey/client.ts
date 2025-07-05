import { AbstractClient, CompositionClient, streamResponse } from '@/client';

export default class APIGetOneApiKey extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(): Promise<object> {
    let res = await this._fetch({
      url: `/v1/secrets/api_keys/get_one_api_key`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
