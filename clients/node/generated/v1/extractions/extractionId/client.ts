import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APISampleDocumentSub from "./sampleDocument/client";
import { ZExtraction, Extraction } from "@/types";

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
    if (res.headers.get("Content-Type") === "application/json") return ZExtraction.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
