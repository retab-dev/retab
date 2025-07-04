import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { DocumentTransformRequest, DocumentTransformResponse } from "@/types";

export default class APICorrectImageOrientation extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: DocumentTransformRequest): Promise<DocumentTransformResponse> {
    let res = await this._fetch({
      url: `/v1/documents/correct_image_orientation`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
