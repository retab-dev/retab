import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { CreateSpreadsheetWithStoredTokenRequest, SpreadsheetDetails } from "@/types";

export default class APICreateSpreadsheetWithStoredToken extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: CreateSpreadsheetWithStoredTokenRequest): Promise<SpreadsheetDetails> {
    let res = await this._fetch({
      url: `/v1/integrations/google_sheets/create-spreadsheet-with-stored-token/`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
