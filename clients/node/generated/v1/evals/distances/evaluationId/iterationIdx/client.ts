import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIMetricsSub from "./metrics/client";

export default class APIIterationIdx extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  metrics = new APIMetricsSub(this._client);

}
