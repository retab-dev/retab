import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIMetricsSub from "./metrics/client";

export default class APIIterationIdx extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  metrics = new APIMetricsSub(this._client);

}
