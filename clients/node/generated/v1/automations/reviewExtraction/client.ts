import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIAutomationIdSub from "./automationId/client";

export default class APIReviewExtraction extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  automationId = new APIAutomationIdSub(this._client);

}
