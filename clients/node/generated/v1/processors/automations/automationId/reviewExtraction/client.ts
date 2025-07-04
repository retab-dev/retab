import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { ReviewExtractionRequest, Extraction } from "@/types";

export default class APIReviewExtraction extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async put(automationId: string, { extractionId, validationState, ...body }: { extractionId: string, validationState: "pending" | "validated" | "invalid" } & ReviewExtractionRequest): Promise<Extraction> {
    let res = await this._fetch({
      url: `/v1/processors/automations/${automationId}/review-extraction`,
      method: "PUT",
      params: { "extraction_id": extractionId, "validation_state": validationState },
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
