import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APISampleDocumentSub from "./sampleDocument/client";
import { Extraction } from "@/types";

export default class APIExtractionId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  sampleDocument = new APISampleDocumentSub(this._client);

  async get(extractionId: string): Promise<Extraction> {
    let res = await this._fetch({
      url: `/v1/extractions/${extractionId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
