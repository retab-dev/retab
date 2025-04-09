import { AbstractClient, CompositionClient } from '@/client';
import APIOutlookPluginIdSub from "./outlookPluginId/client";

export default class APIOpen extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  outlookPluginId = new APIOutlookPluginIdSub(this._client);

}
