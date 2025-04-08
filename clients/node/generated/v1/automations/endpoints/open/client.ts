import { AbstractClient, CompositionClient } from '@/client';
import APIEndpointId from "./endpointId/client";

export default class APIOpen extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  endpointId = new APIEndpointId(this);

}
