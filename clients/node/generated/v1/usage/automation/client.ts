import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIAutomationIdSub from "./automationId/client";

export default class APIAutomation extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  automationId = new APIAutomationIdSub(this._client);

}
