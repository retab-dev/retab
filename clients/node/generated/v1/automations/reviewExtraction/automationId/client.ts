import { AbstractClient, CompositionClient } from '@/client';
import { ReviewExtractionRequest, Extraction } from "@/types";

export default class APIAutomationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async put(automationId: string, { extractionId, validationState, ...body }: { extractionId: string, validationState: "pending" | "validated" | "invalid" } & ReviewExtractionRequest): Promise<Extraction> {
    return this._fetch({
      url: `/v1/automations/review-extraction/${automationId}`,
      method: "PUT",
      params: { "extraction_id": extractionId, "validation_state": validationState },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
