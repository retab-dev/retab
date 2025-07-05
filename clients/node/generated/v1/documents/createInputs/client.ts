import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { DocumentCreateInputRequest, DocumentMessage } from "@/types";

export default class APICreateInputs extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: DocumentCreateInputRequest): Promise<DocumentMessage> {
    let res = await this._fetch({
      url: `/v1/documents/create_inputs`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
