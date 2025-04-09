import { AbstractClient, CompositionClient } from '@/client';
import { DocumentTransformRequest, DocumentTransformResponse } from "@/types";

export default class APICorrectImageOrientation extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: DocumentTransformRequest): Promise<DocumentTransformResponse> {
    return this._fetch({
      url: `/v1/documents/correct_image_orientation`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
