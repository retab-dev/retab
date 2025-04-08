import { AbstractClient, CompositionClient } from '@/client';
import APIExtractComparison from "./extractComparison/client";

export default class APIBenchmarking extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  extractComparison = new APIExtractComparison(this);

}
