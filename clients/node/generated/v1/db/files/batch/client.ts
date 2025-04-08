import { AbstractClient, CompositionClient } from '@/client';
import { BodyCreateFilesV1DbFilesBatchPost, MultipleUploadResponse } from "@/types";

export default class APIBatch extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: BodyCreateFilesV1DbFilesBatchPost): Promise<MultipleUploadResponse> {
    return this._fetch({
      url: `/v1/db/files/batch`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "multipart/form-data",
    });
  }
  
}
