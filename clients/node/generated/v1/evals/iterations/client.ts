import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIListSub from "./list/client";
import APICreateSub from "./create/client";
import APIIterationIdSub from "./iterationId/client";

export default class APIIterations extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  list = new APIListSub(this._client);
  create = new APICreateSub(this._client);
  iterationId = new APIIterationIdSub(this._client);

}
