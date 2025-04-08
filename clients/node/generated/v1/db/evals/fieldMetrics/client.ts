import { AbstractClient, CompositionClient } from '@/client';
import APIFieldPath from "./fieldPath/client";

export default class APIFieldMetrics extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  fieldPath = new APIFieldPath(this);

}
