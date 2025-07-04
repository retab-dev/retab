import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIFieldPathSub from "./fieldPath/client";

export default class APIFieldMetrics extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  fieldPath = new APIFieldPathSub(this._client);

}
