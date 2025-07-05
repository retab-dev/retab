import { AbstractClient, CompositionClient, streamResponse } from '@/client';

export default class APISpreadsheetId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(spreadsheetId: string): Promise<object> {
    let res = await this._fetch({
      url: `/v1/integrations/google_sheets/list-worksheets/${spreadsheetId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
