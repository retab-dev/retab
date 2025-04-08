import { AbstractClient, CompositionClient } from '@/client';

export default class APISheetId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(sheetId: string): Promise<object> {
    return this._fetch({
      url: `/v1/integrations/google_sheets/verify-sheet/${sheetId}`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
