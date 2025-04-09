import { AbstractClient, CompositionClient } from '@/client';
import APIClusteringSub from "./clustering/client";

export default class APISchemaId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  clustering = new APIClusteringSub(this._client);

}
