import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIExtractComparisonSub from "./extractComparison/client";

export default class APIBenchmarking extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  extractComparison = new APIExtractComparisonSub(this._client);

}
