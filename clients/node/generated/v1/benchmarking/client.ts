import { AbstractClient, CompositionClient } from '@/client';
import APIExtractComparisonSub from "./extractComparison/client";

export default class APIBenchmarking extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  extractComparison = new APIExtractComparisonSub(this._client);

}
