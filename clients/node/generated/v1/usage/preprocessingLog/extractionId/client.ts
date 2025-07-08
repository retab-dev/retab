import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZPreprocessingLogResponse, PreprocessingLogResponse } from "@/types";

export default class APIExtractionId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(extractionId: string): Promise<PreprocessingLogResponse> {
    let res = await this._fetch({
      url: `/v1/usage/preprocessing_log/${extractionId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZPreprocessingLogResponse.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
