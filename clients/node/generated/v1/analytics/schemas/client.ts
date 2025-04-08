import { AbstractClient, CompositionClient } from '@/client';
import APICost from "./cost/client";

export default class APISchemas extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  cost = new APICost(this);

}
