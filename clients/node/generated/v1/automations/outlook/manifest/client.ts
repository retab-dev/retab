import { AbstractClient, CompositionClient } from '@/client';
import APIOutlookId from "./outlookId/client";

export default class APIManifest extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  outlookId = new APIOutlookId(this);

}
