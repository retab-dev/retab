import { AbstractClient, CompositionClient } from '@/client';
import APIAutomationId from "./automationId/client";

export default class APIWebhook extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  automationId = new APIAutomationId(this);

}
