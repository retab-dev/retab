import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { LogCompletionRequest, AutomationLog } from "@/types";

export default class APILog extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: LogCompletionRequest): Promise<AutomationLog> {
    let res = await this._fetch({
      url: `/v1/usage/log`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
