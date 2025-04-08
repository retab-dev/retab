import { AbstractClient, CompositionClient } from '@/client';
import APISchemaId from "./schemaId/client";
import APIAutomationId from "./automationId/client";

export default class APITimeSeries extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  schemaId = new APISchemaId(this);
  automationId = new APIAutomationId(this);

}
