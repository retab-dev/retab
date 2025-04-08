import { AbstractClient, CompositionClient } from '@/client';

export default class APIInvitationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async delete(invitationId: string): Promise<object> {
    return this._fetch({
      url: `/v1/iam/team/invitations/${invitationId}`,
      method: "DELETE",
      params: {  },
      headers: {  },
    });
  }
  
}
