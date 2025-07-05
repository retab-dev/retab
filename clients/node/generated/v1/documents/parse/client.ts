import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { ParseRequest, ParseResult } from "@/types";

export default class APIParse extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: ParseRequest): Promise<ParseResult> {
    let res = await this._fetch({
      url: `/v1/documents/parse`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
