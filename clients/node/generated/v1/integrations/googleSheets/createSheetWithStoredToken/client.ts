import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { CreateSheetWithStoredTokenRequest } from "@/types";

export default class APICreateSheetWithStoredToken extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: CreateSheetWithStoredTokenRequest): Promise<object> {
    let res = await this._fetch({
      url: `/v1/integrations/google_sheets/create-sheet-with-stored-token/`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
