import { AbstractClient, CompositionClient } from '@/client';
import APILinkId from "./linkId/client";

export default class APIOpen extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  linkId = new APILinkId(this);

}
