import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APISchemaIdSub from "./schemaId/client";
import APIAutomationIdSub from "./automationId/client";
import APIProcessorIdSub from "./processorId/client";

export default class APITimeSeries extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  schemaId = new APISchemaIdSub(this._client);
  automationId = new APIAutomationIdSub(this._client);
  processorId = new APIProcessorIdSub(this._client);

}
