import { AbstractClient, CompositionClient } from '@/client';
import { DocumentCreateMessageRequest, DocumentMessage } from "@/types";

export default class APICreateMessages extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: DocumentCreateMessageRequest): Promise<DocumentMessage> {
    return this._fetch({
      url: `/v1/documents/create_messages`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
