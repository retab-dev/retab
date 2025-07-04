import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIOutlookIdSub from "./outlookId/client";

export default class APIManifest extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  outlookId = new APIOutlookIdSub(this._client);

}
