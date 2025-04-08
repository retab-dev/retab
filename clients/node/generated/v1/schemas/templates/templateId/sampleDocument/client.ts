import { AbstractClient, CompositionClient } from '@/client';

export default class APISampleDocument extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(templateId: string): Promise<any> {
    return this._fetch({
      url: `/v1/schemas/templates/${templateId}/sample-document`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
