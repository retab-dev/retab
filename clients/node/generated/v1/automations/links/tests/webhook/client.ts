import { AbstractClient, CompositionClient } from '@/client';
import APILinkIdSub from "./linkId/client";

export default class APIWebhook extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  linkId = new APILinkIdSub(this._client);

}
