import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';

export default class APIListSpreadsheets extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get({ pageSize }: { pageSize?: number } = {}): Promise<object> {
    let res = await this._fetch({
      url: `/v1/integrations/google_sheets/list-spreadsheets/`,
      method: "GET",
      params: { "pageSize": pageSize },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return z.object({}).parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
