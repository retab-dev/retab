import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZPerformOCROnlyRequest, PerformOCROnlyRequest, ZPerformOCROnlyResponse, PerformOCROnlyResponse } from "@/types";

export default class APIPerformOcrOnly extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: PerformOCROnlyRequest): Promise<PerformOCROnlyResponse> {
    let res = await this._fetch({
      url: `/v1/documents/perform_ocr_only`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZPerformOCROnlyResponse.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
