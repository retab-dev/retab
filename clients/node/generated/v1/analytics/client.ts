import { AbstractClient, CompositionClient } from '@/client';
import APISchemas from "./schemas/client";

export default class APIAnalytics extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  schemas = new APISchemas(this);

}
