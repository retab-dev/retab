import { AbstractClient, CompositionClient } from '@/client';
import { DocumentUploadRequest, AutomationLog } from "@/types";

export default class APIEmail extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(email: string, { ...body }: DocumentUploadRequest): Promise<AutomationLog> {
    return this._fetch({
      url: `/v1/automations/mailboxes/tests/process/${email}`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
