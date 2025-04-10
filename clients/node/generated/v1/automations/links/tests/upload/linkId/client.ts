import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { DocumentUploadRequest, AutomationLog } from "@/types";

export default class APILinkId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(linkId: string, { ...body }: DocumentUploadRequest): Promise<AutomationLog> {
    let res = await this._fetch({
      url: `/v1/automations/links/tests/upload/${linkId}`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
