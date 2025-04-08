import { AbstractClient, CompositionClient } from '@/client';
import { DocumentUploadRequest, AutomationLog } from "@/types";

export default class APILinkId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(linkId: string, { ...body }: DocumentUploadRequest): Promise<AutomationLog> {
    return this._fetch({
      url: `/v1/automations/links/tests/upload/${linkId}`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
