import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIEndpointIdSub from "./endpointId/client";

export default class APIProcess extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  endpointId = new APIEndpointIdSub(this._client);

}
