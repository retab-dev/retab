import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { DocumentTransformRequest } from "@/types";

export default class APIConvertToPdf extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: DocumentTransformRequest): Promise<any> {
    let res = await this._fetch({
      url: `/v1/documents/convert_to_pdf`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
