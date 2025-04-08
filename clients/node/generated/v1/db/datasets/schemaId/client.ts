import { AbstractClient, CompositionClient } from '@/client';
import APIClustering from "./clustering/client";

export default class APISchemaId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  clustering = new APIClustering(this);

}
