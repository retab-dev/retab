import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZSpreadsheetDetails, SpreadsheetDetails } from "@/types";

export default class APISpreadsheetId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(spreadsheetId: string): Promise<SpreadsheetDetails> {
    let res = await this._fetch({
      url: `/v1/integrations/google_sheets/get-spreadsheet-details/${spreadsheetId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZSpreadsheetDetails.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
