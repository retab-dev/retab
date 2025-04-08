import { AbstractClient, CompositionClient } from '@/client';
import { DocumentCreateInputRequest, DocumentMessage } from "@/types";

export default class APICreateInputs extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: DocumentCreateInputRequest): Promise<DocumentMessage> {
    return this._fetch({
      url: `/v1/documents/create_inputs`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
