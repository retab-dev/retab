import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { ReviewExtractionRequest, Extraction } from "@/types";

export default class APIAutomationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async put(automationId: string, { extractionId, validationState, ...body }: { extractionId: string, validationState: "pending" | "validated" | "invalid" } & ReviewExtractionRequest): Promise<Extraction> {
    let res = await this._fetch({
      url: `/v1/automations/review-extraction/${automationId}`,
      method: "PUT",
      params: { "extraction_id": extractionId, "validation_state": validationState },
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
