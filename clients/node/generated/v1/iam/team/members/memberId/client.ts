import { AbstractClient, CompositionClient } from '@/client';

export default class APIMemberId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async delete(memberId: string): Promise<object> {
    return this._fetch({
      url: `/v1/iam/team/members/${memberId}`,
      method: "DELETE",
      params: {  },
      headers: {  },
    });
  }
  
}
