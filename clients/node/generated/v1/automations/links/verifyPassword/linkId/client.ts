import { AbstractClient, CompositionClient } from '@/client';

export default class APILinkId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(linkId: string, { ...body }: object): Promise<object> {
    return this._fetch({
      url: `/v1/automations/links/verify-password/${linkId}`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
