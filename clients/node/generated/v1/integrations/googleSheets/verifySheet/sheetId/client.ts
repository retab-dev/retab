import { AbstractClient, CompositionClient, streamResponse } from '@/client';

export default class APISheetId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(sheetId: string): Promise<object> {
    let res = await this._fetch({
      url: `/v1/integrations/google_sheets/verify-sheet/${sheetId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
