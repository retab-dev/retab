import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIIterationIdSub from "./iterationId/client";

export default class APIGetPayloadForExport extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  iterationId = new APIIterationIdSub(this._client);

}
