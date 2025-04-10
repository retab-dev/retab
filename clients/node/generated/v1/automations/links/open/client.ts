import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APILinkIdSub from "./linkId/client";

export default class APIOpen extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  linkId = new APILinkIdSub(this._client);

}
