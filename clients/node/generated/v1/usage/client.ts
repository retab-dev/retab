import { AbstractClient, CompositionClient } from '@/client';
import APILogSub from "./log/client";
import APITimeSeriesSub from "./timeSeries/client";
import APISchemaDataIdSub from "./schemaDataId/client";
import APIAutomationSub from "./automation/client";

export default class APIUsage extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  log = new APILogSub(this._client);
  timeSeries = new APITimeSeriesSub(this._client);
  schemaDataId = new APISchemaDataIdSub(this._client);
  automation = new APIAutomationSub(this._client);

}
