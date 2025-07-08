import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIIterationIdxSub from "./iterationIdx/client";

export default class APIEvaluationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  iterationIdx = new APIIterationIdxSub(this._client);

}
