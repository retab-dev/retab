import { AbstractClient, CompositionClient } from '@/client';
import APISampleDocument from "./sampleDocument/client";
import { Extraction } from "@/types";

export default class APIExtractionId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  sampleDocument = new APISampleDocument(this);

  async get(extractionId: string): Promise<Extraction> {
    return this._fetch({
      url: `/v1/extractions_logs/${extractionId}`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
