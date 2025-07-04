import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APICostSub from "./cost/client";

export default class APISchemas extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  cost = new APICostSub(this._client);

}
