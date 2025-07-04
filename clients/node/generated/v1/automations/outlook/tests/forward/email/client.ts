import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { DocumentUploadRequest, EmailDataOutput } from "@/types";

export default class APIEmail extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(email: string, { ...body }: DocumentUploadRequest): Promise<EmailDataOutput> {
    let res = await this._fetch({
      url: `/v1/automations/outlook/tests/forward/${email}`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
