import { AbstractClient, CompositionClient } from '@/client';

export default class APIOrganizationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(): Promise<object> {
    return this._fetch({
      url: `/v1/integrations/google_sheets/organization_id`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
