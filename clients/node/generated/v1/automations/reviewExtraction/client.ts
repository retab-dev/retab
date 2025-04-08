import { AbstractClient, CompositionClient } from '@/client';
import APIAutomationId from "./automationId/client";

export default class APIReviewExtraction extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  automationId = new APIAutomationId(this);

}
