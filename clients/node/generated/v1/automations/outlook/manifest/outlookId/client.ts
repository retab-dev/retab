import { AbstractClient, CompositionClient } from '@/client';

export default class APIOutlookId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(outlookId: string): Promise<any> {
    return this._fetch({
      url: `/v1/automations/outlook/manifest/${outlookId}`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
