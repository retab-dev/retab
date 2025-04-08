import { AbstractClient, CompositionClient } from '@/client';
import APIExtract from "./extract/client";

export default class APISchemaId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  extract = new APIExtract(this);

}
