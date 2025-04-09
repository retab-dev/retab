import { AbstractClient, CompositionClient } from '@/client';
import { DocumentUploadRequest, EmailDataOutput } from "@/types";

export default class APIEmail extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(email: string, { ...body }: DocumentUploadRequest): Promise<EmailDataOutput> {
    return this._fetch({
      url: `/v1/automations/mailboxes/tests/forward/${email}`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
