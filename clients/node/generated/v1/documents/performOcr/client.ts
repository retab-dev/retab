import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZPerformOCRRequest, PerformOCRRequest, ZPerformOCRResponse, PerformOCRResponse } from "@/types";

export default class APIPerformOcr extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: PerformOCRRequest): Promise<PerformOCRResponse> {
    let res = await this._fetch({
      url: `/v1/documents/perform_ocr`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZPerformOCRResponse.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
