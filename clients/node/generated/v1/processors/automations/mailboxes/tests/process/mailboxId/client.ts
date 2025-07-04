import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { DocumentUploadRequest, AutomationLog } from "@/types";

export default class APIMailboxId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(mailboxId: string, { ...body }: DocumentUploadRequest): Promise<AutomationLog> {
    let res = await this._fetch({
      url: `/v1/processors/automations/mailboxes/tests/process/${mailboxId}`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
