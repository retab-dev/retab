import { AbstractClient, CompositionClient } from '@/client';
import APILog from "./log/client";
import APITimeSeries from "./timeSeries/client";
import APISchemaDataId from "./schemaDataId/client";
import APIAutomation from "./automation/client";

export default class APIUsage extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  log = new APILog(this);
  timeSeries = new APITimeSeries(this);
  schemaDataId = new APISchemaDataId(this);
  automation = new APIAutomation(this);

}
