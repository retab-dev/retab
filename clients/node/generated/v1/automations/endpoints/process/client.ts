import { AbstractClient, CompositionClient } from '@/client';
import APIEndpointId from "./endpointId/client";

export default class APIProcess extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  endpointId = new APIEndpointId(this);

}
