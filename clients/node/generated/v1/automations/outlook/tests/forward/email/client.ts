import { AbstractClient, CompositionClient } from '@/client';
import { DocumentUploadRequest, EmailDataOutput } from "@/types";

export default class APIEmail extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(email: string, { ...body }: DocumentUploadRequest): Promise<EmailDataOutput> {
    return this._fetch({
      url: `/v1/automations/outlook/tests/forward/${email}`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
