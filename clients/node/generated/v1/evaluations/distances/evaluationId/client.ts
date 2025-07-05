import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIIterationIdxSub from "./iterationIdx/client";

export default class APIEvaluationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  iterationIdx = new APIIterationIdxSub(this._client);

}
