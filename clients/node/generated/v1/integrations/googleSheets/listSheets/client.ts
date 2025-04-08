import { AbstractClient, CompositionClient } from '@/client';

export default class APIListSheets extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(): Promise<object> {
    return this._fetch({
      url: `/v1/integrations/google_sheets/list-sheets/`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
