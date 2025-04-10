import { AbstractClient, CompositionClient, streamResponse } from '@/client';

export default class APIInvitationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async delete(invitationId: string): Promise<object> {
    let res = await this._fetch({
      url: `/v1/iam/team/invitations/${invitationId}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
