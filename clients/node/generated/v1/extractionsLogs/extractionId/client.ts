import { AbstractClient, CompositionClient } from '@/client';
import APISampleDocumentSub from "./sampleDocument/client";
import { Extraction } from "@/types";

export default class APIExtractionId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  sampleDocument = new APISampleDocumentSub(this._client);

  async get(extractionId: string): Promise<Extraction> {
    return this._fetch({
      url: `/v1/extractions_logs/${extractionId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
