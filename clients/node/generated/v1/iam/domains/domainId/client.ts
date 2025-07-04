import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIDefaultSub from "./default/client";

export default class APIDomainId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  default = new APIDefaultSub(this._client);

  async delete(domainId: string): Promise<any> {
    let res = await this._fetch({
      url: `/v1/iam/domains/${domainId}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
