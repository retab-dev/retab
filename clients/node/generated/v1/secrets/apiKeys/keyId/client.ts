import { AbstractClient, CompositionClient, streamResponse } from '@/client';

export default class APIKeyId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async delete(keyId: string): Promise<object> {
    let res = await this._fetch({
      url: `/v1/secrets/api_keys/${keyId}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
