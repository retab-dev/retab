import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { BodyCreateFilesV1DbFilesBatchPost, MultipleUploadResponse } from "@/types";

export default class APIBatch extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: BodyCreateFilesV1DbFilesBatchPost): Promise<MultipleUploadResponse> {
    let res = await this._fetch({
      url: `/v1/db/files/batch`,
      method: "POST",
      body: body,
      bodyMime: "multipart/form-data",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
