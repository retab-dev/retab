import { AbstractClient, CompositionClient, streamResponse } from '@/client';

export default class APIDefault extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async put(domainId: string): Promise<any> {
    let res = await this._fetch({
      url: `/v1/iam/domains/${domainId}/default`,
      method: "PUT",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
