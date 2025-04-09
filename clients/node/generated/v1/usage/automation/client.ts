import { AbstractClient, CompositionClient } from '@/client';
import APIAutomationIdSub from "./automationId/client";

export default class APIAutomation extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  automationId = new APIAutomationIdSub(this._client);

}
