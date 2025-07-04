import { AbstractClient, CompositionClient, streamResponse } from '@/client';

export default class APIMemberId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async delete(memberId: string): Promise<object> {
    let res = await this._fetch({
      url: `/v1/iam/team/members/${memberId}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
