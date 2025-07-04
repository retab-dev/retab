import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { DocumentUploadRequest, AutomationLog } from "@/types";

export default class APIEmail extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(email: string, { ...body }: DocumentUploadRequest): Promise<AutomationLog> {
    let res = await this._fetch({
      url: `/v1/automations/outlook/tests/process/${email}`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
