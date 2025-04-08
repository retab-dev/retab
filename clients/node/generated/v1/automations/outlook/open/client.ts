import { AbstractClient, CompositionClient } from '@/client';
import APIOutlookPluginId from "./outlookPluginId/client";

export default class APIOpen extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  outlookPluginId = new APIOutlookPluginId(this);

}
