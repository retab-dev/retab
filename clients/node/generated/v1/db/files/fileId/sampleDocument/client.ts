import { AbstractClient, CompositionClient } from '@/client';

export default class APISampleDocument extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(fileId: string): Promise<any> {
    return this._fetch({
      url: `/v1/db/files/${fileId}/sample-document`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
