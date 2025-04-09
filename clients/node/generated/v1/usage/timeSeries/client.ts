import { AbstractClient, CompositionClient } from '@/client';
import APISchemaIdSub from "./schemaId/client";
import APIAutomationIdSub from "./automationId/client";

export default class APITimeSeries extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  schemaId = new APISchemaIdSub(this._client);
  automationId = new APIAutomationIdSub(this._client);

}
